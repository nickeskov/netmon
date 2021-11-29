package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	logLevel               = flag.String("log-level", "INFO", "Logging level. Supported levels: DEV, DEBUG, INFO, WARN, ERROR, FATAL. Default logging level INFO.")
	bindAddr               = flag.String("bind-addr", ":8080", "Local network address to bind the HTTP API of the service on. Default value is ':8080'.")
	network                = flag.String("network", "mainnet", "WAVES network type. Supported values: mainnet, testnet, stagenet.")
	nodeStatsURL           = flag.String("stats-url", "https://waves-nodes-get-height.wavesnodes.com/", "Nodes statistics URL.")
	pollNodesStatsInterval = flag.Duration("stats-poll-interval", time.Minute, "Nodes statistics polling interval. Default value 1m.")
	networkErrorsStreak    = flag.Int("network-errors-streak", 5, "Network will be considered as degraded after that errors streak.")

	criterionNodesDownTotalPercentage                  = flag.Float64("criterion-down-total-percentage", 0.3, "")
	criterionNodesDownNodesDownOnSameVersionPercentage = flag.Float64("criterion-down-on-same-version-percentage", 0.5, "")
	criterionNodesDownRequireMinNodesOnSameVersion     = flag.Int("criterion-down-require-min-nodes-on-same-version", 2, "")

	criterionNodesHeightDiff                    = flag.Int("criterion-height-diff", 5, "")
	criterionNodesHeightRequireMinNodesOnHeight = flag.Int("criterion-height-require-min-nodes-on-same-height", 2, "")

	criterionNodesStateHashMinStateHashGroupsOnSameHeight   = flag.Int("criterion-statehash-min-groups-on-same-height", 2, "")
	criterionNodesStateHashMinValuableStateHashGroups       = flag.Int("criterion-statehash-min-valuable-groups", 2, "")
	criterionNodesStateHashMinNodesInValuableStateHashGroup = flag.Int("criterion-statehash-min-nodes-in-valuable-group", 2, "")
	criterionNodesStateHashRequireMinNodesOnHeight          = flag.Int("criterion-statehash-require-min-nodes-on-same-height", 4, "")
)

func init() {
	flag.Parse()
	_, _ = SetupLogger(*logLevel)
}

func main() {
	criteria := NetworkErrorCriteria{
		NodesDown: NodesDownCriterion{
			TotalDownNodesPercentage:         *criterionNodesDownTotalPercentage,
			NodesDownOnSameVersionPercentage: *criterionNodesDownNodesDownOnSameVersionPercentage,
			RequireMinNodesOnSameVersion:     *criterionNodesDownRequireMinNodesOnSameVersion,
		},
		NodesHeight: NodesHeightCriterion{
			HeightDiff:              *criterionNodesHeightDiff,
			RequireMinNodesOnHeight: *criterionNodesHeightRequireMinNodesOnHeight,
		},
		StateHash: NodesStateHashCriterion{
			MinStateHashGroupsOnSameHeight:   *criterionNodesStateHashMinStateHashGroupsOnSameHeight,
			MinValuableStateHashGroups:       *criterionNodesStateHashMinValuableStateHashGroups,
			MinNodesInValuableStateHashGroup: *criterionNodesStateHashMinNodesInValuableStateHashGroup,
			RequireMinNodesOnHeight:          *criterionNodesStateHashRequireMinNodesOnHeight,
		},
	}

	monitor, err := NewNetworkMonitoring(
		*network,
		NewNodesStatsScraperHTTP(*nodeStatsURL),
		*networkErrorsStreak,
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

		service := NewNetworkMonitoringService(&monitor)
		monitorDone := monitor.RunInBackground(ctx, *pollNodesStatsInterval)

		server := http.Server{Addr: *bindAddr, Handler: nil}
		server.RegisterOnShutdown(func() {
			// wait for monitor
			<-monitorDone
		})

		http.HandleFunc("/health", service.NetworkHealth)
		// TODO: protect this by AUTH middleware
		http.HandleFunc("/state", service.SetMonitorState)

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
