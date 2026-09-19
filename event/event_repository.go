package event

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidTicketCount = errors.New("Ticket count must be greater than zero")
	ErrInvalidTicketPrice = errors.New("Ticket price must be greater than or equal to zero")
	ErrEventNotFound      = errors.New("Event not found")
)

type EventDetails struct {
	Name          string
	Location      string
	PerformerName string
	Description   string
	StartTime     time.Time
	EndTime       time.Time
	TicketCount   int
	TicketPrice   float64
	TicketStatus  string
}

type EventRepository interface {
	CreateEvent(ctx context.Context, input EventDetails) (*Event, error)
	GetByID(ctx context.Context, id int64) (*Event, error)
}

type EventRepo struct {
	db *pgxpool.Pool
}

func NewEventRepo(db *pgxpool.Pool) *EventRepo {
	return &EventRepo{
		db: db,
	}
}

func (r *EventRepo) CreateEvent(ctx context.Context, input EventDetails) (*Event, error) {
	if input.TicketCount == 0 {
		return nil, ErrInvalidTicketCount
	}
	if input.TicketPrice <= 0 {
		return nil, ErrInvalidTicketPrice
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const insertEventQuery = `
		INSERT INTO events (
			name,
			location,
			performer_name,
			description,
			start_time,
			end_time
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			name,
			location,
			performer_name,
			description,
			start_time,
			end_time,
			created_at
	`

	e := &Event{}
	err = tx.QueryRow(
		ctx,
		insertEventQuery,
		input.Name,
		input.Location,
		input.PerformerName,
		input.Description,
		input.StartTime,
		input.EndTime,
	).Scan(
		&e.ID,
		&e.Name,
		&e.Location,
		&e.PerformerName,
		&e.Description,
		&e.StartTime,
		&e.EndTime,
		&e.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	const insertTicketQuery = `
		INSERT INTO tickets (
			event_id, booking_id, status, price
		)
		SELECT $1, NULL, $2, $3
		FROM generate_series(1, $4)
	`

	_, err = tx.Exec(ctx, insertTicketQuery, e.ID, input.TicketStatus, input.TicketPrice, input.TicketCount)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *EventRepo) GetByID(ctx context.Context, id int64) (*Event, error) {
	const query = `
		SELECT 
		id, name, location, performer_name, desciption, start_time, end_time, created_at
		FROM
		events
		WHERE id = $1
	`

	var event Event

	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Location,
		&event.PerformerName,
		&event.Description,
		&event.StartTime,
		&event.EndTime,
		&event.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return &event, nil
}
