package service

import (
	"context"
	"log/slog"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/model"
	"github.com/Sahilkumar121/short_url/internal/security"
)

type UserService struct {
	Repo     UserRepository
	notifier LoginNotifier
	Logger   *slog.Logger
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindById(ctx context.Context, id string) (*model.User, error)
	RegisterUserInDatabase(ctx context.Context, userReq *model.UserRegisterRequest) error
}

type LoginNotifier interface {
	NotifyLogin(email, name string) error
}

func NewUserService(repo UserRepository, notifier LoginNotifier, logger *slog.Logger) *UserService {
	return &UserService{
		Repo:     repo,
		notifier: notifier,
		Logger:   logger,
	}
}

func (h *UserService) CheckLoginCredential(ctx context.Context, userReq *model.UserLoginRequest) (*model.User, error) {

	if userReq.Email == "" || userReq.Password == "" {
		return nil, httpx.ErrBadRequest
	}

	exist, err := h.Repo.FindByEmail(ctx, userReq.Email)

	if err != nil {
		return nil, err
	}

	ok := security.VerifyPassword(userReq.Password, exist.Password)
	if !ok {
		return nil, httpx.ErrForbidden
	}

	if err := h.notifier.NotifyLogin(exist.Email, exist.Name); err != nil {
		h.Logger.Error("failed to enque the email notification", "user", exist.Email, "err", err)
	}

	return exist, nil
}

func (h *UserService) CheckAndRegisterUser(ctx context.Context, userReq *model.UserRegisterRequest) error {

	if userReq.Name == "" || userReq.Email == "" || userReq.Password == "" {
		return httpx.ErrBadRequest
	}

	exist, err := h.Repo.FindByEmail(ctx, userReq.Email)
	if err != nil {
		return err
	}

	if exist != nil {
		return httpx.ErrConflict
	}

	// generate hash password
	hashPassword, err := security.HashPassword(userReq.Password)
	if err != nil {
		return err
	}

	userReq.Password = hashPassword

	// register user

	err = h.Repo.RegisterUserInDatabase(ctx, userReq)
	if err != nil {
		return err
	}

	return nil
}
