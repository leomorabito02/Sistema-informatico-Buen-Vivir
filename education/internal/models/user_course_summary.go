package models

type UserCourseSummary struct {
	UserID  string   `json:"user_id"`
	Courses []string `json:"courses"`
}
