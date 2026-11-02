package repo

import (
	"context"
	"errors"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresqlUserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepositoryHandler(db *pgxpool.Pool) *PostgresqlURLRepositoryHandler {
	return &PostgresqlURLRepositoryHandler{
		DB: db,
	}
}

func (h *PostgresqlURLRepositoryHandler) FindByEmail(ctx context.Context, email string) (*model.User, error) {

	query := `select id, name, email, password, created_at from users where email = $1`
	row := h.DB.QueryRow(ctx, query, email)

	var userData model.User
	err := row.Scan(&userData.ID, &userData.Name, &userData.Email, &userData.Password, &userData.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &userData, nil
}

func (h *PostgresqlURLRepositoryHandler) RegisterUserInDatabase(ctx context.Context, userReq *model.UserRegisterRequest) error {

	query := `insert into users (name, email, password) values($1, $2, $3)`

	_, err := h.DB.Exec(ctx, query, userReq.Name, userReq.Email, userReq.Password)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
			return httpx.ErrEmailAlreadyExist
		}
		return err
	}

	return nil
}

func (h *PostgresqlURLRepositoryHandler) FindById(ctx context.Context, id string) (*model.User, error) {

	query := `select id, name, email, password, created_at from users where id = $1`

	var userData model.User
	row := h.DB.QueryRow(ctx, query, id)

	err := row.Scan(&userData)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}
