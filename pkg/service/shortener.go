package service

import (
	"context"
	"errors"
	"time"

	"github.com/JpUnique/TinyGo/pkg/store"
	"github.com/JpUnique/TinyGo/pkg/utils"
)

var (
	ErrAliasTaken = errors.New("custom alias is already taken")
)

type ShortenerService struct {
	BaseURL string
	pg      *store.Postgres
	rd      *store.RedisStore
	// other fields like storage, etc.
}

func NewShortener(pg *store.Postgres, rd *store.RedisStore, baseURL string) *ShortenerService {
	return &ShortenerService{
		BaseURL: baseURL,
		pg:      pg,
		rd:      rd,
	}
}

func (s *ShortenerService) Create(ctx context.Context, longURL, custom string) (short string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	short = custom
	if short == "" {
		for i := 0; i < 5; i++ {
			short = utils.RandomBase62(7)
			err := s.pg.InsertURL(ctx, short, longURL)
			if err != nil {
				continue
			}

			break
		}
	} else {
		err := s.pg.InsertURL(ctx, short, longURL)
		if err != nil {
			return "", ErrAliasTaken
		}
	}
	// set cache best-effortlike like
	_ = s.rd.SetShort(ctx, short, longURL, 24*time.Hour)
	return short, err
}

func (s *ShortenerService) Resolve(ctx context.Context, short string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// try cache
	if u, err := s.rd.GetShort(ctx, short); err == nil && u != "" {
		// increment click counter in redis
		_, _ = s.rd.IncrClicks(ctx, short)
		return u, nil
	}

	// fallback to DB
	u, err := s.pg.GetByShort(ctx, short)
	if err != nil {
		return "", err
	}
	// populate cache for next time
	_ = s.rd.SetShort(ctx, short, u.LongURL, 24*time.Hour)
	// increment click in DB asynchronously
	go func() { _ = s.pg.IncrementClicks(context.Background(), short, 1) }()

	return u.LongURL, nil
}
