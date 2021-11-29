package monitor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetstatCalculator_AlertDownNodesCriterion(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          NodesWithStats
		expectedResult bool
	}{
		{
			criteria: NetworkErrorCriteria{NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.33}},
			nodes: NodesWithStats{
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: -1}},
			},
			expectedResult: true,
		},
		{
			criteria: NetworkErrorCriteria{NodesDown: NodesDownCriterion{TotalDownNodesPart: 0.4}},
			nodes: NodesWithStats{
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: -1}},
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
