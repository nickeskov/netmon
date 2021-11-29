package monitor

import (
	"encoding/json"
	"strings"
)

// NodeStats is basic node statistics
type NodeStats struct {
	Height          int    `json:"height"`
	StateHash       string `json:"statehash"`
	StateHashHeight string `json:"statehash_height"`
	Version         string `json:"version"`
}

type nodeWithStats struct {
	NodeDomain string `json:"_"`
	NodeStats
}

type NodesWithStats []nodeWithStats

func (n *NodesWithStats) UnmarshalJSON(bytes []byte) error {
	var nodesStats map[string]NodeStats
	if err := json.Unmarshal(bytes, &nodesStats); err != nil {
		return err
	}
	nodes := make(NodesWithStats, 0, len(nodesStats))

	for domain, stats := range nodesStats {
		node := nodeWithStats{
			NodeDomain: domain,
			NodeStats:  stats,
		}
		nodes = append(nodes, node)
	}
	*n = nodes
	return nil
}

func (n NodesWithStats) ForEach(fn func(node *nodeWithStats) error) error {
	for i := range n {
		node := &n[i]
		if err := fn(node); err != nil {
			return err
		}
	}
	return nil
}

func (n NodesWithStats) Filter(condition func(node *nodeWithStats) bool) NodesWithStats {
	var nodes NodesWithStats
	for _, node := range n {
		if condition(&node) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (n NodesWithStats) NodesWithNetworkPrefix(networkPrefix string) NodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return strings.HasPrefix(node.NodeDomain, networkPrefix)
	})
}

func (n NodesWithStats) WorkingNodes() NodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return node.Height > 0
	})
}

func (n NodesWithStats) DownNodes() NodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return node.Height == -1
	})
}

func (n NodesWithStats) SplitByHeight() map[int]NodesWithStats {
	splitMap := make(map[int]NodesWithStats)
	for _, node := range n {
		splitMap[node.Height] = append(splitMap[node.Height], node)
	}
	return splitMap
}

func (n NodesWithStats) SplitByVersion() map[string]NodesWithStats {
	splitMap := make(map[string]NodesWithStats)
	for _, node := range n {
		splitMap[node.Version] = append(splitMap[node.Version], node)
	}
	return splitMap
}

func (n NodesWithStats) SplitByStateHash() map[string]NodesWithStats {
	splitMap := make(map[string]NodesWithStats)
	for _, node := range n {
		splitMap[node.StateHash] = append(splitMap[node.StateHash], node)
	}
	return splitMap
}
