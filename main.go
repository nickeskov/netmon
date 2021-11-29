package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO: add flags
// TODO: create auth middleware which checks HMAC http header
// TODO: remove hardcode

func init() {
	al := zap.NewAtomicLevel()
	al.SetLevel(zap.DebugLevel)
	ec := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(ec), zapcore.Lock(os.Stdout), al)
	logger := zap.New(core)
	zap.ReplaceGlobals(logger.WithOptions())
}

func main() {
	monitor, err := NewNetworkMonitoring(
		"mainnet",
		NewNodesStatsScraperHTTP("https://waves-nodes-get-height.wavesnodes.com/"),
		5,
		networkErrorCriteria{
			nodesDown: nodesDownCriterion{
				totalDownNodesPercentage:         0.5,
				nodesDownOnSameVersionPercentage: 0.5,
				requireMinNodesOnSameVersion:     2,
			},
			height: heightCriterion{
				heightDiff:              10,
				requireMinNodesOnHeight: 2,
			},
			stateHash: stateHashCriterion{
				minStateHashGroupsOnSameHeight:   2,
				minValuableStateHashGroups:       2,
				minNodesInValuableStateHashGroup: 2,
				requireMinNodesOnHeight:          4,
			},
		},
	)
	if err != nil {
		zap.S().Fatalf("failed to init monitor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	httpDone := make(chan error, 1)
	go func() {
		var servErr error

		shutdownDone := make(chan error, 1)
		defer func() {
			// waiting for shutdown, combining and throwing "done" message or error up through the httpDone chan
			if shutdownErr := <-shutdownDone; shutdownErr != nil {
				if servErr != nil {
					httpDone <- errors.Wrapf(shutdownErr, "%v", servErr)
				} else {
					httpDone <- shutdownErr
				}
			} else {
				httpDone <- servErr
			}
		}()

		service := NewNetworkMonitoringService(&monitor)
		// TODO: remove hardcoded duration
		monitorDone := monitor.RunInBackground(ctx, time.Minute)

		// TODO: remove hardcoded address
		server := http.Server{Addr: ":8080", Handler: nil}
		server.RegisterOnShutdown(func() {
			// wait for monitor
			<-monitorDone
		})

		http.HandleFunc("/health", service.NetworkHealth)
		// TODO: uncomment when AUTH middleware will be done
		//http.HandleFunc("/state", service.SetMonitorState)

		// run graceful HTTP shutdown worker
		go func() {
			<-ctx.Done()
			var shutdownErr error
			// waiting for all idle connections
			if shutdownErr = server.Shutdown(context.Background()); shutdownErr != nil {
				zap.S().Errorf("HTTP servers shutdown: %v", shutdownErr)
			}
			// send shutdown done message
			shutdownDone <- shutdownErr
		}()

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.S().Errorf("HTTP ListenAndServe: %v", err)
			servErr = err
		}
	}()

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	sig := <-gracefulStop
	zap.S().Infof("caught signal %q, stopping...", sig)
	cancel()
	if err := <-httpDone; err != nil {
		zap.S().Fatalf("HTTP server error: %v", err)
	}
	zap.S().Infof("server has been stopped successfully")
}
