package service

import (
	"encoding/json"
	"net/http"

	"github.com/nickeskov/netmon/pkg/monitor"
	"go.uber.org/zap"
)

type NetworkMonitoringService struct {
	monitor monitor.Monitor
}

func NewNetworkMonitoringService(monitor monitor.Monitor) NetworkMonitoringService {
	return NetworkMonitoringService{
		monitor: monitor,
	}
}

func (s *NetworkMonitoringService) NetworkHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(s.monitor.NetworkStatusInfo()); err != nil {
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
		zap.S().Warnf("invalid set monitor state request, failed to parse JSON: %v", err)
		return
	}
	prevMonState := s.monitor.State()
	s.monitor.ChangeState(monState)
	if prevMonState != monState {
		zap.S().Infof("monitor state has been successfully changed from %q to %q",
			prevMonState.String(), monState.String(),
		)
	} else {
		zap.S().Infof("monitor state hasn't been changed, current state is %q", prevMonState.String())
	}
}
