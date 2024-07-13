package algorithm

import (
	"context"
	"log/slog"
	"net/http"
	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/lib/response"
	"sync-algo/internal/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

var (
	emptyValue = 0
)

// Service defines the interface for managing algorithm statuses.
//
//go:generate mockgen -source=algorithm.go -destination=mock/mock.go -package=algorithm
type Service interface {
	UpdateStatuses(ctx context.Context, algoStatuses *models.AlgoStatuses) (*models.AlgoStatuses, error)
}

type Handler struct {
	service Service
	log     *slog.Logger
}

func New(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// Register registers the API routes
func (h *Handler) Register() func(r chi.Router) {
	return func(r chi.Router) {
		r.Patch("/", h.updateAlgorithmStatus)
	}
}

// @Summary Update algorithm statuses
// @Description Update the status of algorithms based on the provided data.
// @Tags algorithms
// @Accept json
// @Produce json
// @Param body body models.AlgoStatuses true "Algorithm statuses to update"
// @Success 200 {object} models.AlgoStatuses "Updated algorithm statuses"
// @Failure 400 {object} response.Response "Invalid credentials or data"
// @Failure 500 {object} response.Response "Internal error"
// @Router /algorithm/ [patch]
func (h *Handler) updateAlgorithmStatus(w http.ResponseWriter, r *http.Request) {
	const op = "controller.algorithm.updateAlgorithmStatus"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("updating algorithms status...")

	// Decode the request body
	var algoStatuses models.AlgoStatuses
	err := render.Decode(r, &algoStatuses)
	if err != nil {
		log.Error("failed to extract request body", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid credentials"))
		return
	}

	// Validate the received data
	if algoStatuses.ClientID == emptyValue || !(algoStatuses.VWAP != nil || algoStatuses.TWAP != nil || algoStatuses.HFT != nil) {
		log.Error("invalid data provided")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid data"))
		return
	}

	// Call the service to update the algorithm statuses
	updatedStatuses, err := h.service.UpdateStatuses(r.Context(), &algoStatuses)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Err("Internal error"))
		return
	}

	log.Debug("algorithms statuses updated successfully")

	render.Status(r, http.StatusOK)
	render.JSON(w, r, updatedStatuses)
}
