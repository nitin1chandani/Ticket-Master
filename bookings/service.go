package bookings

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidTicketCount = errors.New("you should have atleast one ticket")
var ErrDuplicateTickets = errors.New("duplicate ticket ids in request")
var ErrInvalidTickets = errors.New("tickets do not belong to event")
var ErrEventNotFound = errors.New("event not found")
var ErrInvalidBookingID = errors.New("invalid booking id")
var ErrBookingNotFound = errors.New("booking not found")
var ErrReservationExpired = errors.New("reservation expired")

const defaultReservationTTL = 10 * time.Minute

type BookingService struct {
	bookingRepo BookingRepository
}

func NewBookingService(repo BookingRepository) *BookingService {
	return &BookingService{
		bookingRepo: repo,
	}
}

type BookTicketInput struct {
	EventID int64
	UserID  int64
	Tickets []int64
}

type ReserveTicketsInput struct {
	EventID int64
	UserID  int64
	Tickets []int64
	TTL     time.Duration
}

type ConfirmBookingInput struct {
	EventID   int64
	UserID    int64
	BookingID int64
}

type ReserveTicketsResult struct {
	BookingID         int64     `json:"booking_id"`
	ReservedTicketIDs []int64   `json:"reserved_ticket_ids"`
	ReservedUntil     time.Time `json:"reserved_until"`
}

func (s *BookingService) BookTickets(ctx context.Context, input *BookTicketInput) (*Booking, error) {
	if input == nil {
		return nil, ErrInvalidTicketCount
	}

	if len(input.Tickets) == 0 {
		return nil, ErrInvalidTicketCount
	}

	seen := make(map[int64]struct{}, len(input.Tickets))
	for _, ticketID := range input.Tickets {
		if _, ok := seen[ticketID]; ok {
			return nil, ErrDuplicateTickets
		}
		seen[ticketID] = struct{}{}
	}

	eventExists, err := s.bookingRepo.EventExists(ctx, input.EventID)
	if err != nil {
		return nil, err
	}
	if !eventExists {
		return nil, ErrEventNotFound
	}

	validTickets, err := s.bookingRepo.ValidateTicketsForEvent(ctx, input.EventID, input.Tickets)
	if err != nil {
		return nil, err
	}
	if !validTickets {
		return nil, ErrInvalidTickets
	}

	return s.bookingRepo.BookTickets(ctx, BookTicketDetails{
		EventID: input.EventID,
		UserID:  input.UserID,
		Tickets: input.Tickets,
	})
}

func (s *BookingService) ReserveTickets(ctx context.Context, input *ReserveTicketsInput) (*ReserveTicketsResult, error) {
	if input == nil {
		return nil, ErrInvalidTicketCount
	}
	if input.TTL <= 0 {
		input.TTL = defaultReservationTTL
	}

	bookInput := &BookTicketInput{
		EventID: input.EventID,
		UserID:  input.UserID,
		Tickets: input.Tickets,
	}

	if err := s.validateBookInput(ctx, bookInput); err != nil {
		return nil, err
	}

	bookingID, err := s.bookingRepo.CreatePendingBooking(ctx, input.EventID, input.UserID)
	if err != nil {
		return nil, err
	}

	reservedUntil := time.Now().UTC().Add(input.TTL)
	reservedIDs, err := s.bookingRepo.ReserveTickets(ctx, ReserveTicketDetails{
		EventID:       input.EventID,
		UserID:        input.UserID,
		Tickets:       input.Tickets,
		ReservedUntil: reservedUntil,
		BookingID:     bookingID,
	})
	if err != nil {
		_ = s.bookingRepo.MarkBookingExpired(ctx, bookingID)
		return nil, err
	}

	if len(reservedIDs) != len(input.Tickets) {
		_ = s.bookingRepo.ReleaseReservedByBooking(ctx, bookingID)
		_ = s.bookingRepo.MarkBookingExpired(ctx, bookingID)
		return nil, ErrTicketConflict
	}

	return &ReserveTicketsResult{
		BookingID:         bookingID,
		ReservedTicketIDs: reservedIDs,
		ReservedUntil:     reservedUntil,
	}, nil
}

func (s *BookingService) ConfirmBooking(ctx context.Context, input *ConfirmBookingInput) (*Booking, error) {
	if input == nil || input.BookingID <= 0 {
		return nil, ErrInvalidBookingID
	}

	soldIDs, err := s.bookingRepo.ConfirmReservedTickets(ctx, input.EventID, input.BookingID, input.UserID)
	if err != nil {
		return nil, err
	}
	if len(soldIDs) == 0 {
		_ = s.bookingRepo.MarkBookingExpired(ctx, input.BookingID)
		return nil, ErrReservationExpired
	}

	if err := s.bookingRepo.MarkBookingConfirmed(ctx, input.EventID, input.BookingID, input.UserID); err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	booking, err := s.bookingRepo.GetBookingByID(ctx, input.BookingID)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) validateBookInput(ctx context.Context, input *BookTicketInput) error {
	if input == nil {
		return ErrInvalidTicketCount
	}

	if len(input.Tickets) == 0 {
		return ErrInvalidTicketCount
	}

	seen := make(map[int64]struct{}, len(input.Tickets))
	for _, ticketID := range input.Tickets {
		if _, ok := seen[ticketID]; ok {
			return ErrDuplicateTickets
		}
		seen[ticketID] = struct{}{}
	}

	eventExists, err := s.bookingRepo.EventExists(ctx, input.EventID)
	if err != nil {
		return err
	}
	if !eventExists {
		return ErrEventNotFound
	}

	validTickets, err := s.bookingRepo.ValidateTicketsForEvent(ctx, input.EventID, input.Tickets)
	if err != nil {
		return err
	}
	if !validTickets {
		return ErrInvalidTickets
	}

	return nil
}
