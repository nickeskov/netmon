package main

import (
	"math"

	"github.com/pkg/errors"
)

// TODO: add criteria validation

type nodesDownCriterion struct {
	totalDownNodesPercentage         float64
	nodesDownOnSameVersionPercentage float64
	requireMinNodesOnSameVersion     int // minimum required count of nodes that have save version to activate this criterion
}

type heightCriterion struct {
	heightDiff              int // TODO: check version
	requireMinNodesOnHeight int // minimum required count of nodes on the same height to activate this criterion
}

type stateHashCriterion struct {
	minStateHashGroupsOnSameHeight int // I guess, default minimum should be 2

	minValuableStateHashGroups       int // I guess, default minimum should be 2
	minNodesInValuableStateHashGroup int // I guess, default minimum should be 2

	requireMinNodesOnHeight int // minimum required count of nodes on the same height to activate this criterion
}

type networkErrorCriteria struct {
	nodesDown nodesDownCriterion
	height    heightCriterion
	stateHash stateHashCriterion
}

type netstatCalculator struct {
	criteria             networkErrorCriteria
	allNodes             nodesWithStats
	workingNodes         nodesWithStats
	workingNodesOnHeight map[int]nodesWithStats
}

func newNetstatCalculator(criteria networkErrorCriteria, allNodes nodesWithStats) (netstatCalculator, error) {
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

	if totalDownPercentage >= n.criteria.nodesDown.totalDownNodesPercentage {
		return true
	}

	allNodesByVersion := n.allNodes.SplitByVersion()
	for version, downNodesWithVersion := range downNodes.SplitByVersion() {
		// check requirement
		if len(downNodesWithVersion) < n.criteria.nodesDown.requireMinNodesOnSameVersion {
			continue
		}

		// check criterion
		nodesDownPercentageOnSameHeight := float64(len(downNodesWithVersion)) / float64(len(allNodesByVersion[version]))

		if nodesDownPercentageOnSameHeight >= n.criteria.nodesDown.nodesDownOnSameVersionPercentage {
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
	if len(n.workingNodesOnHeight[minHeight]) < n.criteria.height.requireMinNodesOnHeight ||
		len(n.workingNodesOnHeight[maxHeight]) < n.criteria.height.requireMinNodesOnHeight {
		return false
	}
	// check criteria
	if maxHeight-minHeight >= n.criteria.height.heightDiff {
		return true
	}
	return false
}

func (n *netstatCalculator) AlertStateHashCriterion() bool {
	for _, nodesOnHeight := range n.workingNodesOnHeight {
		// check requirement
		if len(nodesOnHeight) < n.criteria.stateHash.requireMinNodesOnHeight {
			continue
		}

		splitByStateHash := nodesOnHeight.SplitByStateHash()

		// first criteria part
		if len(splitByStateHash) < n.criteria.stateHash.minStateHashGroupsOnSameHeight {
			continue
		}

		valuableGroupsCnt := 0
		// check second criterion, count valuable groups
		for _, nodesOnHeightWithSameStateHash := range splitByStateHash {
			if len(nodesOnHeightWithSameStateHash) >= n.criteria.stateHash.minNodesInValuableStateHashGroup {
				valuableGroupsCnt++
			}
		}
		// several node groups with different stateHash
		if valuableGroupsCnt >= n.criteria.stateHash.minValuableStateHashGroups {
			return true
		}
	}
	return false
}
