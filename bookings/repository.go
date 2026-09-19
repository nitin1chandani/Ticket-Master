package bookings

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoTicketsRequested = errors.New("no tickets requested")
	ErrTicketConflict     = errors.New("one or more tickets are not available")
)

type BookTicketDetails struct {
	EventID int64
	UserID  int64
	Tickets []int64
}

type BookingRepository interface {
	BookTickets(ctx context.Context, details BookTicketDetails) (*Booking, error)
}

type BookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{
		db: db,
	}
}

func (r *BookingRepo) BookTickets(ctx context.Context, details BookTicketDetails) (*Booking, error) {
	if len(details.Tickets) == 0 {
		return nil, ErrNoTicketsRequested
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const insertBookingQuery = `
        INSERT INTO bookings (event_id, user_id, created_by)
        VALUES ($1, $2, $3)
        RETURNING id, event_id, user_id, created_by, created_at
    `

	b := &Booking{}
	err = tx.QueryRow(
		ctx,
		insertBookingQuery,
		details.EventID,
		details.UserID,
		"self_booking",
	).Scan(
		&b.ID,
		&b.EventID,
		&b.UserID,
		&b.CreatedBy,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	const updateTicketQuery = `
        UPDATE tickets
        SET booking_id = $1,
            status = 'sold'
        WHERE event_id = $2
          AND id = ANY($3::bigint[])
          AND booking_id IS NULL
          AND status = 'available'
        RETURNING id
    `

	rows, err := tx.Query(ctx, updateTicketQuery, b.ID, details.EventID, details.Tickets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedCount := 0
	for rows.Next() {
		var ticketID int64
		if err := rows.Scan(&ticketID); err != nil {
			return nil, err
		}
		updatedCount++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if updatedCount != len(details.Tickets) {
		return nil, ErrTicketConflict
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return b, nil
}
