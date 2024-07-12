package client

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mock_service "sync-algo/internal/controller/client/mock"
	"sync-algo/internal/lib/logger/handlers/slogdiscard"
	"sync-algo/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_addClient(t *testing.T) {
	type mockBehavior func(s *mock_service.MockService, clientInfo *models.Client)

	tt := []struct {
		name                 string
		inputBody            string
		inputClient          models.Client
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         mockBehavior
	}{
		{
			name:                 "Create client only with name",
			inputBody:            `{"client_name": "clientName"}`,
			inputClient:          models.Client{ClientName: "clientName"},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"client_name":"clientName","spawned_at":"0001-01-01T00:00:00Z","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
			mockBehavior: func(s *mock_service.MockService, clientInfo *models.Client) {
				s.EXPECT().AddClient(gomock.Any(), clientInfo).Return(&models.Client{
					ID:         1,
					ClientName: "clientName",
					SpawnedAt:  time.Time{},
					CreatedAt:  time.Time{},
					UpdatedAt:  time.Time{},
				}, nil)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_service.NewMockService(ctrl)
			tc.mockBehavior(service, &tc.inputClient)

			log := slogdiscard.NewDiscardLogger()
			handler := New(service, log)

			// Test server
			r := chi.NewRouter()
			r.Post("/", handler.addClient)

			// Test request
			reqBody := bytes.NewBufferString(tc.inputBody)
			req := httptest.NewRequest("POST", "/", reqBody)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}
