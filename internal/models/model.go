package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	ServiceName string       `db:"service_name" json:"service_name"`
	Price       int          `db:"price" json:"price"`
	StartDate   time.Time    `db:"start_date" json:"start_date"`
	EndDate     sql.NullTime `db:"end_date" json:"end_date"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	UserID      string       `db:"user_id" json:"user_id"`
}

type CreateSubRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}
