package bookings

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/nitin1chandani/ticketmaster/internal/httpx"
	"github.com/nitin1chandani/ticketmaster/internal/middleware"
)

type BookingHandler struct {
	bookingService *BookingService
}

func NewBookingHandler(bookingService *BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

type BookEventRequest struct {
	TicketIDs []int64 `json:"ticket_ids" validate:"required"`
}

func (h *BookingHandler) BookTickets(c *fiber.Ctx) error {
	stringEventID := c.Params("event_id")
	eventID, err := strconv.Atoi(stringEventID)
	if err != nil || eventID == 0 {
		return &httpx.AppError{
			StatusCode: fiber.StatusBadRequest,
			Code:       "INVALID_EVENT_ID",
			Message:    "valid event_id is required",
		}
	}

	// Bind and validate req
	var req BookEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	// pass it to service
	serviceInput := &BookTicketInput{
		EventID: int64(eventID),
		Tickets: req.TicketIDs,
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok || userID <= 0 {
		return &httpx.AppError{
			StatusCode: fiber.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "invalid token subject",
		}
	}
	serviceInput.UserID = int64(userID)

	booking, err := h.bookingService.BookTickets(c.Context(), serviceInput)
	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			return &httpx.AppError{
				StatusCode: fiber.StatusNotFound,
				Code:       "EVENT_NOT_FOUND",
				Message:    "event not found",
			}
		case errors.Is(err, ErrInvalidTicketCount), errors.Is(err, ErrInvalidTickets):
			return &httpx.AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       "INVALID_TICKETS",
				Message:    err.Error(),
			}
		case errors.Is(err, ErrTicketConflict):
			return &httpx.AppError{
				StatusCode: fiber.StatusConflict,
				Code:       "TICKET_CONFLICT",
				Message:    "one or more tickets are already sold",
			}
		default:
			return &httpx.AppError{
				StatusCode: fiber.StatusInternalServerError,
				Code:       "BOOKING_FAILED",
				Message:    "failed to book tickets",
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tickets booked successfully",
		"data":    booking,
	})
}
