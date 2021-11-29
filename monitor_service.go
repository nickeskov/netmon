package main

import (
	"encoding/json"
	"net/http"
)

type NetworkMonitoringService struct {
	monitor *NetworkMonitor
}

func NewNetworkMonitoringService(monitor *NetworkMonitor) NetworkMonitoringService {
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
	if err := json.NewEncoder(w).Encode(statusResponse{Status: s.monitor.NetworkOperatesStable()}); err != nil {
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	monState, err := NewNetworkMonitoringStateFromString(jsonRequest.State)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	s.monitor.ChangeState(monState)
}
