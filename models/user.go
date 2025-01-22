package models

import "time"

type User struct {
	ID           int        `json:"id" csv:"id"`
	FirstName    string     `json:"first_name" csv:"first_name"`
	LastName     string     `json:"last_name" csv:"last_name"`
	EmailAddress string     `json:"email_address" csv:"email_address"`
	CreatedAt    time.Time  `json:"created_at" csv:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at" csv:"deleted_at"`
	MergedAt     *time.Time `json:"merged_at" csv:"merged_at"`
	ParentUserID *int       `json:"parent_user_id" csv:"parent_user_id"`
}
