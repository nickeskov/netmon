package monitor

import (
	"encoding/json"
)

// nodeStats is basic node statistics
type nodeStats struct {
	NetByte         NetworkSchemeChar `json:"netbyte"`
	Height          int               `json:"height"`
	StateHash       string            `json:"statehash"`
	StateHashHeight string            `json:"statehash_height"`
	Version         string            `json:"version"`
}

type nodeWithStats struct {
	NodeDomain string `json:"_"`
	nodeStats
}

type nodesWithStats []nodeWithStats

func (n *nodesWithStats) UnmarshalJSON(bytes []byte) error {
	var nodesStats map[string]nodeStats
	if err := json.Unmarshal(bytes, &nodesStats); err != nil {
		return err
	}
	nodes := make(nodesWithStats, 0, len(nodesStats))

	for domain, stats := range nodesStats {
		node := nodeWithStats{
			NodeDomain: domain,
			nodeStats:  stats,
		}
		nodes = append(nodes, node)
	}
	*n = nodes
	return nil
}

func (n nodesWithStats) Filter(condition func(node *nodeWithStats) bool) nodesWithStats {
	var nodes nodesWithStats
	for _, node := range n {
		if condition(&node) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (n nodesWithStats) NodesWithNetworkSchemeChar(netSchemeChar NetworkSchemeChar) nodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return node.NetByte == netSchemeChar
	})
}

func (n nodesWithStats) WorkingNodes() nodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return node.Height > 0
	})
}

func (n nodesWithStats) DownNodes() nodesWithStats {
	return n.Filter(func(node *nodeWithStats) bool {
		return node.Height == -1
	})
}

func (n nodesWithStats) SplitByHeight() map[int]nodesWithStats {
	splitMap := make(map[int]nodesWithStats)
	for _, node := range n {
		splitMap[node.Height] = append(splitMap[node.Height], node)
	}
	return splitMap
}

func (n nodesWithStats) SplitByVersion() map[string]nodesWithStats {
	splitMap := make(map[string]nodesWithStats)
	for _, node := range n {
		splitMap[node.Version] = append(splitMap[node.Version], node)
	}
	return splitMap
}

func (n nodesWithStats) SplitByStateHash() map[string]nodesWithStats {
	splitMap := make(map[string]nodesWithStats)
	for _, node := range n {
		splitMap[node.StateHash] = append(splitMap[node.StateHash], node)
	}
	return splitMap
}
