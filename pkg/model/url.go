package model

import (
	"errors"
	"strings"
	"time"
)

const (
	// DefaultTTL applies when no TTL is provided.
	DefaultTTL = 24 * time.Hour

	// MinShortLength ensures short codes aren't too short.
	MinShortLength = 4

	// MaxTTL is to prevent abuse, e.g., setting 20 years.
	MaxTTL = 30 * 24 * time.Hour // 30 days
)

var (
	ErrInvalidLongURL   = errors.New("invalid long URL")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrInvalidTTL       = errors.New("invalid TTL duration")
)

// URL is the domain model stored in PostgreSQL.
type URL struct {
	ID        int64     `json:"id"`
	ShortCode string    `json:"short_code"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateURLRequest represents an API request body for creating a shortened link.
type CreateURLRequest struct {
	LongURL string        `json:"long_url"`
	TTL     time.Duration `json:"ttl,omitempty"`
}

// CreateURLResponse represents the response object returned to clients.
type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}

// Validate ensures the request contains valid fields.
func (r *CreateURLRequest) Validate() error {
	if len(strings.TrimSpace(r.LongURL)) < 5 || !strings.Contains(r.LongURL, "://") {
		return ErrInvalidLongURL
	}

	if r.TTL < 0 {
		return ErrInvalidTTL
	}

	if r.TTL > MaxTTL {
		return ErrInvalidTTL
	}

	return nil
}

// NewURL constructs a URL object with the proper timestamps.
func NewURL(short, long string, ttl time.Duration) (*URL, error) {
	if short == "" || len(short) < MinShortLength {
		return nil, ErrInvalidShortCode
	}

	if len(long) < 5 || !strings.Contains(long, "://") {
		return nil, ErrInvalidLongURL
	}

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	now := time.Now().UTC()

	return &URL{
		ShortCode: short,
		LongURL:   long,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, nil
}

// IsExpired checks if the URL has already expired.
func (u *URL) IsExpired() bool {
	return time.Now().UTC().After(u.ExpiresAt)
}
