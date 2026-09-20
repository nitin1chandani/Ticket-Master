package event

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidStartTime = errors.New("Invalid start time")
	ErrInvalidEndTime   = errors.New("Invalid end time")
)

type EventService struct {
	EventRepo EventRepository
}

func NewEventService(eventRepo EventRepository) (e *EventService) {
	return &EventService{
		EventRepo: eventRepo,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, input *EventDetails) (*Event, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Location = strings.TrimSpace(input.Location)
	input.PerformerName = strings.TrimSpace(input.PerformerName)
	input.Description = strings.TrimSpace(input.Description)

	if input.StartTime.IsZero() || !input.StartTime.After(time.Now().UTC()) {
		return nil, ErrInvalidStartTime
	}

	if input.EndTime.IsZero() || !input.EndTime.After(input.StartTime) {
		return nil, ErrInvalidEndTime
	}

	if input.TicketStatus == "" {
		input.TicketStatus = "available"
	}

	return s.EventRepo.CreateEvent(ctx, *input)
}

func (s *EventService) GetAvailableTicketsWithEventDetails(ctx context.Context, eventID int64) (*EventDetailsWithAvailableTickets, error) {
	return s.EventRepo.GetEventDetailsWithTickets(ctx, eventID)
}
