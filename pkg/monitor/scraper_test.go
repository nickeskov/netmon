package monitor

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodesStatsScrapper_ScrapeNodeStats(t *testing.T) {
	const nodesJson = `
{
  "mainnet-aws-fr-4.wavesnodes.com": {
    "height": 2878787, 
    "statehash": "801c38b4960d45125e621aa718a68aa6db74bd25c09c9373c17daa49cac04cfe", 
    "statehash_height": "2878785", 
    "version": "Waves v1.3.10-12-g2fb491a"
  },
  "stagenet-htz-nbg1-2.wavesnodes.com": {
    "height": 1098732, 
    "statehash": "b8aec310cdb50d874261c0b8aa9f5358e948eef1c4e5102125cec959bd342afd", 
    "statehash_height": "1098730", 
    "version": "Waves v1.4.1"
  },
  "testnet-htz-nbg1-2.wavesnodes.com": {
    "height": 1813844, 
    "statehash": "5d11b19998ff03f4e9ab2fb5d55588050d3973806542cf3feeac0964efbec531", 
    "statehash_height": "1813842", 
    "version": "Waves v1.3.9-1-gca8c26b"
  }
}
`
	expected := nodesWithStats{
		nodeWithStats{
			NodeDomain: "mainnet-aws-fr-4.wavesnodes.com",
			nodeStats: nodeStats{
				Height:          2878787,
				StateHash:       "801c38b4960d45125e621aa718a68aa6db74bd25c09c9373c17daa49cac04cfe",
				StateHashHeight: "2878785",
				Version:         "Waves v1.3.10-12-g2fb491a",
			},
		},
		nodeWithStats{
			NodeDomain: "stagenet-htz-nbg1-2.wavesnodes.com",
			nodeStats: nodeStats{
				Height:          1098732,
				StateHash:       "b8aec310cdb50d874261c0b8aa9f5358e948eef1c4e5102125cec959bd342afd",
				StateHashHeight: "1098730",
				Version:         "Waves v1.4.1",
			},
		},
		nodeWithStats{
			NodeDomain: "testnet-htz-nbg1-2.wavesnodes.com",
			nodeStats: nodeStats{
				Height:          1813844,
				StateHash:       "5d11b19998ff03f4e9ab2fb5d55588050d3973806542cf3feeac0964efbec531",
				StateHashHeight: "1813842",
				Version:         "Waves v1.3.9-1-gca8c26b",
			},
		},
	}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].NodeDomain < expected[j].NodeDomain
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		trimmedJson := strings.ReplaceAll(nodesJson, "\t", "")
		_, err := writer.Write([]byte(trimmedJson))
		require.NoError(t, err)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	scraper := NewNodesStatsScraperHTTP(srv.URL)
	actual, err := scraper.ScrapeNodeStats()
	require.NoError(t, err)
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].NodeDomain < actual[j].NodeDomain
	})

	require.Equal(t, expected, actual)
}
