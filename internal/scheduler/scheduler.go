package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/models"
)

// Deployer defines the interface for managing Pods
//
//go:generate mockgen -source=scheduler.go -destination=../deployer/mock/mock.go -package=mock
type Deployer interface {
	CreatePod(name string) error
	DeletePod(name string) error
	GetPodList() ([]string, error)
}

// Storage defines the interface for fetching algorithm statuses
type Storage interface {
	FetchCurrentStatuses(ctx context.Context) ([]models.AlgoStatuses, error)
}

// Scheduler is responsible for synchronizing the state of algorithms
type Scheduler struct {
	storage          Storage
	log              *slog.Logger
	previousStatuses map[int64]models.AlgoStatuses
}

// New creates a new Scheduler instance
func New(log *slog.Logger, storage Storage) *Scheduler {
	return &Scheduler{
		storage:          storage,
		log:              log,
		previousStatuses: make(map[int64]models.AlgoStatuses),
	}
}

// Start begins the scheduling process
func (s *Scheduler) Start(ctx context.Context, deployer Deployer) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncAlgorithmStatus(ctx, deployer)
		case <-ctx.Done():
			return
		}
	}
}

// syncAlgorithmStatus synchronizes the state of algorithms with the pod state
func (s *Scheduler) syncAlgorithmStatus(ctx context.Context, deployer Deployer) {
	const op = "scheduler.syncAlgorithmStatus"

	log := s.log.With(slog.String("op", op))

	currentStatuses, err := s.storage.FetchCurrentStatuses(ctx)
	if err != nil {
		s.log.Error("error fetching algorithm statuses:", sl.Error(err))
		return
	}

	for _, currentStatus := range currentStatuses {
		previousStatus, exists := s.previousStatuses[int64(currentStatus.ClientID)]

		// Create or update the pod if status has changed to enabled
		if (*currentStatus.VWAP || *currentStatus.TWAP || *currentStatus.HFT) &&
			(!exists || !*previousStatus.VWAP && !*previousStatus.TWAP && !*previousStatus.HFT) {
			podName := fmt.Sprintf("client-%d-pod", currentStatus.ClientID)
			if err := deployer.CreatePod(podName); err != nil {
				log.Error("error creating pod", sl.Error(err))
			}
		}

		// Delete the pod if status has changed to disabled
		if !*currentStatus.VWAP && !*currentStatus.TWAP && !*currentStatus.HFT &&
			exists && (*previousStatus.VWAP || *previousStatus.TWAP || *previousStatus.HFT) {
			podName := fmt.Sprintf("client-%d-pod", currentStatus.ClientID)
			if err := deployer.DeletePod(podName); err != nil {
				log.Error("error deleting pod:", sl.Error(err))
			}
		}

		// Update the previous status
		s.previousStatuses[int64(currentStatus.ClientID)] = currentStatus
	}
}
