package main

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

	networkPrefix string
	scrapper      NodesStatsScrapper

	// state fields
	monitorState       NetworkMonitoringState
	networkErrorStreak int

	// criteria fields
	alertOnNetworkErrorStreak int
	criteria                  networkErrorCriteria
}

func NewNetworkMonitoring(
	networkPrefix string,
	nodesStatsScraper NodesStatsScrapper,
	alertOnNetworkErrorStreak int,
	criteria networkErrorCriteria,
) (NetworkMonitor, error) {
	if alertOnNetworkErrorStreak < 1 {
		return NetworkMonitor{}, errors.New("alertOnNetworkErrorStreak should be greater that zero")
	}
	return NetworkMonitor{
		monitorState:              MonitorActive,
		networkPrefix:             networkPrefix,
		scrapper:                  nodesStatsScraper,
		alertOnNetworkErrorStreak: alertOnNetworkErrorStreak,
		criteria:                  criteria,
	}, nil
}

func (m *NetworkMonitor) CheckNodes() error {
	if m.State() != MonitorActive {
		// monitor is frozen - skip check
		return nil
	}

	allNodes, err := m.scrapper.ScrapeNodeStats()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.monitorState != MonitorActive {
		return nil // monitor is frozen -
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
	case MonitorActive:
		return m.networkErrorStreak < m.alertOnNetworkErrorStreak
	case MonitorFrozenNetworkDegraded:
		return false
	case MonitorFrozenNetworkOperatesStable:
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
