package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NetworkMonitoringState int32

const (
	MonitorActive NetworkMonitoringState = iota + 1
	MonitorFrozenNetworkOperatesStable
	MonitorFrozenNetworkDegraded
)

func NewNetworkMonitoringStateFromString(state string) (NetworkMonitoringState, error) {
	switch state {
	case "active":
		return MonitorActive, nil
	case "frozen_operates_stable":
		return MonitorFrozenNetworkOperatesStable, nil
	case "frozen_degraded":
		return MonitorFrozenNetworkDegraded, nil
	default:
		return 0, errors.Errorf("failed parse network monitoring state from string: invalid state string %q", state)
	}
}

func (n NetworkMonitoringState) String() string {
	switch n {
	case MonitorActive:
		return "active"
	case MonitorFrozenNetworkOperatesStable:
		return "frozen_operates_stable"
	case MonitorFrozenNetworkDegraded:
		return "frozen_degraded"
	default:
		return fmt.Sprintf("unknown state (%d)", n)
	}
}

type NetworkMonitor struct {
	mu sync.RWMutex

	// state fields
	monitorState       NetworkMonitoringState
	networkErrorStreak int

	// data fields
	networkPrefix string
	nodesStatsUrl string

	// criteria fields
	alertOnNetworkErrorStreak int
	criteria                  networkErrorCriteria
}

func NewNetworkMonitoring(
	networkPrefix string,
	nodesStatsUrl string,
	alertOnNetworkErrorStreak int,
	criteria networkErrorCriteria,
) (NetworkMonitor, error) {
	if alertOnNetworkErrorStreak < 1 {
		return NetworkMonitor{}, errors.New("alertOnNetworkErrorStreak should be greater that zero")
	}
	return NetworkMonitor{
		monitorState:              MonitorActive,
		networkPrefix:             networkPrefix,
		nodesStatsUrl:             nodesStatsUrl,
		alertOnNetworkErrorStreak: alertOnNetworkErrorStreak,
		criteria:                  criteria,
	}, nil
}

func (n *NetworkMonitor) CheckNodes() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.monitorState != MonitorActive {
		return nil // monitor is frozen -
	}
	resp, err := http.Get(n.nodesStatsUrl)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			zap.S().Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to get nodes statuses from %q, HTTP code(%d) %q",
			n.nodesStatsUrl,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
		)
	}

	nodes := nodesWithStats{}
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return err
	}

	networkNodes := nodes.NodesWithNetworkPrefix(n.networkPrefix)
	calc, err := newNetstatCalculator(n.criteria, networkNodes)
	if err != nil {
		return err
	}

	zap.S().Debugf("stats successfully received from %q, calculate network error criteria", n.nodesStatsUrl)

	// TODO: add other types of checks
	if calc.AlertDownNodesCriterion() || calc.AlertHeightCriterion() || calc.AlertStateHashCriterion() {
		zap.S().Debugf("network %q error has been detected, increasing networkErrorStreak counter", n.networkPrefix)
		// increment error streak counter
		n.networkErrorStreak++

	} else {
		// all ok - reset streak
		zap.S().Debugf("network %q operates normally and alert hasn't been generated", n.networkPrefix)
		n.networkErrorStreak = 0
	}
	return nil
}

func (n *NetworkMonitor) NetworkOperatesStable() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch n.monitorState {
	case MonitorActive:
		return n.networkErrorStreak < n.alertOnNetworkErrorStreak
	case MonitorFrozenNetworkDegraded:
		return false
	case MonitorFrozenNetworkOperatesStable:
		return true
	default:
		panic("unknown monitor state")
	}
}

func (n *NetworkMonitor) ChangeState(state NetworkMonitoringState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.monitorState == state {
		return // state the same - do nothing
	}

	zap.S().Debugf("changing monitor state to %q", state.String())
	n.monitorState = state
	// we have to reset the streak in case of state changing
	n.networkErrorStreak = 0
}

func (n *NetworkMonitor) Run(ctx context.Context, pollNodesStatsInterval time.Duration) {
	for {
		if err := n.CheckNodes(); err != nil {
			zap.S().Errorf("failed to check nodes status: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollNodesStatsInterval):
			continue
		}
	}
}

func (n *NetworkMonitor) RunInBackground(ctx context.Context, pollNodesStatsInterval time.Duration) <-chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		n.Run(ctx, pollNodesStatsInterval)
	}()
	return done
}
