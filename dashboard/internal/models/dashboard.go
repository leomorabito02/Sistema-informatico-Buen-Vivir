package models

type Metrics struct {
	TotalUsers     int `json:"total_users"`
	ActiveCampaigns int `json:"active_campaigns"`
	CompletedCourses int `json:"completed_courses"`
}
