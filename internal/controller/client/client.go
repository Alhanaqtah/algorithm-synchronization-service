package client

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"sync-algo/internal/lib/logger/sl"
	"sync-algo/internal/lib/response"
	"sync-algo/internal/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Service defines the interface for client operations.
type Service interface {
	AddClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error)
	UpdateClient(ctx context.Context, clientInfo *models.Client) error
	DeleteClient(ctx context.Context, clientID int) error
}

// Handler handles HTTP requests related to clients.
type Handler struct {
	service Service
	log     *slog.Logger
}

// New creates a new instance of Handler.
func New(service Service, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// Register registers the client routes with a router.
func (h *Handler) Register() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", h.addClient)
		r.Put("/{id}", h.updateClient)
		r.Delete("/{id}", h.deleteClient)
	}
}

// @Summary Add a new client
// @Description Add a new client to the system
// @Tags clients
// @Accept json
// @Produce json
// @Param request body models.Client true "Client information"
// @Success 201 {object} models.Client
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /clients/ [post]
func (h *Handler) addClient(w http.ResponseWriter, r *http.Request) {
	const op = "controller.client.addClient"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("creating client...")

	var clientInfo models.Client
	err := render.Decode(r, &clientInfo)
	if err != nil {
		log.Error("failed extract user info from request body", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid credentials"))
		return
	}

	if clientInfo.ClientName == "" {
		log.Error(`client's name is empty`)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid credentials"))
		return
	}

	client, err := h.service.AddClient(r.Context(), &clientInfo)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Err("Internal error"))
		return
	}

	log.Debug("client created successfully")

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, client)
}

// @Summary Update an existing client
// @Description Update an existing client in the system
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Param request body models.Client true "Updated client information"
// @Success 200 {object} models.Client
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /clients/{id} [put]
func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	const op = "controller.client.updateClient"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("updating client...")

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to extract client id from request params", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid client id"))
		return
	}

	var clientInfo models.Client
	err = render.Decode(r, &clientInfo)
	if err != nil {
		log.Error("failed to extract client info from request body", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid request body"))
		return
	}

	if clientInfo.ClientName == "" {
		log.Error("client name is required")
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Client name is required"))
		return
	}

	clientInfo.ID = int64(id)

	err = h.service.UpdateClient(r.Context(), &clientInfo)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Err("Internal error"))
		return
	}

	log.Debug("client updated successfully")

	render.Status(r, http.StatusOK)
}

// @Summary Delete a client
// @Description Delete a client from the system
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /clients/{id} [delete]
func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	const op = "controller.client.deleteClient"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("removing client...")

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("failed to extract client id from request params", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid client id"))
		return
	}

	err = h.service.DeleteClient(r.Context(), id)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Err("Internal error"))
		return
	}

	log.Debug("client removed successfully")

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response.Ok("Client removed successfully"))
}
