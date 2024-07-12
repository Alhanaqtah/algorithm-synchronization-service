package algorithm

import (
	"context"
	"fmt"
	"log/slog"
	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/models"
)

type Storage interface {
	UpdateStatuses(ctx context.Context, fields []string, values []interface{}) (*models.AlgoStatuses, error)
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

func (s *Service) UpdateStatuses(ctx context.Context, algoStatuses *models.AlgoStatuses) (*models.AlgoStatuses, error) {
	const op = "service.algorithm.UpdateStatuses"

	log := s.log.With(slog.String("op", op))

	var fields []string
	var values []interface{}
	order := 1

	if algoStatuses.HFT != nil {
		fields = append(fields, fmt.Sprintf("hft = $%d", order))
		values = append(values, fmt.Sprint(*algoStatuses.HFT))
		order++
	}
	if algoStatuses.TWAP != nil {
		fields = append(fields, fmt.Sprintf("twap = $%d", order))
		values = append(values, fmt.Sprint(*algoStatuses.TWAP))
		order++
	}
	if algoStatuses.VWAP != nil {
		fields = append(fields, fmt.Sprintf("vwap = $%d", order))
		values = append(values, fmt.Sprint(*algoStatuses.VWAP))
		order++
	}

	values = append(values, algoStatuses.ClientID)

	log.Debug("Payload", "fields", fields, "values", values)

	updatedAlgoStatuses, err := s.storage.UpdateStatuses(ctx, fields, values)
	if err != nil {
		log.Error("failed to update algorithms", sl.Error(err))
		return nil, err
	}

	return updatedAlgoStatuses, nil
}
