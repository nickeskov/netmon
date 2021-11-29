package monitor

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NodesStatsScrapper interface {
	ScrapeNodeStats() (NodesWithStats, error)
}

type nodesStatsScrapper struct {
	nodesStatsUrl string
}

func NewNodesStatsScraperHTTP(nodesStatsUrl string) NodesStatsScrapper {
	return nodesStatsScrapper{nodesStatsUrl: nodesStatsUrl}
}

func (s nodesStatsScrapper) ScrapeNodeStats() (NodesWithStats, error) {
	resp, err := http.Get(s.nodesStatsUrl)
	if err != nil {
		return NodesWithStats{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			zap.S().Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return NodesWithStats{},
			errors.Errorf("failed to get nodes statuses from %q, HTTP code(%d) %q",
				s.nodesStatsUrl,
				resp.StatusCode,
				http.StatusText(resp.StatusCode),
			)
	}

	allNodes := NodesWithStats{}
	if err := json.NewDecoder(resp.Body).Decode(&allNodes); err != nil {
		return NodesWithStats{}, err
	}

	zap.S().Debugf("stats successfully received from %q, calculate network error criteria", s.nodesStatsUrl)

	return allNodes, nil
}
