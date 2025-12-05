package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	cfg.MinConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	//ping to verify connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) InsertURL(ctx context.Context, short, long string) error {
	sql := `INSERT INTO urls (short_key, long_url, created_at) VALUES ($1, $2, NOW())`
	_, err := p.pool.Exec(ctx, sql, short, long)
	if err != nil {
		return err
	}
	return nil
}

// URLRow maps DB
type URLRow struct {
	ShortKey   string
	LongURL    string
	ClickCount int64
}

func (p *Postgres) GetByShort(ctx context.Context, short string) (*URLRow, error) {
	sql := `SELECT short_key, long_url, click_count FROM urls WHERE short_key = $1`
	row := p.pool.QueryRow(ctx, sql, short)
	var u URLRow
	if err := row.Scan(&u.ShortKey, &u.LongURL, &u.ClickCount); err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &u, nil
}

func (p *Postgres) IncrementClicks(ctx context.Context, short string, delta int64) error {
	sql := `UPDATE urls SET click_count = click_count + $1 WHERE short_key = $2`
	_, err := p.pool.Exec(ctx, sql, delta, short)
	return err
}
