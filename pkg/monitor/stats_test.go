package monitor

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodesWithStats_UnmarshalJSON(t *testing.T) {
	const nodesJson = `
{
  "mainnet-aws-fr-4.wavesnodes.com": {
    "netbyte": "W",
    "height": 2878787, 
    "statehash": "801c38b4960d45125e621aa718a68aa6db74bd25c09c9373c17daa49cac04cfe", 
    "statehash_height": 2878785, 
    "version": "Waves v1.3.10-12-g2fb491a"
  },
  "stagenet-htz-nbg1-2.wavesnodes.com": {
    "netbyte": "S",
    "height": 1098732, 
    "statehash": "b8aec310cdb50d874261c0b8aa9f5358e948eef1c4e5102125cec959bd342afd", 
    "statehash_height": 1098730, 
    "version": "Waves v1.4.1"
  },
  "testnet-htz-nbg1-2.wavesnodes.com": {
    "netbyte": "T",
    "height": 1813844, 
    "statehash": "5d11b19998ff03f4e9ab2fb5d55588050d3973806542cf3feeac0964efbec531", 
    "statehash_height": 1813842, 
    "version": "Waves v1.3.9-1-gca8c26b"
  }
}
`
	expected := nodesWithStats{
		nodeWithStats{
			NodeDomain: "mainnet-aws-fr-4.wavesnodes.com",
			nodeStats: nodeStats{
				NetByte:         MainNetSchemeChar,
				Height:          2878787,
				StateHash:       "801c38b4960d45125e621aa718a68aa6db74bd25c09c9373c17daa49cac04cfe",
				StateHashHeight: 2878785,
				Version:         "Waves v1.3.10-12-g2fb491a",
			},
		},
		nodeWithStats{
			NodeDomain: "stagenet-htz-nbg1-2.wavesnodes.com",
			nodeStats: nodeStats{
				NetByte:         StageNetSchemeChar,
				Height:          1098732,
				StateHash:       "b8aec310cdb50d874261c0b8aa9f5358e948eef1c4e5102125cec959bd342afd",
				StateHashHeight: 1098730,
				Version:         "Waves v1.4.1",
			},
		},
		nodeWithStats{
			NodeDomain: "testnet-htz-nbg1-2.wavesnodes.com",
			nodeStats: nodeStats{
				NetByte:         TestNetSchemeChar,
				Height:          1813844,
				StateHash:       "5d11b19998ff03f4e9ab2fb5d55588050d3973806542cf3feeac0964efbec531",
				StateHashHeight: 1813842,
				Version:         "Waves v1.3.9-1-gca8c26b",
			},
		},
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].NodeDomain < expected[j].NodeDomain
	})

	actual := nodesWithStats{}
	err := json.Unmarshal([]byte(nodesJson), &actual)
	require.NoError(t, err)
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].NodeDomain < actual[j].NodeDomain
	})

	require.Equal(t, expected, actual)
}

func TestNodesWithStats_DownNodes(t *testing.T) {
	data := nodesWithStats{
		nodeWithStats{NodeDomain: "11", nodeStats: nodeStats{Height: -1}},
		nodeWithStats{NodeDomain: "22", nodeStats: nodeStats{Height: -1}},
		nodeWithStats{NodeDomain: "33", nodeStats: nodeStats{Height: 10}},
		nodeWithStats{NodeDomain: "44", nodeStats: nodeStats{Height: 12}},
	}
	expected := nodesWithStats{
		nodeWithStats{NodeDomain: "11", nodeStats: nodeStats{Height: -1}},
		nodeWithStats{NodeDomain: "22", nodeStats: nodeStats{Height: -1}},
	}
	require.Equal(t, expected, data.DownNodes())
}

func TestNodesWithStats_Filter(t *testing.T) {
	data := nodesWithStats{
		nodeWithStats{NodeDomain: "11", nodeStats: nodeStats{Height: -1}},
		nodeWithStats{NodeDomain: "22", nodeStats: nodeStats{Height: -1}},
		nodeWithStats{NodeDomain: "2255", nodeStats: nodeStats{Height: 8}},
		nodeWithStats{NodeDomain: "33", nodeStats: nodeStats{Height: 10}},
		nodeWithStats{NodeDomain: "44", nodeStats: nodeStats{Height: 12}},
	}
	expected := nodesWithStats{
		nodeWithStats{NodeDomain: "33", nodeStats: nodeStats{Height: 10}},
		nodeWithStats{NodeDomain: "44", nodeStats: nodeStats{Height: 12}},
	}

	actual := data.Filter(func(node *nodeWithStats) bool {
		return node.Height != -1 && len(node.NodeDomain) == 2
	})
	require.Equal(t, expected, actual)
}
