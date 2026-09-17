package event

type Event struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Location      string `json:"location"`
	PerformerName string `json:"performer_name"`
	Description   string `json:"description"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}
