package bookings

import (
	"errors"
	"strconv"
	"time"

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

type ReserveTicketsRequest struct {
	TicketIDs  []int64 `json:"ticket_ids" validate:"required,min=1,dive,gt=0"`
	TTLSeconds int64   `json:"ttl_seconds" validate:"omitempty,gt=0"`
}

type ConfirmBookingRequest struct {
	BookingID int64 `json:"booking_id" validate:"required,gt=0"`
}

func (h *BookingHandler) BookTickets(c *fiber.Ctx) error {
	stringEventID := c.Params("event_id")
	eventID, err := strconv.Atoi(stringEventID)
	if err != nil || eventID <= 0 {
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
		case errors.Is(err, ErrInvalidTicketCount), errors.Is(err, ErrInvalidTickets), errors.Is(err, ErrDuplicateTickets):
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

func (h *BookingHandler) ReserveTickets(c *fiber.Ctx) error {
	serviceInput, err := reserveInputFromCtx(c)
	if err != nil {
		return err
	}

	result, err := h.bookingService.ReserveTickets(c.Context(), serviceInput)
	if err != nil {
		switch {
		case errors.Is(err, ErrEventNotFound):
			return &httpx.AppError{
				StatusCode: fiber.StatusNotFound,
				Code:       "EVENT_NOT_FOUND",
				Message:    "event not found",
			}
		case errors.Is(err, ErrInvalidTicketCount), errors.Is(err, ErrInvalidTickets), errors.Is(err, ErrDuplicateTickets):
			return &httpx.AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       "INVALID_TICKETS",
				Message:    err.Error(),
			}
		case errors.Is(err, ErrTicketConflict):
			return &httpx.AppError{
				StatusCode: fiber.StatusConflict,
				Code:       "TICKET_CONFLICT",
				Message:    "one or more tickets are already reserved or sold",
			}
		default:
			return &httpx.AppError{
				StatusCode: fiber.StatusInternalServerError,
				Code:       "RESERVATION_FAILED",
				Message:    "failed to reserve tickets",
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "tickets reserved successfully",
		"data":    result,
	})
}

func (h *BookingHandler) ConfirmBooking(c *fiber.Ctx) error {
	stringEventID := c.Params("event_id")
	eventID, err := strconv.Atoi(stringEventID)
	if err != nil || eventID <= 0 {
		return &httpx.AppError{
			StatusCode: fiber.StatusBadRequest,
			Code:       "INVALID_EVENT_ID",
			Message:    "valid event_id is required",
		}
	}

	var req ConfirmBookingRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok || userID <= 0 {
		return &httpx.AppError{
			StatusCode: fiber.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "invalid token subject",
		}
	}

	booking, err := h.bookingService.ConfirmBooking(c.Context(), &ConfirmBookingInput{
		EventID:   int64(eventID),
		UserID:    int64(userID),
		BookingID: req.BookingID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBookingID):
			return &httpx.AppError{
				StatusCode: fiber.StatusBadRequest,
				Code:       "INVALID_BOOKING_ID",
				Message:    err.Error(),
			}
		case errors.Is(err, ErrBookingNotFound):
			return &httpx.AppError{
				StatusCode: fiber.StatusNotFound,
				Code:       "BOOKING_NOT_FOUND",
				Message:    "booking not found",
			}
		case errors.Is(err, ErrReservationExpired), errors.Is(err, ErrTicketConflict):
			return &httpx.AppError{
				StatusCode: fiber.StatusConflict,
				Code:       "RESERVATION_EXPIRED",
				Message:    "reservation expired or already processed",
			}
		default:
			return &httpx.AppError{
				StatusCode: fiber.StatusInternalServerError,
				Code:       "BOOKING_CONFIRM_FAILED",
				Message:    "failed to confirm booking",
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "booking confirmed",
		"data":    booking,
	})
}

func reserveInputFromCtx(c *fiber.Ctx) (*ReserveTicketsInput, error) {
	stringEventID := c.Params("event_id")
	eventID, err := strconv.Atoi(stringEventID)
	if err != nil || eventID <= 0 {
		return nil, &httpx.AppError{
			StatusCode: fiber.StatusBadRequest,
			Code:       "INVALID_EVENT_ID",
			Message:    "valid event_id is required",
		}
	}

	var req ReserveTicketsRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return nil, err
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok || userID <= 0 {
		return nil, &httpx.AppError{
			StatusCode: fiber.StatusUnauthorized,
			Code:       "UNAUTHORIZED",
			Message:    "invalid token subject",
		}
	}

	ttl := defaultReservationTTL
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	return &ReserveTicketsInput{
		EventID: int64(eventID),
		UserID:  int64(userID),
		Tickets: req.TicketIDs,
		TTL:     ttl,
	}, nil
}
