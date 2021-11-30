package monitor

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NodesStatsScrapper interface {
	ScrapeNodeStats() (nodesWithStats, error)
}

type nodesStatsScrapper struct {
	nodesStatsUrl string
}

func NewNodesStatsScraperHTTP(nodesStatsUrl string) NodesStatsScrapper {
	return nodesStatsScrapper{nodesStatsUrl: nodesStatsUrl}
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

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
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
	if err := json.NewDecoder(resp.Body).Decode(&allNodes); err != nil {
		return nodesWithStats{}, err
	}

	zap.S().Debugf("stats successfully received from %q", s.nodesStatsUrl)

	return allNodes, nil
}
