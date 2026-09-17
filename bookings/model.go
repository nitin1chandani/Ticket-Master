package bookings

type Booking struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	EventID   int    `json:"event_id"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

type Ticket struct {
	ID        int `json:"id"`
	EventID   int `json:"event_id"`
	BookingID int `json:"booking_id"`
	// Status can be "available", "reserved", or "sold"
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}
