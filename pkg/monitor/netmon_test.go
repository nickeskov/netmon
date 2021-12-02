package monitor

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNetworkMonitor_CheckNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	scraperMock := NewMockNodesStatsScrapper(ctrl)
	scraperMock.EXPECT().ScrapeNodeStats().Times(1).Return(
		nodesWithStats{
			{nodeStats: nodeStats{Height: 11, NetByte: MainNetSchemeChar}},
			{nodeStats: nodeStats{Height: 11, NetByte: MainNetSchemeChar}},
			{nodeStats: nodeStats{Height: -1, NetByte: MainNetSchemeChar}},
		},
		nil,
	)

	mon, err := NewNetworkMonitoring(
		StateActive,
		MainNetSchemeChar,
		10,
		scraperMock,
		1,
		NetworkErrorCriteria{
			NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.3},
		},
	)
	require.NoError(t, err)

	now := time.Now()

	err = mon.CheckNodes(now)
	require.NoError(t, err)

	require.False(t, mon.NetworkOperatesStable())
	require.Equal(t, mon.networkErrorStreak, 1)

	expectedInfo := NetworkStatusInfo{
		Updated: now,
		Status:  false,
		Height:  11,
	}
	require.Equal(t, expectedInfo, mon.NetworkStatusInfo())
}

func TestNetworkMonitor_ChangeState(t *testing.T) {
	mon, err := NewNetworkMonitoring(
		StateActive,
		MainNetSchemeChar,
		10,
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
			MainNetSchemeChar,
			10,
			nil,
			5,
			NetworkErrorCriteria{},
		)
		require.NoError(t, err)
		mon.networkErrorStreak = tc.networkErrorStreak
		require.Equal(t, tc.operatesStable, mon.NetworkOperatesStable(), "failed testcase #%d", i)
	}
}

func TestNetworkMonitor_NetworkStatusInfo(t *testing.T) {
	tests := []struct {
		networkErrorStreak int
		operatesStable     bool
		height             int
	}{
		{6, false, 10},
		{5, false, 20},
		{4, true, 30},
	}

	for i, tc := range tests {
		mon, err := NewNetworkMonitoring(
			StateActive,
			MainNetSchemeChar,
			10,
			nil,
			5,
			NetworkErrorCriteria{},
		)
		require.NoError(t, err)
		mon.networkErrorStreak = tc.networkErrorStreak

		now := time.Now()

		back := mon.statsHistory.PushFront(&statsDataSnapshot{maxHeight: tc.height, snapshotCreationTime: now})
		require.Nil(t, back)

		expected := NetworkStatusInfo{
			Updated: now,
			Status:  tc.operatesStable,
			Height:  tc.height,
		}
		actual := mon.NetworkStatusInfo()
		require.Equal(t, expected, actual, "failed testcase #%d", i)
	}
}
