package bookings

type Booking struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	EventID   int    `json:"event_id"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}
