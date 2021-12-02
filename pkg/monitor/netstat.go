package monitor

import (
	"math"

	"github.com/pkg/errors"
)

type NodesDownCriterion struct {
	TotalDownNodesPart float64
}

func (c *NodesDownCriterion) Validate() error {
	if c.TotalDownNodesPart <= 0 || c.TotalDownNodesPart >= 1 {
		return errors.Errorf("NodesDownCriterion.TotalDownNodesPart value should be 0.0 < n < 1.0")
	}
	return nil
}

type NodesHeightCriterion struct {
	HeightDiff              int
	RequireMinNodesOnHeight int // minimum required count of nodes on the same height to activate this criterion
}

func (c *NodesHeightCriterion) Validate() error {
	if c.HeightDiff <= 0 {
		return errors.Errorf("NodesHeightCriterion.HeightDiff value should be greater than zero")
	}
	if c.RequireMinNodesOnHeight <= 0 {
		return errors.Errorf("NodesHeightCriterion.RequireMinNodesOnHeight value should be greater than zero")
	}
	return nil
}

type NodesStateHashCriterion struct {
	MinStateHashGroupsOnSameHeight   int
	MinValuableStateHashGroups       int
	MinNodesInValuableStateHashGroup int
	RequireMinNodesOnHeight          int // minimum required count of nodes on the same height to activate this criterion
}

func (c *NodesStateHashCriterion) Validate() error {
	if c.MinStateHashGroupsOnSameHeight <= 0 {
		return errors.Errorf("NodesStateHashCriterion.MinStateHashGroupsOnSameHeight value should be greater than zero")
	}
	if c.MinValuableStateHashGroups <= 0 {
		return errors.Errorf("NodesStateHashCriterion.MinValuableStateHashGroups value should be greater than zero")
	}
	if c.MinNodesInValuableStateHashGroup <= 0 {
		return errors.Errorf("NodesStateHashCriterion.MinNodesInValuableStateHashGroup value should be greater than zero")
	}
	if c.RequireMinNodesOnHeight <= 0 {
		return errors.Errorf("NodesStateHashCriterion.RequireMinNodesOnHeight value should be greater than zero")
	}
	return nil
}

type NetworkErrorCriteria struct {
	NodesDown   NodesDownCriterion
	NodesHeight NodesHeightCriterion
	StateHash   NodesStateHashCriterion
}

func (c *NetworkErrorCriteria) Validate() error {
	if err := c.NodesDown.Validate(); err != nil {
		return err
	}
	if err := c.NodesHeight.Validate(); err != nil {
		return err
	}
	if err := c.StateHash.Validate(); err != nil {
		return err
	}
	return nil
}

type netstatCalculator struct {
	criteria             NetworkErrorCriteria
	allNodes             nodesWithStats
	downNodes            nodesWithStats
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
		downNodes:            allNodes.DownNodes(),
		workingNodes:         workingNodes,
		workingNodesOnHeight: workingNodes.SplitByHeight(),
	}, nil
}

func (n *netstatCalculator) AlertDownNodesCriterion() bool {
	totalDownPart := float64(len(n.downNodes)) / float64(len(n.allNodes))
	return totalDownPart >= n.criteria.NodesDown.TotalDownNodesPart
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
