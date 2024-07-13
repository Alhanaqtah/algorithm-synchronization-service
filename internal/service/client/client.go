package client

import (
	"context"
	"log/slog"

	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/models"
)

const emptyValue = 0

//go:generate mockgen -source=client.go -destination=mock/mock.go -package=mock_storage
type Storage interface {
	CreateClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error)
	UpdateClient(ctx context.Context, clientInfo *models.Client) error
	RemoveClient(ctx context.Context, id int) error
}

type Service struct {
	storage Storage
	log     *slog.Logger
}

func New(storage Storage, log *slog.Logger) *Service {
	return &Service{
		storage: storage,
		log:     log,
	}
}
func (s *Service) AddClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error) {
	const op = "service.client.AddClient"

	log := s.log.With(slog.String("op", op))

	client, err := s.storage.CreateClient(ctx, clientInfo)
	if err != nil {
		log.Error("failed to save client", sl.Error(err))
		return nil, err
	}

	return client, nil
}

func (s *Service) UpdateClient(ctx context.Context, clientInfo *models.Client) error {
	const op = "service.client.UpdateClient"

	log := s.log.With(slog.String("op", op))

	err := s.storage.UpdateClient(ctx, clientInfo)
	if err != nil {
		log.Error("failed to update client", sl.Error(err))
		return err
	}

	return nil
}

func (s *Service) DeleteClient(ctx context.Context, clientID int) error {
	const op = "service.client.DeleteClient"

	log := s.log.With(slog.String("op", op))

	err := s.storage.RemoveClient(ctx, clientID)
	if err != nil {
		log.Error("failed to remove client", sl.Error(err))
		return err
	}

	return nil
}
