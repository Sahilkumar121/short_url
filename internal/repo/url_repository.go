package repo

import (
	"context"
	"errors"

	"github.com/Sahilkumar121/short_url/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresqlURLRepositoryHandler struct {
	DB *pgxpool.Pool
}

func NewURLRepositoryHandler(db *pgxpool.Pool) *PostgresqlURLRepositoryHandler {
	return &PostgresqlURLRepositoryHandler{
		DB: db,
	}
}

// FindByLongURL: return URL if long url exist
func (h *PostgresqlURLRepositoryHandler) FindByLongURL(ctx context.Context, longUrl string) (*model.URL, bool, error) {

	query := `select id, long_url, short_code, created_at from urls where long_url = $1`

	rows := h.DB.QueryRow(ctx, query, longUrl)

	var url model.URL
	err := rows.Scan(&url.Id, &url.LongUrl, &url.ShortCode, &url.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return &url, true, nil
}

// FindByShortCode: return URL if short code exist
func (h *PostgresqlURLRepositoryHandler) FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {

	query := `select id, long_url, short_code, created_at from urls where short_code = $1`
	rows := h.DB.QueryRow(ctx, query, shortCode)

	var url model.URL
	err := rows.Scan(&url.Id, &url.LongUrl, &url.ShortCode, &url.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (h *PostgresqlURLRepositoryHandler) Create(ctx context.Context, url *model.URL) error {

	query := `insert into urls (long_url, short_code) values ($1, $2)`
	_, err := h.DB.Exec(ctx, query, url.LongUrl, url.ShortCode)

	return err
}
