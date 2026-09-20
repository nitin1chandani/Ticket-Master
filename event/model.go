package event

import "time"

type Event struct {
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

type Ticket struct {
	ID            int        `json:"id"`
	EventID       int        `json:"event_id"`
	BookingID     int        `json:"booking_id"`
	ReservedUntil *time.Time `json:"reserved_until"`

	// Status can be "available", "reserved", or "sold"
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}
