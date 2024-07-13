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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockService(ctrl)
	logger := slogdiscard.NewDiscardLogger()
	handler := New(mockService, logger)

	r := chi.NewRouter()
	r.Post("/clients", handler.addClient)

	tt := []struct {
		name                 string
		inputBody            string
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:                 "Create client with valid data",
			inputBody:            `{"client_name": "clientName"}`,
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"client_name":"clientName","spawned_at":"0001-01-01T00:00:00Z","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`,
			mockBehavior: func() {
				mockService.EXPECT().AddClient(gomock.Any(), gomock.Any()).Return(&models.Client{
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
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         func() {},
		},
		{
			name:                 "Empty client name",
			inputBody:            `{"client_name": ""}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid credentials"}`,
			mockBehavior:         func() {},
		},
		{
			name:                 "Service error",
			inputBody:            `{"client_name": "clientName"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"status":"Error","error":"Internal error"}`,
			mockBehavior: func() {
				mockService.EXPECT().AddClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			reqBody := bytes.NewBufferString(tc.inputBody)
			req := httptest.NewRequest("POST", "/clients", reqBody)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_deleteClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_service.NewMockService(ctrl)
	logger := slogdiscard.NewDiscardLogger()
	handler := New(mockService, logger)

	r := chi.NewRouter()
	r.Delete("/clients/{id}", handler.deleteClient)

	tt := []struct {
		name                 string
		url                  string
		expectedStatusCode   int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:                 "Delete client successfully",
			url:                  "/clients/1",
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"status":"OK","message":"Client removed successfully"}`,
			mockBehavior: func() {
				mockService.EXPECT().DeleteClient(gomock.Any(), 1).Return(nil)
			},
		},
		{
			name:                 "Invalid client ID in URL",
			url:                  "/clients/invalid",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"status":"Error","error":"Invalid client id"}`,
			mockBehavior:         func() {},
		},
		{
			name:                 "Service error",
			url:                  "/clients/1",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"status":"Error","error":"Internal error"}`,
			mockBehavior: func() {
				mockService.EXPECT().DeleteClient(gomock.Any(), 1).Return(errors.New("internal service error"))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("DELETE", tc.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}
