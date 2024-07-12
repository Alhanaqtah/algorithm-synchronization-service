package algorithm

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_service "sync-algo/internal/controller/algorithm/mock"
	"sync-algo/internal/lib/logger/handlers/slogdiscard"
	models "sync-algo/internal/models"

	"github.com/go-chi/chi"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_updateAlgorithmStatus(t *testing.T) {
	type mockBehavior func(s *mock_service.MockService, algoStatuses *models.AlgoStatuses)

	tt := []struct {
		name                 string
		inputBody            string
		inputAlgoStatuses    models.AlgoStatuses
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         mockBehavior
	}{
		{
			name:                 "Update algorithm statuses successfully",
			inputBody:            `{"client_id": 1, "vwap": true, "twap": false, "hft": true}`,
			inputAlgoStatuses:    models.AlgoStatuses{ClientID: 1, VWAP: boolPtr(true), TWAP: boolPtr(false), HFT: boolPtr(true)},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"client_id": 1, "vwap": true, "twap": false, "hft": true}`,
			mockBehavior: func(s *mock_service.MockService, algoStatuses *models.AlgoStatuses) {
				s.EXPECT().UpdateStatuses(gomock.Any(), algoStatuses).Return(algoStatuses, nil)
			},
		},
		{
			name:                 "Invalid JSON body",
			inputBody:            `invalid JSON`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         nil,
		},
		{
			name:                 "Missing ClientID",
			inputBody:            `{"vwap": true}`,
			inputAlgoStatuses:    models.AlgoStatuses{VWAP: boolPtr(true)},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid data"}`,
			mockBehavior:         nil,
		},
		{
			name:                 "Service error",
			inputBody:            `{"client_id": 1, "vwap": true, "twap": false, "hft": true}`,
			inputAlgoStatuses:    models.AlgoStatuses{ClientID: 1, VWAP: boolPtr(true), TWAP: boolPtr(false), HFT: boolPtr(true)},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"status":"Error","error":"Internal error"}`,
			mockBehavior: func(s *mock_service.MockService, algoStatuses *models.AlgoStatuses) {
				s.EXPECT().UpdateStatuses(gomock.Any(), algoStatuses).Return(nil, errors.New("internal service error"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_service.NewMockService(ctrl)
			if tc.mockBehavior != nil {
				tc.mockBehavior(service, &tc.inputAlgoStatuses)
			}

			log := slogdiscard.NewDiscardLogger()
			handler := New(service, log)

			// Test server
			r := chi.NewRouter()
			r.Patch("/algorithms/statuses", handler.updateAlgorithmStatus)

			// Test request
			req := httptest.NewRequest("PATCH", "/algorithms/statuses", bytes.NewBufferString(tc.inputBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
