package event

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
)

type EventHandler struct {
	eventService *EventService
}

func NewEventHandler(eventService *EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

type CreateEventRequest struct {
	Name          string    `json:"name" validate:"required,min=1,max=200"`
	Location      string    `json:"location" validate:"required,min=4,max=200"`
	PerformerName string    `json:"performer_name" validate:"required,min=4,max=200"`
	Description   string    `json:"description" validate:"required,min=10,max=1000"`
	StartTime     time.Time `json:"start_time" validate:"required"`
	EndTime       time.Time `json:"end_time" validate:"required"`
	TicketCount   int       `json:"ticket_count" validate:"required,gt=0"`
	TicketPrice   float64   `json:"ticket_price" validate:"required,gte=0"`
}

type CreateEventResponse struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Location      string    `json:"location"`
	PerformerName string    `json:"performer_name"`
	Description   string    `json:"description"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *EventHandler) CreateNewEvent(c *fiber.Ctx) error {
	var req CreateEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	event, err := h.eventService.CreateEvent(c.Context(), &EventDetails{
		Name:          req.Name,
		Location:      req.Location,
		PerformerName: req.PerformerName,
		Description:   req.Description,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		TicketCount:   req.TicketCount,
		TicketPrice:   req.TicketPrice,
	})

	if err != nil {
		if errors.Is(err, ErrInvalidStartTime) {
			return &httpx.AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       "INVALID_START_TIME",
				Message:    err.Error(),
			}
		}

		if errors.Is(err, ErrInvalidEndTime) {
			return &httpx.AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       "INVALID_END_TIME",
				Message:    err.Error(),
			}
		}

		return &httpx.AppError{
			StatusCode: fiber.StatusInternalServerError,
			Code:       "EVENT_CREATION_FAILED",
			Message:    "Failed to create event",
		}
	}

	resp := CreateEventResponse{
		ID:            event.ID,
		Name:          event.Name,
		Location:      event.Location,
		PerformerName: event.PerformerName,
		Description:   event.Description,
		StartTime:     event.StartTime,
		EndTime:       event.EndTime,
		CreatedBy:     event.CreatedBy,
		CreatedAt:     event.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *EventHandler) GetEventDetailsWithAvailableTickets(c *fiber.Ctx) error {
	eventID, err := strconv.Atoi(c.Params("event_id"))
	if err != nil || eventID <= 0 {
		return &httpx.AppError{
			StatusCode: fiber.StatusBadRequest,
			Code:       "INVALID_EVENT_ID",
			Message:    "Invalid event_id",
		}
	}

	eventDetails, err := h.eventService.GetAvailableTicketsWithEventDetails(c.Context(), int64(eventID))

	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			return &httpx.AppError{
				StatusCode: fiber.StatusNotFound,
				Code:       "EVENT_DOES_NOT_EXIST",
				Message:    "event does not exist",
			}
		}
		return &httpx.AppError{
			StatusCode: fiber.StatusInternalServerError,
			Code:       "EVENT_DETAILS_FETCH_FAILED",
			Message:    "failed to fetch event details",
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    eventDetails,
	})

}
