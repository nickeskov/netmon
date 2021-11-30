package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickeskov/netmon/pkg/common"
	"github.com/nickeskov/netmon/pkg/monitor"
	"github.com/nickeskov/netmon/pkg/service"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	logLevel               = flag.String("log-level", "INFO", "Logging level. Supported levels: 'DEV', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'.")
	bindAddr               = flag.String("bind-addr", ":2048", "Local network address to bind the HTTP API of the service on.")
	network                = flag.String("network", "mainnet", "WAVES network type. Supported networks: 'mainnet', 'testnet', 'stagenet'.")
	nodeStatsURL           = flag.String("stats-url", "https://waves-nodes-get-height.wavesnodes.com/", "Nodes statistics URL.")
	pollNodesStatsInterval = flag.Duration("stats-poll-interval", time.Minute, "Nodes statistics polling interval.")
	networkErrorsStreak    = flag.Int("network-errors-streak", 5, "Network will be considered as degraded after that errors streak.")
	initialMonState        = flag.String("initial-mon-state", "active", "Initial monitoring state. Possible states: 'active', 'frozen_operates_stable', 'frozen_degraded'.")

	criterionNodesDownTotalPart = flag.Float64("criterion-down-total-part", 0.3, "Alert will be generated if detected down nodes part greater than that criterion.")

	criterionNodesHeightDiff                    = flag.Int("criterion-height-diff", 5, "Alert will be generated if detected height diff greater than that criterion.")
	criterionNodesHeightRequireMinNodesOnHeight = flag.Int("criterion-height-require-min-nodes-on-same-height", 2, "Minimum required amount of nodes on same height for height-diff criterion.")

	criterionNodesStateHashMinStateHashGroupsOnSameHeight   = flag.Int("criterion-statehash-min-groups-on-same-height", 2, "Alert won't be generated if detected amount of statehash groups on same height lower than that criterion.")
	criterionNodesStateHashMinValuableStateHashGroups       = flag.Int("criterion-statehash-min-valuable-groups", 2, "Alert won't be generated if detected amount of statehash 'valuable' groups on same height lower than that criterion.")
	criterionNodesStateHashMinNodesInValuableStateHashGroup = flag.Int("criterion-statehash-min-nodes-in-valuable-group", 2, "StateHash group will be considered as 'valuable' if contains 'criterion-statehash-min-valuable-groups'.")
	criterionNodesStateHashRequireMinNodesOnHeight          = flag.Int("criterion-statehash-require-min-nodes-on-same-height", 4, "Minimum required amount of nodes on same height for statehash criterion.")
)

func init() {
	flag.Parse()
	_, _ = common.SetupLogger(*logLevel)
}

func main() {
	zap.S().Info("starting server...")

	if n := *network; n != "mainnet" && n != "testnet" && n != "stagenet" {
		zap.S().Fatalf("invalid network %q", n)
	}

	initialState, err := monitor.NewNetworkMonitoringStateFromString(*initialMonState)
	if err != nil {
		zap.S().Fatalf("invalid monitoring initial state %q", initialState)
	}

	criteria := monitor.NetworkErrorCriteria{
		NodesDown: monitor.NodesDownCriterion{
			TotalDownNodesPart: *criterionNodesDownTotalPart,
		},
		NodesHeight: monitor.NodesHeightCriterion{
			HeightDiff:              *criterionNodesHeightDiff,
			RequireMinNodesOnHeight: *criterionNodesHeightRequireMinNodesOnHeight,
		},
		StateHash: monitor.NodesStateHashCriterion{
			MinStateHashGroupsOnSameHeight:   *criterionNodesStateHashMinStateHashGroupsOnSameHeight,
			MinValuableStateHashGroups:       *criterionNodesStateHashMinValuableStateHashGroups,
			MinNodesInValuableStateHashGroup: *criterionNodesStateHashMinNodesInValuableStateHashGroup,
			RequireMinNodesOnHeight:          *criterionNodesStateHashRequireMinNodesOnHeight,
		},
	}
	if err := criteria.Validate(); err != nil {
		zap.S().Fatalf("invalid criteria: %v", err)
	}

	mon, err := monitor.NewNetworkMonitoring(
		initialState,
		*network,
		monitor.NewNodesStatsScraperHTTP(*nodeStatsURL),
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

		monitoringService := service.NewNetworkMonitoringService(&mon)
		monitorDone := mon.RunInBackground(ctx, *pollNodesStatsInterval)

		server := http.Server{Addr: *bindAddr, Handler: nil}
		server.RegisterOnShutdown(func() {
			// wait for monitor
			<-monitorDone
		})

		http.HandleFunc("/health", monitoringService.NetworkHealth)
		// TODO: protect this by AUTH middleware
		http.HandleFunc("/state", monitoringService.SetMonitorState)

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
