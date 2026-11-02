package service

import (
	"context"
	"strings"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/model"
)

type URLRepository interface {
	FindByLongURL(ctx context.Context, longUrl string) (*model.URL, bool, error)

	FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error)

	Create(ctx context.Context, url *model.URL) error
}

type URLService struct {
	Repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
	return &URLService{
		Repo: repo,
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

func (h *URLService) GenerateLongUrl(ctx context.Context, shortCode string) (*model.URL, error) {

	// check if short code was provided
	if strings.TrimSpace(shortCode) == "" {
		return nil, httpx.ErrBadRequest
	}

	// check if short code exist
	existing, err := h.Repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return existing, nil
}
