package bookings

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoTicketsRequested = errors.New("no tickets requested")
	ErrTicketConflict     = errors.New("one or more tickets are not available")
)

type CreateBooking struct {
	eventID int64
	userID  int64
}

type BookTicketDetails struct {
	EventID int64
	UserID  int64
	Tickets []int64
}

type ReserveTicketDetails struct {
	EventID       int64
	UserID        int64
	Tickets       []int64
	ReservedUntil time.Time
	BookingID     int64
}

type BookingRepository interface {
	EventExists(ctx context.Context, eventID int64) (bool, error)
	ValidateTicketsForEvent(ctx context.Context, eventID int64, ticketIDs []int64) (bool, error)
	CreatePendingBooking(ctx context.Context, eventID, userID int64) (int64, error)
	ReserveTickets(ctx context.Context, details ReserveTicketDetails) ([]int64, error)
	ReleaseReservedByBooking(ctx context.Context, bookingID int64) error
	MarkBookingExpired(ctx context.Context, bookingID int64) error
	ConfirmReservedTickets(ctx context.Context, eventID, bookingID, userID int64) ([]int64, error)
	MarkBookingConfirmed(ctx context.Context, eventID, bookingID, userID int64) error
	GetBookingByID(ctx context.Context, bookingID int64) (*Booking, error)
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

func (r *BookingRepo) EventExists(ctx context.Context, eventID int64) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, eventID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *BookingRepo) ValidateTicketsForEvent(ctx context.Context, eventID int64, ticketIDs []int64) (bool, error) {
	if len(ticketIDs) == 0 {
		return false, nil
	}

	const query = `
		SELECT COUNT(*)
		FROM tickets
		WHERE event_id = $1
		  AND id = ANY($2::bigint[])
	`

	var count int
	if err := r.db.QueryRow(ctx, query, eventID, ticketIDs).Scan(&count); err != nil {
		return false, err
	}

	return count == len(ticketIDs), nil
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

func (r *BookingRepo) ReserveTickets(
	ctx context.Context,
	details ReserveTicketDetails,
) ([]int64, error) {

	const query = `
		UPDATE tickets
		SET
			status = 'reserved',
			booking_id = $1,
			reserved_until = $2
		WHERE event_id = $3
		  AND id = ANY($4::bigint[])
		  AND status = 'available'
		  AND booking_id IS NULL
		RETURNING id
	`

	rows, err := r.db.Query(
		ctx,
		query,
		details.BookingID,
		details.ReservedUntil,
		details.EventID,
		details.Tickets,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservedIDs []int64

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		reservedIDs = append(reservedIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservedIDs, nil
}

func (r *BookingRepo) CreateBooking(ctx context.Context, input CreateBooking) error {
	const query = `
		INSERT INTO bookings
		(event_id, user_id, created_by)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var bookingID int64
	err := r.db.QueryRow(ctx, query, input.eventID, input.userID, "self_booking").Scan(&bookingID)

	if err != nil {
		return err
	}

	return nil
}

func (r *BookingRepo) CreatePendingBooking(ctx context.Context, eventID, userID int64) (int64, error) {
	const query = `
		INSERT INTO bookings (event_id, user_id, created_by, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id
	`

	var bookingID int64
	if err := r.db.QueryRow(ctx, query, eventID, userID, "self_booking").Scan(&bookingID); err != nil {
		return 0, err
	}

	return bookingID, nil
}

func (r *BookingRepo) ReleaseReservedByBooking(ctx context.Context, bookingID int64) error {
	const query = `
		UPDATE tickets
		SET
			status = 'available',
			booking_id = NULL,
			reserved_until = NULL
		WHERE booking_id = $1
		  AND status = 'reserved'
	`

	_, err := r.db.Exec(ctx, query, bookingID)
	return err
}

func (r *BookingRepo) MarkBookingExpired(ctx context.Context, bookingID int64) error {
	const query = `
		UPDATE bookings
		SET status = 'expired'
		WHERE id = $1
		  AND status = 'pending'
	`

	_, err := r.db.Exec(ctx, query, bookingID)
	return err
}

func (r *BookingRepo) ConfirmReservedTickets(ctx context.Context, eventID, bookingID, userID int64) ([]int64, error) {
	const query = `
		UPDATE tickets t
		SET
			status = 'sold',
			reserved_until = NULL
		FROM bookings b
		WHERE b.id = $1
		  AND b.event_id = $2
		  AND b.user_id = $3
		  AND b.status = 'pending'
		  AND t.booking_id = b.id
		  AND t.event_id = $2
		  AND t.status = 'reserved'
		  AND t.reserved_until IS NOT NULL
		  AND t.reserved_until > NOW()
		RETURNING t.id
	`

	rows, err := r.db.Query(ctx, query, bookingID, eventID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var soldIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		soldIDs = append(soldIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return soldIDs, nil
}

func (r *BookingRepo) MarkBookingConfirmed(ctx context.Context, eventID, bookingID, userID int64) error {
	const query = `
		UPDATE bookings
		SET status = 'confirmed'
		WHERE id = $1
		  AND event_id = $2
		  AND user_id = $3
		  AND status = 'pending'
	`

	cmdTag, err := r.db.Exec(ctx, query, bookingID, eventID, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *BookingRepo) GetBookingByID(ctx context.Context, bookingID int64) (*Booking, error) {
	const query = `
		SELECT id, user_id, event_id, status, created_by, created_at
		FROM bookings
		WHERE id = $1
	`

	b := &Booking{}
	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&b.ID,
		&b.UserID,
		&b.EventID,
		&b.Status,
		&b.CreatedBy,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *BookingRepo) ReleaseExpiredTickets(ctx context.Context, ticketIDs []int64) error {

	if len(ticketIDs) == 0 {
		return nil
	}
	const query = `
		UPDATE tickets
		SET 
			status = 'available',
			reserved_until = NULL,
			booking_id = NULL
		WHERE id = ANY ($1::bigint[]) AND status = 'reserved'
	`

	_, err := r.db.Exec(ctx, query, ticketIDs)
	if err != nil {
		return err
	}
	return nil
}
