package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickeskov/netmon/pkg/common"
	"github.com/nickeskov/netmon/pkg/monitor"
	"github.com/nickeskov/netmon/pkg/service"
	"github.com/nickeskov/netmon/pkg/service/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
	config := appConfig{}
	// setup logger for config parsing
	_, s := common.SetupLogger("INFO")
	config.registerAndParseAll(s)
	// setup logger again for further usage
	_, _ = common.SetupLogger(config.logLevel)
	zap.S().Info("starting server...")

	// basic validations
	if config.statsHistorySize < 1 {
		zap.S().Fatal("'stats-history-size' parameter should be greater than zero")
	}
	if config.maxPollResponseSize < 1 {
		zap.S().Fatal("'max-poll-response-size' parameter should be greater than zero")
	}
	initialState, err := monitor.NewNetworkMonitoringStateFromString(config.initialMonState)
	if err != nil {
		zap.S().Fatalf("invalid monitoring initial state %q", initialState.String())
	}
	if config.httpAuthHeader == "" {
		zap.S().Fatal("please, provide non empty 'http-auth-header' parameter")
	}
	if config.httpAuthToken == "" {
		zap.S().Fatal("please, provide 'http-auth-token' parameter")
	}

	criteria := monitor.NetworkErrorCriteria{
		NodesDown: monitor.NodesDownCriterion{
			TotalDownNodesPart: config.criterionNodesDownTotalPart,
		},
		NodesHeight: monitor.NodesHeightCriterion{
			HeightDiff:              config.criterionNodesHeightDiff,
			RequireMinNodesOnHeight: config.criterionNodesHeightRequireMinNodesOnHeight,
		},
		StateHash: monitor.NodesStateHashCriterion{
			MinStateHashGroupsOnSameHeight:   config.criterionNodesStateHashMinStateHashGroupsOnSameHeight,
			MinValuableStateHashGroups:       config.criterionNodesStateHashMinValuableStateHashGroups,
			MinNodesInValuableStateHashGroup: config.criterionNodesStateHashMinNodesInValuableStateHashGroup,
			RequireMinNodesOnHeight:          config.criterionNodesStateHashRequireMinNodesOnHeight,
		},
	}
	if err := criteria.Validate(); err != nil {
		zap.S().Fatalf("invalid criteria: %v", err)
	}

	mon, err := monitor.NewNetworkMonitoring(
		initialState,
		monitor.NetworkSchemeChar(config.networkScheme),
		config.statsHistorySize,
		monitor.NewNodesStatsScraperHTTP(config.nodeStatsURL, int64(config.maxPollResponseSize)),
		config.networkErrorsStreak,
		criteria,
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

		monitoringService := service.NewNetworkMonitoringService(&mon)
		authMiddleWare := middleware.NewHTTPAuthTokenMiddleware(config.httpAuthHeader, config.httpAuthToken)

		// public URLs
		http.HandleFunc("/health", monitoringService.NetworkHealth)
		// private URLs
		http.Handle("/state", authMiddleWare(http.HandlerFunc(monitoringService.SetMonitorState)))

		// run monitor service
		monitorDone := mon.RunInBackground(ctx, config.pollNodesStatsInterval)

		server := http.Server{Addr: config.bindAddr, Handler: nil, ReadHeaderTimeout: time.Second, ReadTimeout: 10 * time.Second}
		server.RegisterOnShutdown(func() {
			// wait for monitor
			<-monitorDone
		})

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

	zap.S().Info("sever successfully started")

	sig := <-gracefulStop
	zap.S().Infof("caught signal %q, stopping...", sig)
	cancel()
	if err := <-httpDone; err != nil {
		zap.S().Fatalf("HTTP server error: %v", err)
	}
	zap.S().Infof("server has been stopped successfully")
}
