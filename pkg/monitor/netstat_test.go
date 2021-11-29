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

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertDownNodesCriterion(), "failed testcase", i)
	}
}

func TestNetstatCalculator_AlertHeightCriterion(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          NodesWithStats
		expectedResult bool
	}{
		{
			criteria: NetworkErrorCriteria{
				NodesHeight: NodesHeightCriterion{
					HeightDiff:              5,
					RequireMinNodesOnHeight: 2,
				},
			},
			nodes: NodesWithStats{
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 8}},
				{NodeStats: NodeStats{Height: 4}},
				{NodeStats: NodeStats{Height: 4}},
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 8}},
				{NodeStats: NodeStats{Height: -1}},
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 11}},
				{NodeStats: NodeStats{Height: 4}},
				{NodeStats: NodeStats{Height: -1}},
			},
			expectedResult: false,
		},
	}

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertHeightCriterion(), "failed testcase", i)
	}
}

func TestNodesWithStats_SplitByStateHash(t *testing.T) {
	tests := []struct {
		criteria       NetworkErrorCriteria
		nodes          NodesWithStats
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "22", Height: 1}},
				{NodeStats: NodeStats{StateHash: "22", Height: 1}},
				{NodeStats: NodeStats{StateHash: "33", Height: 1}},
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "22", Height: 1}},
				{NodeStats: NodeStats{StateHash: "33", Height: 1}},
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "22", Height: 1}},
				{NodeStats: NodeStats{StateHash: "33", Height: 2}},
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
			nodes: NodesWithStats{
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "11", Height: 1}},
				{NodeStats: NodeStats{StateHash: "22", Height: 2}},
				{NodeStats: NodeStats{StateHash: "22", Height: 2}},
				{NodeStats: NodeStats{StateHash: "33", Height: 2}},
				{NodeStats: NodeStats{StateHash: "33", Height: 2}},
			},
			expectedResult: true,
		},
	}

	for i, tc := range tests {
		calc, err := newNetstatCalculator(tc.criteria, tc.nodes)
		require.NoError(t, err)

		require.Equal(t, tc.expectedResult, calc.AlertStateHashCriterion(), "failed testcase", i)
	}
}
