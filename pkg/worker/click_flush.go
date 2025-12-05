package worker

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/JpUnique/TinyGo/pkg/store"
)

// ClickFlusher periodically scans Redis click:* keys, atomically reads & clears them,
// then persists aggregated counts into Postgres.
type ClickFlusher struct {
	pg        *store.Postgres
	rd        *store.RedisStore
	interval  time.Duration
	batchSize int64
	stopCh    chan struct{}
}

// NewClickFlusher creates a new flusher.
func NewClickFlusher(pg *store.Postgres, rd *store.RedisStore, interval time.Duration) *ClickFlusher {
	return &ClickFlusher{
		pg:        pg,
		rd:        rd,
		interval:  interval,
		batchSize: 100,
		stopCh:    make(chan struct{}),
	}
}

// Start runs the flusher loop until context is cancelled or Stop is called.
func (f *ClickFlusher) Start(ctx context.Context) {
	t := time.NewTicker(f.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("click flusher stopping (ctx done)")
			return
		case <-f.stopCh:
			log.Println("click flusher stopping (stop requested)")
			return
		case <-t.C:
			f.flushOnce(ctx)
		}
	}
}

func (f *ClickFlusher) Stop() {
	select {
	case f.stopCh <- struct{}{}:
	default:
	}
}

// flushOnce scans for click keys, transfers counts to Postgres
func (f *ClickFlusher) flushOnce(ctx context.Context) {
	var cursor uint64 = 0
	for {
		newCursor, keys, err := f.rd.ScanClickKeys(ctx, cursor, f.batchSize)
		if err != nil {
			log.Printf("click flusher: scan error: %v", err)
			return
		}

		for _, k := range keys {
			// k is like "click:abc123" - trim prefix
			short := strings.TrimPrefix(k, "click:")
			count, err := f.rd.GetAndClearClicks(ctx, short)
			if err != nil {
				if err == store.ErrClickNotFound {
					continue
				}
				log.Printf("click flusher: get-and-clear error for %s: %v", short, err)
				continue
			}
			// Persist to Postgres
			if err := f.pg.IncrementClicks(ctx, short, count); err != nil {
				log.Printf("click flusher: pg increment error for %s: %v", short, err)
				// If persisting failed, we lost the counter (since it was cleared). In production, consider atomic LIST or streams.
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
}
