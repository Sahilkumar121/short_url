package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Sahilkumar121/short_url/internal/auth"
	"github.com/Sahilkumar121/short_url/internal/config"
	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/middleware"
	"github.com/Sahilkumar121/short_url/internal/model"
	"github.com/Sahilkumar121/short_url/internal/service"
	"github.com/jackc/pgx/v5"
)

type UserHandler struct {
	Logger  *slog.Logger
	Service *service.UserService
	Config  config.Config
}

func NewUserHandler(logger *slog.Logger, service *service.UserService, config config.Config) *UserHandler {
	return &UserHandler{
		Logger:  logger,
		Service: service,
		Config:  config,
	}
}

// LoginUser return the token if user is valid
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {

	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error("bad request", "err", err)
		httpx.HandleError(w, http.StatusBadRequest, "please provide all information", httpx.ErrBadRequest)
		return
	}

	exist, err := h.Service.CheckLoginCredential(r.Context(), &req)

	if err != nil {
		if errors.Is(err, httpx.ErrBadRequest) {
			httpx.HandleError(w, http.StatusBadRequest, "please provided the email and password correctly", httpx.ErrBadRequest)
			return
		}

		if errors.Is(err, httpx.ErrForbidden) {
			httpx.HandleError(w, http.StatusForbidden, "Invalid credential", httpx.ErrForbidden)
			return
		}

		if errors.Is(err, pgx.ErrNoRows) {
			httpx.HandleError(w, http.StatusNotFound, "No user found", httpx.ErrNotFound)
			return
		}

		httpx.HandleError(w, http.StatusInternalServerError, "something went wrong", httpx.ErrInternalServer)
		return
	}

	token, err := auth.GenerateToken(exist.ID.String(), []byte(h.Config.SecretKey))
	if err != nil {
		h.Logger.Error("failed generate token", "err", err)
		httpx.HandleError(w, http.StatusInternalServerError, "something went wrong", httpx.ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":    "loging successful",
		"token":      token,
		"token-type": "Bearer",
	}); err != nil {
		return
	}
}

// RegisterUser register the user if not exist, else return error
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {

	var req model.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error("bad request", "err", err)
		httpx.HandleError(w, http.StatusBadRequest, "please prode all information correctly", httpx.ErrBadRequest)
		return
	}

	err := h.Service.CheckAndRegisterUser(r.Context(), &req)

	if err != nil {

		if errors.Is(err, httpx.ErrBadRequest) {
			h.Logger.Error("failed user service", "err", err)
			httpx.HandleError(w, http.StatusBadRequest, "please provide the information carefully", httpx.ErrBadRequest)
			return
		}

		if errors.Is(err, httpx.ErrConflict) || errors.Is(err, httpx.ErrEmailAlreadyExist) {
			h.Logger.Error("failed user service", "err", err)
			httpx.HandleError(w, http.StatusConflict, "user already exist", httpx.ErrConflict)
			return
		}

		h.Logger.Error("failed user service", "err", err)
		httpx.HandleError(w, http.StatusInternalServerError, "something went wrong", httpx.ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "register successful",
	}); err != nil {
		return
	}
}

// GetUserDetail auth the user and then return their data, else return error
func (h *UserHandler) GetUserDetail(w http.ResponseWriter, r *http.Request) {

	reqId := middleware.GetRequestID(r.Context())
	userId := middleware.GetUserId(r.Context())

	exist, err := h.Service.Repo.FindById(r.Context(), userId)

	if err != nil {
		h.Logger.Error("failed user service", "err", err, "reqId", reqId)
		httpx.HandleError(w, http.StatusInternalServerError, "something wernt wrong", httpx.ErrInternalServer)
		return
	}

	res := model.UserResponse{
		Id:    exist.ID.String(),
		Name:  exist.Name,
		Email: exist.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return
	}
}
