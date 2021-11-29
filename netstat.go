package main

import (
	"math"

	"github.com/pkg/errors"
)

// TODO: add criteria validation

type NodesDownCriterion struct {
	TotalDownNodesPercentage         float64
	NodesDownOnSameVersionPercentage float64
	RequireMinNodesOnSameVersion     int // minimum required count of nodes that have save version to activate this criterion
}

type NodesHeightCriterion struct {
	HeightDiff              int // TODO: check version
	RequireMinNodesOnHeight int // minimum required count of nodes on the same height to activate this criterion
}

type NodesStateHashCriterion struct {
	MinStateHashGroupsOnSameHeight   int // I guess, default minimum should be 2
	MinValuableStateHashGroups       int // I guess, default minimum should be 2
	MinNodesInValuableStateHashGroup int // I guess, default minimum should be 2
	RequireMinNodesOnHeight          int // minimum required count of nodes on the same height to activate this criterion
}

type NetworkErrorCriteria struct {
	NodesDown   NodesDownCriterion
	NodesHeight NodesHeightCriterion
	StateHash   NodesStateHashCriterion
}

type netstatCalculator struct {
	criteria             NetworkErrorCriteria
	allNodes             nodesWithStats
	workingNodes         nodesWithStats
	workingNodesOnHeight map[int]nodesWithStats
}

func newNetstatCalculator(criteria NetworkErrorCriteria, allNodes nodesWithStats) (netstatCalculator, error) {
	if len(allNodes) == 0 {
		return netstatCalculator{}, errors.New("nodes with stats are empty")
	}
	workingNodes := allNodes.WorkingNodes()
	return netstatCalculator{
		criteria:             criteria,
		allNodes:             allNodes,
		workingNodes:         workingNodes,
		workingNodesOnHeight: workingNodes.SplitByHeight(),
	}, nil
}

func (n *netstatCalculator) AlertDownNodesCriterion() bool {
	downNodes := n.allNodes.DownNodes()
	totalDownPercentage := float64(len(downNodes)) / float64(len(n.allNodes))

	if totalDownPercentage >= n.criteria.NodesDown.TotalDownNodesPercentage {
		return true
	}

	allNodesByVersion := n.allNodes.SplitByVersion()
	for version, downNodesWithVersion := range downNodes.SplitByVersion() {
		// check requirement
		if len(downNodesWithVersion) < n.criteria.NodesDown.RequireMinNodesOnSameVersion {
			continue
		}

		// check criterion
		nodesDownPercentageOnSameHeight := float64(len(downNodesWithVersion)) / float64(len(allNodesByVersion[version]))

		if nodesDownPercentageOnSameHeight >= n.criteria.NodesDown.NodesDownOnSameVersionPercentage {
			return true
		}
	}
	return false
}

func (n *netstatCalculator) AlertHeightCriterion() bool {
	minHeight := math.MaxInt
	maxHeight := math.MinInt

	for height := range n.workingNodesOnHeight {
		if maxHeight < height {
			maxHeight = height
		}
		if minHeight > height {
			minHeight = height
		}
	}

	// check criteria requirement
	if len(n.workingNodesOnHeight[minHeight]) < n.criteria.NodesHeight.RequireMinNodesOnHeight ||
		len(n.workingNodesOnHeight[maxHeight]) < n.criteria.NodesHeight.RequireMinNodesOnHeight {
		return false
	}
	// check criteria
	if maxHeight-minHeight >= n.criteria.NodesHeight.HeightDiff {
		return true
	}
	return false
}

func (n *netstatCalculator) AlertStateHashCriterion() bool {
	for _, nodesOnHeight := range n.workingNodesOnHeight {
		// check requirement
		if len(nodesOnHeight) < n.criteria.StateHash.RequireMinNodesOnHeight {
			continue
		}

		splitByStateHash := nodesOnHeight.SplitByStateHash()

		// first criteria part
		if len(splitByStateHash) < n.criteria.StateHash.MinStateHashGroupsOnSameHeight {
			continue
		}

		valuableGroupsCnt := 0
		// check second criterion, count valuable groups
		for _, nodesOnHeightWithSameStateHash := range splitByStateHash {
			if len(nodesOnHeightWithSameStateHash) >= n.criteria.StateHash.MinNodesInValuableStateHashGroup {
				valuableGroupsCnt++
			}
		}
		// several node groups with different stateHash
		if valuableGroupsCnt >= n.criteria.StateHash.MinValuableStateHashGroups {
			return true
		}
	}
	return false
}
