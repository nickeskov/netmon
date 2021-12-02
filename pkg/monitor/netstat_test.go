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

func TestNetstatCalculator_CurrentMaxHeight(t *testing.T) {
	tests := []struct {
		nodes          nodesWithStats
		expectedHeight int
	}{
		{
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 2}},
				{nodeStats: nodeStats{Height: 2}},
				{nodeStats: nodeStats{Height: 3}},
				{nodeStats: nodeStats{Height: 4}},
				{nodeStats: nodeStats{Height: 2}},
			},
			expectedHeight: 4,
		},
		{
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: -1}},
				{nodeStats: nodeStats{Height: -1}},
			},
			expectedHeight: -1,
		},
		{
			nodes: nodesWithStats{
				{nodeStats: nodeStats{Height: 2}},
				{nodeStats: nodeStats{Height: 3}},
			},
			expectedHeight: 3,
		},
	}
	for _, tc := range tests {
		calc, err := newNetstatCalculator(NetworkErrorCriteria{}, tc.nodes)
		require.NoError(t, err)
		require.Equal(t, tc.expectedHeight, calc.CurrentMaxHeight())
	}
}

func TestNodesDownCriterion_Validate(t *testing.T) {
	tests := []struct {
		criterion NodesDownCriterion
		ok        bool
	}{
		{NodesDownCriterion{TotalDownNodesPart: -1.0}, false},
		{NodesDownCriterion{TotalDownNodesPart: 0.0}, false},
		{NodesDownCriterion{TotalDownNodesPart: 0.5}, true},
		{NodesDownCriterion{TotalDownNodesPart: 1.0}, false},
		{NodesDownCriterion{TotalDownNodesPart: 2.0}, false},
	}
	for _, tc := range tests {
		err := tc.criterion.Validate()
		if tc.ok {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestNodesHeightCriterion_Validate(t *testing.T) {
	tests := []struct {
		criterion NodesHeightCriterion
		ok        bool
	}{
		{NodesHeightCriterion{RequireMinNodesOnHeight: -1, HeightDiff: 1}, false},
		{NodesHeightCriterion{RequireMinNodesOnHeight: 0, HeightDiff: 1}, false},
		{NodesHeightCriterion{RequireMinNodesOnHeight: 1, HeightDiff: -1}, false},
		{NodesHeightCriterion{RequireMinNodesOnHeight: 1, HeightDiff: 0}, false},
		{NodesHeightCriterion{RequireMinNodesOnHeight: -1, HeightDiff: -1}, false},
		{NodesHeightCriterion{RequireMinNodesOnHeight: 1, HeightDiff: 1}, true},
	}
	for _, tc := range tests {
		err := tc.criterion.Validate()
		if tc.ok {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestNodesStateHashCriterion_Validate(t *testing.T) {
	tests := []struct {
		criterion NodesStateHashCriterion
		ok        bool
	}{
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   -1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       -1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: -1,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          -1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   0,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       0,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 0,
				RequireMinNodesOnHeight:          1,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          0,
			},
			ok: false,
		},
		{
			criterion: NodesStateHashCriterion{
				MinStateHashGroupsOnSameHeight:   1,
				MinValuableStateHashGroups:       1,
				MinNodesInValuableStateHashGroup: 1,
				RequireMinNodesOnHeight:          1,
			},
			ok: true,
		},
	}
	for _, tc := range tests {
		err := tc.criterion.Validate()
		if tc.ok {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestNetworkErrorCriteria_Validate(t *testing.T) {
	tests := []struct {
		criteria NetworkErrorCriteria
		ok       bool
	}{
		{
			criteria: NetworkErrorCriteria{
				NodesDown:   NodesDownCriterion{TotalDownNodesPart: 0.0},
				NodesHeight: NodesHeightCriterion{1, 1},
				StateHash:   NodesStateHashCriterion{1, 1, 1, 1},
			},
			ok: false,
		},
		{
			criteria: NetworkErrorCriteria{
				NodesDown:   NodesDownCriterion{TotalDownNodesPart: 1.5},
				NodesHeight: NodesHeightCriterion{0, 1},
				StateHash:   NodesStateHashCriterion{1, 1, 1, 1},
			},
			ok: false,
		},
		{
			criteria: NetworkErrorCriteria{
				NodesDown:   NodesDownCriterion{TotalDownNodesPart: 1.5},
				NodesHeight: NodesHeightCriterion{1, 1},
				StateHash:   NodesStateHashCriterion{0, 1, 1, 1},
			},
			ok: false,
		},
	}
	for _, tc := range tests {
		err := tc.criteria.Validate()
		if tc.ok {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
