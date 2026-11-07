package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/model"
)

type URLRepository interface {
	FindByLongURL(ctx context.Context, longUrl string) (*model.URL, bool, error)

	FindByShortCode(ctx context.Context, shortCode string) (*model.UrlLongUrl, error)

	Create(ctx context.Context, url *model.URL) error
}

type RedisCache interface {
	Get(ctx context.Context, shortCode string) (string, bool, error)
	Set(ctx context.Context, shortCode string, longUrl string) error
	Delete(ctx context.Context, shortCode string) error
}

type URLService struct {
	Repo   URLRepository
	Redis  RedisCache
	Logger *slog.Logger
}

func NewURLService(repo URLRepository, redis RedisCache, logger *slog.Logger) *URLService {
	return &URLService{
		Repo:   repo,
		Redis:  redis,
		Logger: logger,
	}
}

func (h *URLService) CreateShortCode(ctx context.Context, longUrl string) (*model.URL, error) {

	// check if the long url was provided or not
	if strings.TrimSpace(longUrl) == "" {
		return nil, httpx.ErrBadRequest
	}

	// check if long url exist
	existing, found, err := h.Repo.FindByLongURL(ctx, longUrl)

	if err != nil {
		return nil, err
	}

	// return data if already present
	if found {
		return existing, nil
	}

	// generate short code
	shortCode, err := GenerateShortCode()
	if err != nil {
		return nil, err
	}

	newURL := &model.URL{
		LongUrl:   longUrl,
		ShortCode: shortCode,
	}

	err = h.Repo.Create(ctx, newURL)
	if err != nil {
		return nil, err
	}

	return newURL, nil
}

func (h *URLService) FindLongUrl(ctx context.Context, shortCode string) (*model.UrlLongUrl, error) {

	// check if short code was provided
	if strings.TrimSpace(shortCode) == "" {
		return nil, httpx.ErrBadRequest
	}

	// check in redis
	if h.Redis != nil {
		longUrl, found, err := h.Redis.Get(ctx, shortCode)

		if err != nil {
			return nil, err
		}

		if found {
			return &model.UrlLongUrl{
				LongUrl: longUrl,
			}, nil
		}
	}

	h.Logger.Info("no data found in redis, redirecting to database")

	// check if short code exist
	existing, err := h.Repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	err = h.Redis.Set(ctx, shortCode, existing.LongUrl)
	if err != nil {
		return nil, err
	}

	return existing, nil
}
