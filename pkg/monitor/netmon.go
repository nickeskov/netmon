package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	StateActive NetworkMonitoringState = iota + 1
	StateFrozenNetworkOperatesStable
	StateFrozenNetworkDegraded
)

type NetworkMonitoringState int32

func (s NetworkMonitoringState) Validate() error {
	switch s {
	case StateActive, StateFrozenNetworkOperatesStable, StateFrozenNetworkDegraded:
		return nil
	default:
		return errors.Errorf("invalid state (%d)", s)
	}
}

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

func (s NetworkMonitoringState) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateFrozenNetworkOperatesStable:
		return "frozen_operates_stable"
	case StateFrozenNetworkDegraded:
		return "frozen_degraded"
	default:
		return fmt.Sprintf("unknown state (%d)", s)
	}
}

type NetworkStatusInfo struct {
	Updated time.Time `json:"updated,omitempty"`
	Status  bool      `json:"status"`
	Height  int       `json:"height"`
}

type NetworkMonitor struct {
	mu sync.RWMutex

	networkPrefix string
	scrapper      NodesStatsScrapper

	// state fields
	monitorState       NetworkMonitoringState
	statsHistory       statsHistoryDeque
	networkErrorStreak int

	// criteria fields
	alertOnNetworkErrorStreak int
	criteria                  NetworkErrorCriteria
}

func NewNetworkMonitoring(
	initialState NetworkMonitoringState,
	networkPrefix string,
	maxStatsHistoryLen int,
	nodesStatsScraper NodesStatsScrapper,
	alertOnNetworkErrorStreak int,
	criteria NetworkErrorCriteria,
) (NetworkMonitor, error) {
	if maxStatsHistoryLen < 1 {
		return NetworkMonitor{}, errors.New("maxStatsHistoryLen should be greater than zero")
	}
	if alertOnNetworkErrorStreak < 1 {
		return NetworkMonitor{}, errors.New("alertOnNetworkErrorStreak should be greater than zero")
	}
	if err := initialState.Validate(); err != nil {
		return NetworkMonitor{}, err
	}
	return NetworkMonitor{
		monitorState:              initialState,
		networkPrefix:             networkPrefix,
		scrapper:                  nodesStatsScraper,
		statsHistory:              newStatsDeque(maxStatsHistoryLen),
		alertOnNetworkErrorStreak: alertOnNetworkErrorStreak,
		criteria:                  criteria,
	}, nil
}

func (m *NetworkMonitor) CheckNodes(now time.Time) error {
	if state := m.State(); state != StateActive {
		zap.S().Debugf("monitor is frozen, current state is %q", state)
		return nil
	}

	allNetworksNodes, err := m.scrapper.ScrapeNodeStats()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if state := m.monitorState; state != StateActive {
		zap.S().Debugf("monitor is frozen, current state is %q", state)
		return nil
	}

	currentNetworkNodes := allNetworksNodes.NodesWithNetworkPrefix(m.networkPrefix)
	calc, err := newNetstatCalculator(m.criteria, currentNetworkNodes)
	if err != nil {
		return err
	}

	statsSnapshot := &statsDataSnapshot{
		snapshotCreationTime: now,
		nodes:                currentNetworkNodes,
		maxHeight:            calc.CurrentMaxHeight(),
		nodesDownCriterion:   calc.AlertDownNodesCriterion(),
		heightCriterion:      calc.AlertHeightCriterion(),
		stateHashCriterion:   calc.AlertStateHashCriterion(),
	}
	m.statsHistory.PushFront(statsSnapshot)

	if statsSnapshot.nodesDownCriterion || statsSnapshot.heightCriterion || statsSnapshot.stateHashCriterion {
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

func (m *NetworkMonitor) NetworkStatusInfo() NetworkStatusInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusInfo := NetworkStatusInfo{
		Status: m.unsafeNetworkOperatesStable(),
		Height: -1,
	}
	if m.statsHistory.Len() != 0 {
		front := m.statsHistory.Front()

		statusInfo.Height = front.maxHeight
		statusInfo.Updated = front.snapshotCreationTime
	}
	return statusInfo
}

func (m *NetworkMonitor) NetworkOperatesStable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.unsafeNetworkOperatesStable()
}

func (m *NetworkMonitor) unsafeNetworkOperatesStable() bool {
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
		if err := m.CheckNodes(time.Now().UTC()); err != nil {
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
