package client

import (
	"bytes"
	"errors"
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
		{
			name:                 "Invalid JSON body",
			inputBody:            `{"client_name":}`,
			inputClient:          models.Client{},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         func(s *mock_service.MockService, clientInfo *models.Client) {},
		},
		{
			name:                 "Empty client name",
			inputBody:            `{"client_name": ""}`,
			inputClient:          models.Client{},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         func(s *mock_service.MockService, clientInfo *models.Client) {},
		},
		{
			name:                 "Service error",
			inputBody:            `{"client_name": "clientName"}`,
			inputClient:          models.Client{ClientName: "clientName"},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"status":"Error","error":"Internal error"}`,
			mockBehavior: func(s *mock_service.MockService, clientInfo *models.Client) {
				s.EXPECT().AddClient(gomock.Any(), clientInfo).Return(nil, errors.New(""))
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
			r.Post("/clients", handler.addClient)

			// Test request
			reqBody := bytes.NewBufferString(tc.inputBody)
			req := httptest.NewRequest("POST", "/clients", reqBody)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_updateClient(t *testing.T) {
	type mockBehavior func(s *mock_service.MockService, clientInfo *models.Client)

	tt := []struct {
		name                 string
		url                  string
		inputBody            string
		inputClient          models.Client
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         mockBehavior
	}{
		{
			name:                 "Update client name",
			url:                  "/clients/1",
			inputBody:            `{"client_name": "NewName"}`,
			inputClient:          models.Client{ID: 1, ClientName: "NewName"},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"client_name":"NewName","spawned_at":"0001-01-01T00:00:00Z","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
			mockBehavior: func(s *mock_service.MockService, clientInfo *models.Client) {
				s.EXPECT().UpdateClient(gomock.Any(), clientInfo).Return(&models.Client{
					ID:         clientInfo.ID,
					ClientName: "NewName",
					SpawnedAt:  time.Time{},
					CreatedAt:  time.Time{},
					UpdatedAt:  time.Time{},
				}, nil)
			},
		},
		{
			name:                 "Invalid client ID in URL",
			url:                  "/clients/invalid",
			inputBody:            `{"client_name": "NewName"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid client id"}`,
			mockBehavior:         nil,
		},
		{
			name:                 "Invalid JSON body",
			url:                  "/clients/1",
			inputBody:            `invalid JSON`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_service.NewMockService(ctrl)

			// Установка mockBehavior только если он определен
			if tc.mockBehavior != nil {
				tc.mockBehavior(service, &tc.inputClient)
			}

			log := slogdiscard.NewDiscardLogger()
			handler := New(service, log)

			// Test server
			r := chi.NewRouter()
			r.Patch("/clients/{id}", handler.updateClient)

			// Test request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", tc.url, bytes.NewBufferString(tc.inputBody))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_deleteClient(t *testing.T) {
	type mockBehavior func(s *mock_service.MockService, id int)

	tt := []struct {
		name                 string
		url                  string
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         mockBehavior
	}{
		{
			name:                 "Delete client successfully",
			url:                  "/clients/1",
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"status":"OK","message":"Client removed succesfully"}`,
			mockBehavior: func(s *mock_service.MockService, id int) {
				s.EXPECT().DeleteClient(gomock.Any(), id).Return(nil)
			},
		},
		{
			name:                 "Invalid client ID in URL",
			url:                  "/clients/invalid",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid client id"}`,
			mockBehavior:         func(s *mock_service.MockService, id int) {},
		},
		{
			name:                 "Service error",
			url:                  "/clients/1",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"status":"Error","error":"Internal error"}`,
			mockBehavior: func(s *mock_service.MockService, id int) {
				s.EXPECT().DeleteClient(gomock.Any(), id).Return(errors.New("internal service error"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_service.NewMockService(ctrl)
			tc.mockBehavior(service, 1) // Using 1 as a placeholder for id

			log := slogdiscard.NewDiscardLogger()
			handler := New(service, log)

			// Test server
			r := chi.NewRouter()
			r.Delete("/clients/{id}", handler.deleteClient)

			// Test request
			req := httptest.NewRequest("DELETE", tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}
