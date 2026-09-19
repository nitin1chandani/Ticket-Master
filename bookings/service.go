package bookings

import (
	"context"
	"errors"
)

var ErrInvalidTicketCount = errors.New("you should have atleast one ticket")
var ErrInvalidTickets = errors.New("tickets")

type BookingService struct {
	repo *BookingRepo
}

func NewBookingService(repo *BookingRepo) *BookingService {
	return &BookingService{
		repo: repo,
	}
}

type BookTicketInput struct {
	EventID int64
	UserID  int64
	Tickets []int64
}

func (s *BookingService) BookTickets(ctx context.Context, input *BookTicketInput) (*Booking, error) {
	if len(input.Tickets) == 0 {
		return nil, ErrInvalidTicketCount
	}
	return nil, nil
}
