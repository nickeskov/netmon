package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NetworkMonitoringState int32

const (
	StateActive NetworkMonitoringState = iota + 1
	StateFrozenNetworkOperatesStable
	StateFrozenNetworkDegraded
)

func NewNetworkMonitoringStateFromString(state string) (NetworkMonitoringState, error) {
	switch state {
	case "active":
		return StateActive, nil
	case "frozen_operates_stable":
		return StateFrozenNetworkOperatesStable, nil
	case "frozen_degraded":
		return StateFrozenNetworkDegraded, nil
	default:
		return 0, errors.Errorf("failed parse network monitoring state from string, invalid state string %q", state)
	}
}

func (n NetworkMonitoringState) String() string {
	switch n {
	case StateActive:
		return "active"
	case StateFrozenNetworkOperatesStable:
		return "frozen_operates_stable"
	case StateFrozenNetworkDegraded:
		return "frozen_degraded"
	default:
		return fmt.Sprintf("unknown state (%d)", n)
	}
}

type NetworkMonitor struct {
	mu sync.RWMutex

	networkPrefix string
	scrapper      NodesStatsScrapper

	// state fields
	monitorState       NetworkMonitoringState
	networkErrorStreak int

	// criteria fields
	alertOnNetworkErrorStreak int
	criteria                  NetworkErrorCriteria
}

func NewNetworkMonitoring(
	networkPrefix string,
	nodesStatsScraper NodesStatsScrapper,
	alertOnNetworkErrorStreak int,
	criteria NetworkErrorCriteria,
) (NetworkMonitor, error) {
	if alertOnNetworkErrorStreak < 1 {
		return NetworkMonitor{}, errors.New("alertOnNetworkErrorStreak should be greater that zero")
	}
	return NetworkMonitor{
		monitorState:              StateActive,
		networkPrefix:             networkPrefix,
		scrapper:                  nodesStatsScraper,
		alertOnNetworkErrorStreak: alertOnNetworkErrorStreak,
		criteria:                  criteria,
	}, nil
}

func (m *NetworkMonitor) CheckNodes() error {
	if m.State() != StateActive {
		// monitor is frozen - skip check
		return nil
	}

	allNodes, err := m.scrapper.ScrapeNodeStats()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.monitorState != StateActive {
		// monitor is frozen - skip check
		return nil
	}

	calc, err := newNetstatCalculator(m.criteria, allNodes.NodesWithNetworkPrefix(m.networkPrefix))
	if err != nil {
		return err
	}

	if calc.AlertDownNodesCriterion() || calc.AlertHeightCriterion() || calc.AlertStateHashCriterion() {
		zap.S().Debugf("network %q error has been detected, increasing networkErrorStreak counter", m.networkPrefix)
		// increment error streak counter
		m.networkErrorStreak++
	} else {
		// all ok - reset streak
		zap.S().Debugf("network %q operates normally and alert hasn't been generated", m.networkPrefix)
		m.networkErrorStreak = 0
	}
	return nil
}

func (m *NetworkMonitor) NetworkOperatesStable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.monitorState {
	case StateActive:
		return m.networkErrorStreak < m.alertOnNetworkErrorStreak
	case StateFrozenNetworkDegraded:
		return false
	case StateFrozenNetworkOperatesStable:
		return true
	default:
		panic("unknown monitor state")
	}
}

func (m *NetworkMonitor) State() NetworkMonitoringState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.monitorState
}

func (m *NetworkMonitor) ChangeState(state NetworkMonitoringState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.monitorState == state {
		return // state the same - do nothing
	}

	zap.S().Debugf("changing monitor state to %q", state.String())
	m.monitorState = state
	// we have to reset the streak in case of state changing
	m.networkErrorStreak = 0
}

func (m *NetworkMonitor) Run(ctx context.Context, pollNodesStatsInterval time.Duration) {
	for {
		if err := m.CheckNodes(); err != nil {
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

func (m *NetworkMonitor) RunInBackground(ctx context.Context, pollNodesStatsInterval time.Duration) <-chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		m.Run(ctx, pollNodesStatsInterval)
	}()
	return done
}
