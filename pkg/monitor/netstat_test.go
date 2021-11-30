package monitor

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

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertDownNodesCriterion(), "failed testcase #%d", i)
	}
}

func TestNetstatCalculator_AlertHeightCriterion(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          nodesWithStats
		expectedResult bool
	}{
		{
			criteria: NetworkErrorCriteria{
				NodesHeight: NodesHeightCriterion{
					HeightDiff:              5,
					RequireMinNodesOnHeight: 2,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 8}},
				{nodeStats: nodeStats{Height: 4}},
				{nodeStats: nodeStats{Height: 4}},
			},
			expectedResult: true,
		},
		{
			criteria: NetworkErrorCriteria{
				NodesHeight: NodesHeightCriterion{
					HeightDiff:              5,
					RequireMinNodesOnHeight: 2,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 8}},
				{nodeStats: nodeStats{Height: -1}},
			},
			expectedResult: false,
		},
		{
			criteria: NetworkErrorCriteria{
				NodesHeight: NodesHeightCriterion{
					HeightDiff:              5,
					RequireMinNodesOnHeight: 2,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 11}},
				{nodeStats: nodeStats{Height: 4}},
				{nodeStats: nodeStats{Height: -1}},
			},
			expectedResult: false,
		},
	}

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertHeightCriterion(), "failed testcase #%d", i)
	}
}

func TestNodesWithStats_SplitByStateHash(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          nodesWithStats
		expectedResult bool
	}{
		{
			criteria: NetworkErrorCriteria{
				StateHash: NodesStateHashCriterion{
					MinStateHashGroupsOnSameHeight:   2,
					MinValuableStateHashGroups:       2,
					MinNodesInValuableStateHashGroup: 2,
					RequireMinNodesOnHeight:          4,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "22", Height: 1}},
				{nodeStats: nodeStats{StateHash: "22", Height: 1}},
				{nodeStats: nodeStats{StateHash: "33", Height: 1}},
			},
			expectedResult: true,
		},
		{
			criteria: NetworkErrorCriteria{
				StateHash: NodesStateHashCriterion{
					MinStateHashGroupsOnSameHeight:   2,
					MinValuableStateHashGroups:       2,
					MinNodesInValuableStateHashGroup: 2,
					RequireMinNodesOnHeight:          4,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "22", Height: 1}},
				{nodeStats: nodeStats{StateHash: "33", Height: 1}},
			},
			expectedResult: false,
		},
		{
			criteria: NetworkErrorCriteria{
				StateHash: NodesStateHashCriterion{
					MinStateHashGroupsOnSameHeight:   2,
					MinValuableStateHashGroups:       2,
					MinNodesInValuableStateHashGroup: 1,
					RequireMinNodesOnHeight:          1,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "22", Height: 1}},
				{nodeStats: nodeStats{StateHash: "33", Height: 2}},
			},
			expectedResult: true,
		},

		{
			criteria: NetworkErrorCriteria{
				StateHash: NodesStateHashCriterion{
					MinStateHashGroupsOnSameHeight:   2,
					MinValuableStateHashGroups:       2,
					MinNodesInValuableStateHashGroup: 2,
					RequireMinNodesOnHeight:          2,
				},
			},
			nodes: nodesWithStats{
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "11", Height: 1}},
				{nodeStats: nodeStats{StateHash: "22", Height: 2}},
				{nodeStats: nodeStats{StateHash: "22", Height: 2}},
				{nodeStats: nodeStats{StateHash: "33", Height: 2}},
				{nodeStats: nodeStats{StateHash: "33", Height: 2}},
			},
			expectedResult: true,
		},
	}

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertStateHashCriterion(), "failed testcase #%d", i)
	}
}
