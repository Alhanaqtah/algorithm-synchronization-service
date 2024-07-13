package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"sync-algo/internal/lib/logger/handlers/slogdiscard"
	"sync-algo/internal/models"
	mock_storage "sync-algo/internal/service/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_AddClient(t *testing.T) {
	type mockBehavior func(s *mock_storage.MockStorage, clientInfo *models.Client)

	tt := []struct {
		name           string
		inputClient    models.Client
		mockBehavior   mockBehavior
		expectedClient *models.Client
		expectedError  error
	}{
		{
			name:        "Successful client creation",
			inputClient: models.Client{ClientName: "clientName"},
			mockBehavior: func(s *mock_storage.MockStorage, clientInfo *models.Client) {
				s.EXPECT().CreateClient(gomock.Any(), clientInfo).Return(&models.Client{
					ID:         1,
					ClientName: "clientName",
					SpawnedAt:  time.Time{},
					CreatedAt:  time.Time{},
					UpdatedAt:  time.Time{},
				}, nil)
			},
			expectedClient: &models.Client{
				ID:         1,
				ClientName: "clientName",
				SpawnedAt:  time.Time{},
				CreatedAt:  time.Time{},
				UpdatedAt:  time.Time{},
			},
			expectedError: nil,
		},
		{
			name:        "Storage error",
			inputClient: models.Client{ClientName: "clientName"},
			mockBehavior: func(s *mock_storage.MockStorage, clientInfo *models.Client) {
				s.EXPECT().CreateClient(gomock.Any(), clientInfo).Return(nil, errors.New("internal storage error"))
			},
			expectedClient: nil,
			expectedError:  errors.New("internal storage error"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mock_storage.NewMockStorage(ctrl)
			tc.mockBehavior(storage, &tc.inputClient)

			log := slogdiscard.NewDiscardLogger()
			service := New(storage, log)

			// Test method
			client, err := service.AddClient(context.Background(), &tc.inputClient)

			// Assert results
			assert.Equal(t, tc.expectedClient, client)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestService_DeleteClient(t *testing.T) {
	type mockBehavior func(s *mock_storage.MockStorage, clientID int)

	tests := []struct {
		name          string
		clientID      int
		mockBehavior  mockBehavior
		expectedError error
	}{
		{
			name:     "Successful client deletion",
			clientID: 1,
			mockBehavior: func(s *mock_storage.MockStorage, clientID int) {
				s.EXPECT().RemoveClient(gomock.Any(), clientID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "Storage error",
			clientID: 1,
			mockBehavior: func(s *mock_storage.MockStorage, clientID int) {
				s.EXPECT().RemoveClient(gomock.Any(), clientID).Return(errors.New("internal storage error"))
			},
			expectedError: errors.New("internal storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init deps
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mock_storage.NewMockStorage(ctrl)
			tt.mockBehavior(storage, tt.clientID)

			log := slogdiscard.NewDiscardLogger()
			service := New(storage, log)

			// Test method
			err := service.DeleteClient(context.Background(), tt.clientID)

			// Assert results
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
