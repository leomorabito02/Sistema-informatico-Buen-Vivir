package models

import "time"

// Company represents a company entity in Firestore.
type Company struct {
	ID        string    `firestore:"id,omitempty" json:"id"`
	Name      string    `firestore:"name" json:"name"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
