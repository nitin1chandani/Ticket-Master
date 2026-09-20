package bookings

import (
	"context"
	"errors"
)

var ErrInvalidTicketCount = errors.New("you should have atleast one ticket")
var ErrDuplicateTickets = errors.New("duplicate ticket ids in request")
var ErrInvalidTickets = errors.New("tickets do not belong to event")
var ErrEventNotFound = errors.New("event not found")

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
