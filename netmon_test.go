package main

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

	require.Equal(t, mon.networkErrorStreak, 1)
}
