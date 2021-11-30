package monitor

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNetworkMonitor_CheckNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	scraperMock := NewMockNodesStatsScrapper(ctrl)
	scraperMock.EXPECT().ScrapeNodeStats().Times(1).Return(
		nodesWithStats{
			{nodeStats: nodeStats{Height: 11}},
			{nodeStats: nodeStats{Height: 11}},
			{nodeStats: nodeStats{Height: -1}},
		},
		nil,
	)

	mon, err := NewNetworkMonitoring(
		StateActive,
		"",
		scraperMock,
		1,
		NetworkErrorCriteria{
			NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.3},
		},
	)
	require.NoError(t, err)

	err = mon.CheckNodes()
	require.NoError(t, err)

	require.False(t, mon.NetworkOperatesStable())
	require.Equal(t, mon.networkErrorStreak, 1)
}

func TestNetworkMonitor_ChangeState(t *testing.T) {
	mon, err := NewNetworkMonitoring(
		StateActive,
		"",
		nil,
		5,
		NetworkErrorCriteria{},
	)
	require.NoError(t, err)

	require.Equal(t, StateActive, mon.State())

	mon.networkErrorStreak = 10
	mon.ChangeState(StateActive)
	require.Equal(t, 10, mon.networkErrorStreak)

	mon.ChangeState(StateFrozenNetworkDegraded)
	require.Equal(t, 0, mon.networkErrorStreak)
}

func TestNetworkMonitor_NetworkOperatesStable(t *testing.T) {
	tests := []struct {
		networkErrorStreak int
		operatesStable     bool
		state              NetworkMonitoringState
	}{
		{6, false, StateActive},
		{5, false, StateActive},
		{4, true, StateActive},

		{0, false, StateFrozenNetworkDegraded},
		{5, false, StateFrozenNetworkDegraded},
		{10, false, StateFrozenNetworkDegraded},

		{10, true, StateFrozenNetworkOperatesStable},
		{5, true, StateFrozenNetworkOperatesStable},
		{6, true, StateFrozenNetworkOperatesStable},
		{0, true, StateFrozenNetworkOperatesStable},
	}

	for i, tc := range tests {
		mon, err := NewNetworkMonitoring(
			tc.state,
			"",
			nil,
			5,
			NetworkErrorCriteria{},
		)
		require.NoError(t, err)
		mon.networkErrorStreak = tc.networkErrorStreak
		require.Equal(t, tc.operatesStable, mon.NetworkOperatesStable(), "failed testcase #%d", i)
	}
}
