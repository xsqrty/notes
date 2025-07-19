package dto

import (
	"time"

	"github.com/google/uuid"
)

// NoteRequest represents the data required to create or update a note.
type NoteRequest struct {
	Name string `json:"name" validate:"required,min=5,max=200"`
	Text string `json:"text" validate:"required,min=5,max=2000"`
}

// NoteResponse represents the response structure for a note, including metadata and ownership details.
type NoteResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Text      string    `json:"text"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

// NoteSearchResponse represents the response for a note search query containing the total rows and list of notes.
type NoteSearchResponse struct {
	TotalRows uint64          `json:"total_rows"`
	Rows      []*NoteResponse `json:"rows"`
}
