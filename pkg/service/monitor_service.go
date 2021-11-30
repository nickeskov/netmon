package service

import (
	"encoding/json"
	"net/http"

	"github.com/nickeskov/netmon/pkg/monitor"
	"go.uber.org/zap"
)

type NetworkMonitoringService struct {
	monitor *monitor.NetworkMonitor
}

func NewNetworkMonitoringService(monitor *monitor.NetworkMonitor) NetworkMonitoringService {
	return NetworkMonitoringService{
		monitor: monitor,
	}
}

func (s *NetworkMonitoringService) NetworkHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	type statusResponse struct {
		Status bool `json:"status"`
	}
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(statusResponse{Status: s.monitor.NetworkOperatesStable()}); err != nil {
		zap.S().Errorf("failed to marshal status response struct: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// SetMonitorState MUST be protected by auth middleware
func (s *NetworkMonitoringService) SetMonitorState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	type stateChangeRequest struct {
		State string `json:"state"`
	}

	var jsonRequest stateChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&jsonRequest); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		zap.S().Warnf("invalid set monitor state request: %v", err)
		return
	}

	monState, err := monitor.NewNetworkMonitoringStateFromString(jsonRequest.State)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		zap.S().Warnf("invalid set monitor state request: %v", err)
		return
	}

	s.monitor.ChangeState(monState)
	zap.S().Infof("monitor state has been successfully changed, current state is %q", monState)
}
