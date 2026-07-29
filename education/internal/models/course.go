package models

type Course struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	DurationHrs int    `json:"duration_hrs"`
}
