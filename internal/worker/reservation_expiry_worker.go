package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReservationExpiryWorker struct {
	db       *pgxpool.Pool
	interval time.Duration
	timeout  time.Duration
	logger   *slog.Logger
}

func NewReservationExpiryWorker(db *pgxpool.Pool, interval, timeout time.Duration, logger *slog.Logger) *ReservationExpiryWorker {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &ReservationExpiryWorker{
		db:       db,
		interval: interval,
		timeout:  timeout,
		logger:   logger,
	}
}

func (w *ReservationExpiryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("reservation expiry worker stopped")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *ReservationExpiryWorker) runOnce(ctx context.Context) {
	runCtx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	released, expired, err := w.releaseExpired(runCtx)
	if err != nil {
		w.logger.Error("reservation expiry run failed", "error", err)
		return
	}

	if released > 0 || expired > 0 {
		w.logger.Info("reservation expiry run completed", "released_tickets", released, "expired_bookings", expired)
	}
}

func (w *ReservationExpiryWorker) releaseExpired(ctx context.Context) (int64, int64, error) {
	const query = `
        WITH expired AS (
            SELECT id, booking_id
            FROM tickets
            WHERE status = 'reserved'
              AND reserved_until IS NOT NULL
              AND reserved_until <= NOW()
        ),
        updated_tickets AS (
            UPDATE tickets t
            SET status = 'available',
                booking_id = NULL,
                reserved_until = NULL
            FROM expired e
            WHERE t.id = e.id
            RETURNING e.booking_id
        ),
        updated_bookings AS (
            UPDATE bookings b
            SET status = 'expired'
            WHERE b.id IN (
                SELECT DISTINCT booking_id
                FROM updated_tickets
                WHERE booking_id IS NOT NULL
            )
              AND b.status = 'pending'
            RETURNING b.id
        )
        SELECT
            (SELECT COUNT(*) FROM updated_tickets),
            (SELECT COUNT(*) FROM updated_bookings);
    `

	var releasedTickets int64
	var expiredBookings int64

	if err := w.db.QueryRow(ctx, query).Scan(&releasedTickets, &expiredBookings); err != nil {
		return 0, 0, err
	}

	return releasedTickets, expiredBookings, nil
}
