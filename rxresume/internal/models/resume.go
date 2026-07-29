package models

type Resume struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	Title     string   `json:"title"`
	Skills    []string `json:"skills"`
	Education string   `json:"education"`
}
