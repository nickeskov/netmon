package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nickeskov/netmon/pkg/monitor"
	"github.com/stretchr/testify/require"
)

func TestNetworkMonitoringService_NetworkHealth(t *testing.T) {
	now := time.Now()
	tests := []struct {
		testName       string
		httpMethod     string
		httpStatusCode int
		mockCallTimes  int
		statusInfo     monitor.NetworkStatusInfo
	}{
		{
			testName:       "PositiveScenario",
			httpMethod:     http.MethodGet,
			httpStatusCode: http.StatusOK,
			mockCallTimes:  1,
			statusInfo: monitor.NetworkStatusInfo{
				Updated: now,
				Network: monitor.MainNetSchemeChar,
				Status:  true,
				Height:  12345,
			},
		},
		{
			testName:       "HTTPMethodPost",
			httpMethod:     http.MethodPost,
			httpStatusCode: http.StatusMethodNotAllowed,
			mockCallTimes:  0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMonitor := monitor.NewMockMonitor(ctrl)

			mockMonitor.EXPECT().NetworkStatusInfo().Times(tc.mockCallTimes).Return(tc.statusInfo)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.httpMethod, "/health", nil)

			netMon := NewNetworkMonitoringService(mockMonitor)
			netMon.NetworkHealth(w, r)
			defer func() {
				require.NoError(t, w.Result().Body.Close())
			}()

			if tc.httpStatusCode == http.StatusOK {
				require.Equal(t, "application/json", w.Header().Get("content-type"))

				data, err := io.ReadAll(w.Result().Body)
				require.NoError(t, err)

				var info monitor.NetworkStatusInfo
				err = json.Unmarshal(data, &info)
				require.NoError(t, err)

				require.True(t, tc.statusInfo.Updated.Equal(info.Updated))

				info.Updated = tc.statusInfo.Updated
				require.EqualValues(t, tc.statusInfo, info)
			} else {
				require.Equal(t, tc.httpStatusCode, w.Result().StatusCode)
				require.Equal(t, fmt.Sprintf("%d %s", tc.httpStatusCode, http.StatusText(tc.httpStatusCode)), w.Result().Status)
			}
		})
	}
}

func TestNetworkMonitoringService_SetMonitorState(t *testing.T) {
	tests := []struct {
		testName        string
		httpMethod      string
		httpRequestBody io.ReadCloser
		httpStatusCode  int
		mockCallTimes   int
		stateUpdate     monitor.NetworkMonitoringState
		prevState       monitor.NetworkMonitoringState
	}{
		{
			testName:       "HTTPMethodGet",
			httpMethod:     http.MethodGet,
			httpStatusCode: http.StatusMethodNotAllowed,
			mockCallTimes:  0,
		},
		{
			testName:        "InvalidJSON",
			httpMethod:      http.MethodPost,
			httpStatusCode:  http.StatusBadRequest,
			httpRequestBody: io.NopCloser(strings.NewReader(`{"state":"ac`)),
			mockCallTimes:   0,
		},
		{
			testName:        "InvalidState",
			httpMethod:      http.MethodPost,
			httpStatusCode:  http.StatusBadRequest,
			httpRequestBody: io.NopCloser(strings.NewReader(`{"state":"blah-blah-blah"}`)),
			mockCallTimes:   0,
		},
		{
			testName:        "PositiveScenarioActiveToActive",
			httpMethod:      http.MethodPost,
			httpRequestBody: io.NopCloser(strings.NewReader(`{"state":"active"}`)),
			httpStatusCode:  http.StatusOK,
			mockCallTimes:   1,
			stateUpdate:     monitor.StateActive,
			prevState:       monitor.StateActive,
		},
		{
			testName:        "PositiveScenarioDegradedToOperatesStable",
			httpMethod:      http.MethodPost,
			httpRequestBody: io.NopCloser(strings.NewReader(`{"state":"frozen_operates_stable"}`)),
			httpStatusCode:  http.StatusOK,
			mockCallTimes:   1,
			stateUpdate:     monitor.StateFrozenNetworkOperatesStable,
			prevState:       monitor.StateFrozenNetworkDegraded,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMonitor := monitor.NewMockMonitor(ctrl)
			if tc.httpStatusCode == http.StatusOK {
				stateCall := mockMonitor.EXPECT().State().Times(1).Return(tc.prevState)
				mockMonitor.EXPECT().ChangeState(tc.stateUpdate).After(stateCall).Times(1)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.httpMethod, "/state", tc.httpRequestBody)

			netMon := NewNetworkMonitoringService(mockMonitor)
			netMon.SetMonitorState(w, r)
			defer func() {
				require.NoError(t, w.Result().Body.Close())
			}()

			if tc.httpStatusCode == http.StatusOK {
				data, err := io.ReadAll(w.Result().Body)
				require.NoError(t, err)
				require.Empty(t, data)
			} else {
				require.Equal(t, fmt.Sprintf("%d %s", tc.httpStatusCode, http.StatusText(tc.httpStatusCode)), w.Result().Status)
				require.Equal(t, tc.httpStatusCode, w.Result().StatusCode)
			}
		})
	}

}
