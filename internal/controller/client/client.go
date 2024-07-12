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

//go:generate mockgen -source=client.go -destination=mock/mock.go
type Service interface {
	AddClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error)
	UpdateClient(ctx context.Context, clientInfo *models.Client) (*models.Client, error)
	DeleteClient(ctx context.Context, clientID int) error
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

func (h *Handler) Register() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", h.addClient)
		r.Patch("/{id}", h.updateClient)
		r.Delete("/{id}", h.deleteClient)
	}
}

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
		log.Error(`client's 'name' is empty`)
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

func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	const op = "controller.client.updateClient"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("updating client...")

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("faile to extract client id from request params", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid client id"))
		return
	}

	var clientInfo models.Client
	err = render.Decode(r, &clientInfo)
	if err != nil {
		log.Error("failed extract user info from request body", sl.Error(err))
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.Err("Invalid credentials"))
		return
	}

	clientInfo.ID = int64(id)

	client, err := h.service.UpdateClient(r.Context(), &clientInfo)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.Err("Internal error"))
		return
	}

	log.Debug("client updated successfully")

	render.Status(r, http.StatusOK)
	render.JSON(w, r, client)
}

func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	const op = "controller.client.deleteClient"

	log := h.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	log.Debug("removing client...")

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Error("faile to extract client id from request params", sl.Error(err))
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
	render.JSON(w, r, response.Ok("Client removed succesfully"))
}
