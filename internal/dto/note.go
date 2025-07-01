package dto

import (
	"github.com/google/uuid"
	"time"
)

type NoteRequest struct {
	Name string `json:"name" validate:"required,min=5,max=200"`
	Text string `json:"text" validate:"required,min=5,max=2000"`
}

type NoteResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

type NoteSearchResponse struct {
	TotalRows uint64          `json:"total_rows"`
	Rows      []*NoteResponse `json:"rows"`
}
