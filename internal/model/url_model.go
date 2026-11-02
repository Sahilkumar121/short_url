package model

import (
	"time"

	"github.com/google/uuid"
)

// url dto
type URL struct {
	Id        uuid.UUID  `json:"id"`
	LongUrl   string     `json:"long_url"`
	ShortCode string     `json:"short_code"`
	UserId    uuid.UUID  `json:"user_id"`
	CreatedAt *time.Time `json:"created_at"`
}

type UrlRequest struct {
	LongUrl string `json:"long_url"`
}
