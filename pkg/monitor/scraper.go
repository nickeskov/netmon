package monitor

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const DefaultNodeStatsPollResponseSize = 128 * 1024

type NodesStatsScrapper interface {
	ScrapeNodeStats() (nodesWithStats, error)
}

type nodesStatsScrapper struct {
	nodesStatsUrl   string
	maxResponseSize int64
}

func NewNodesStatsScraperHTTP(nodesStatsUrl string, maxResponseSize int64) NodesStatsScrapper {
	return nodesStatsScrapper{nodesStatsUrl: nodesStatsUrl, maxResponseSize: maxResponseSize}
}

func (s nodesStatsScrapper) ScrapeNodeStats() (nodesWithStats, error) {
	resp, err := http.Get(s.nodesStatsUrl)
	if err != nil {
		return nodesWithStats{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			zap.S().Errorf("failed to close response body: %v", err)
		}
	}()
	responseBody := io.LimitReader(resp.Body, s.maxResponseSize)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(responseBody)
		if err != nil {
			return nil, err
		}
		zap.S().Errorf("stats sevice returned response with HTTP code %d %q, response is %q",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			string(body),
		)
		return nodesWithStats{},
			errors.Errorf("failed to get nodes statuses from %q, HTTP code(%d) %q",
				s.nodesStatsUrl,
				resp.StatusCode,
				http.StatusText(resp.StatusCode),
			)
	}

	allNodes := nodesWithStats{}
	if err := json.NewDecoder(responseBody).Decode(&allNodes); err != nil {
		return nodesWithStats{}, err
	}

	zap.S().Debugf("stats successfully received from %q", s.nodesStatsUrl)

	return allNodes, nil
}
