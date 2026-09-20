package bookings

import "time"

type Booking struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	EventID   int       `json:"event_id"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}
