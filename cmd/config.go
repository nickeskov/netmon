package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type appConfig struct {
	logLevel               string
	bindAddr               string
	networkScheme          string
	nodeStatsURL           string
	pollNodesStatsInterval time.Duration
	statsHistorySize       int
	networkErrorsStreak    int
	initialMonState        string

	httpAuthHeader string
	httpAuthToken  string

	criterionNodesDownTotalPart float64

	criterionNodesHeightDiff                    int
	criterionNodesHeightRequireMinNodesOnHeight int

	criterionNodesStateHashMinStateHashGroupsOnSameHeight   int
	criterionNodesStateHashMinValuableStateHashGroups       int
	criterionNodesStateHashMinNodesInValuableStateHashGroup int
	criterionNodesStateHashRequireMinNodesOnHeight          int
}

func (c *appConfig) parseENVAndRegisterCLI(l *zap.SugaredLogger) {
	flag.StringVar(&c.logLevel, "log-level", lookupEnvOrString("LOG_LEVEL", "INFO"), "Logging level. Supported levels: 'DEV', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'. ENV: 'LOG_LEVEL'.")
	flag.StringVar(&c.bindAddr, "bind-addr", lookupEnvOrString("BIND_ADDR", ":2048"), "Local network address to bind the HTTP API of the service on. ENV: 'BIND_ADDR'.")
	flag.StringVar(&c.networkScheme, "network-scheme", lookupEnvOrString("NETWORK_SCHEME", "W"), "WAVES network scheme character. Supported networks: 'W' (mainnet), 'T' (testnet), 'S' (stagenet). ENV: 'NETWORK_SCHEME'.")
	flag.StringVar(&c.nodeStatsURL, "stats-url", lookupEnvOrString("STATS_URL", "https://waves-nodes-get-height.wavesnodes.com/"), "Nodes statistics URL. ENV: 'STATS_URL'.")
	flag.DurationVar(&c.pollNodesStatsInterval, "stats-poll-interval", lookupEnvOrDuration(l, "STATS_POLL_INTERVAL", time.Minute), "Nodes statistics polling interval. ENV: 'STATS_POLL_INTERVAL'.")
	flag.IntVar(&c.statsHistorySize, "stats-history-size", lookupEnvOrInt(l, "STATS_HISTORY_SIZE", 10), "Exact amount of latest nodes stats that will be kept. ENV: 'STATS_HISTORY_SIZE'.")
	flag.IntVar(&c.networkErrorsStreak, "network-errors-streak", lookupEnvOrInt(l, "NETWORK_ERRORS_STREAK", 5), "Network will be considered as degraded after that errors streak. ENV: 'NETWORK_ERRORS_STREAK'.")
	flag.StringVar(&c.initialMonState, "initial-mon-state", lookupEnvOrString("INITIAL_MON_STATE", "active"), "Initial monitoring state. Possible states: 'active', 'frozen_operates_stable', 'frozen_degraded'. ENV: 'INITIAL_MON_STATE'.")

	flag.StringVar(&c.httpAuthHeader, "http-auth-header", lookupEnvOrString("HTTP_AUTH_HEADER", "X-Waves-Monitor-Auth"), "HTTP header which will be used for private routes authentication. ENV: 'HTTP_AUTH_HEADER'.")
	flag.StringVar(&c.httpAuthToken, "http-auth-token", lookupEnvOrString("HTTP_AUTH_TOKEN", ""), "HTTP auth token which will be used for private routes authentication. ENV: 'HTTP_AUTH_TOKEN'.")

	flag.Float64Var(&c.criterionNodesDownTotalPart, "criterion-down-total-part", lookupEnvOrFloat64(l, "CRITERION_DOWN_TOTAL_PART", 0.3), "Alert will be generated if detected down nodes part greater than that criterion. ENV: 'CRITERION_DOWN_TOTAL_PART'.")

	flag.IntVar(&c.criterionNodesHeightDiff, "criterion-height-diff", lookupEnvOrInt(l, "CRITERION_HEIGHT_DIFF", 5), "Alert will be generated if detected height diff greater than that criterion. ENV: 'CRITERION_HEIGHT_DIFF'.")
	flag.IntVar(&c.criterionNodesHeightRequireMinNodesOnHeight, "criterion-height-require-min-nodes-on-same-height", lookupEnvOrInt(l, "CRITERION_HEIGHT_REQUIRE_MIN_NODES_ON_SAME_HEIGHT", 2), "Minimum required amount of nodes on same height for height-diff criterion. ENV: 'CRITERION_HEIGHT_REQUIRE_MIN_NODES_ON_SAME_HEIGHT'.")

	flag.IntVar(&c.criterionNodesStateHashMinStateHashGroupsOnSameHeight, "criterion-statehash-min-groups-on-same-height", lookupEnvOrInt(l, "CRITERION_STATEHASH_MIN_GROUPS_ON_SAME_HEIGHT", 2), "Alert won't be generated if detected amount of statehash groups on same height lower than that criterion. ENV: 'CRITERION_STATEHASH_MIN_GROUPS_ON_SAME_HEIGHT'.")
	flag.IntVar(&c.criterionNodesStateHashMinValuableStateHashGroups, "criterion-statehash-min-valuable-groups", lookupEnvOrInt(l, "CRITERION_STATEHASH_MIN_VALUABLE_GROUPS", 2), "Alert won't be generated if detected amount of statehash 'valuable' groups on same height lower than that criterion. ENV: 'CRITERION_STATEHASH_MIN_VALUABLE_GROUPS'.")
	flag.IntVar(&c.criterionNodesStateHashMinNodesInValuableStateHashGroup, "criterion-statehash-min-nodes-in-valuable-group", lookupEnvOrInt(l, "CRITERION_STATEHASH_MIN_NODES_IN_VALUABLE_GROUP", 2), "StateHash group will be considered as 'valuable' if contains 'criterion-statehash-min-valuable-groups'. ENV: 'CRITERION_STATEHASH_MIN_NODES_IN_VALUABLE_GROUP'.")
	flag.IntVar(&c.criterionNodesStateHashRequireMinNodesOnHeight, "criterion-statehash-require-min-nodes-on-same-height", lookupEnvOrInt(l, "CRITERION_STATEHASH_REQUIRE_MIN_NODES_ON_SAME_HEIGHT", 4), "Minimum required amount of nodes on same height for statehash criterion. ENV: 'CRITERION_STATEHASH_REQUIRE_MIN_NODES_ON_SAME_HEIGHT'.")
}

func lookupEnvOrString(envKey string, defaultVal string) string {
	if val, ok := os.LookupEnv(envKey); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrFloat64(l *zap.SugaredLogger, envKey string, defaultVal float64) float64 {
	if val, ok := os.LookupEnv(envKey); ok {
		intVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			l.Fatalf("failed to parse %q env variable value=%q as 'float64': %v", envKey, val, err)
		}
		return intVal
	}
	return defaultVal
}

func lookupEnvOrInt(l *zap.SugaredLogger, envKey string, defaultVal int) int {
	if val, ok := os.LookupEnv(envKey); ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			l.Fatalf("failed to parse %q env variable value=%q as 'int': %v", envKey, val, err)
		}
		return intVal
	}
	return defaultVal
}

func lookupEnvOrDuration(l *zap.SugaredLogger, envKey string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(envKey); ok {
		intVal, err := time.ParseDuration(val)
		if err != nil {
			l.Fatalf("failed to parse %q env variable value=%q as 'Duration': %v", envKey, val, err)
		}
		return intVal
	}
	return defaultVal
}
