package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetstatCalculator_AlertDownNodesCriterion(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          nodesWithStats
		expectedResult bool
	}{
		{
			criteria: NetworkErrorCriteria{NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.33}},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: -1}},
			},
			expectedResult: true,
		},
		{
			criteria: NetworkErrorCriteria{NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.4}},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: -1}},
			},
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertDownNodesCriterion())
	}
}
