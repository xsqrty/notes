package dtoadapter

import (
	"time"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/dto"
)

// NoteRequestDtoToCreateData converts a NoteRequest DTO to a CreateData model for note creation.
func NoteRequestDtoToCreateData(request *dto.NoteRequest) *note.CreateData {
	return &note.CreateData{
		Name: request.Name,
		Text: request.Text,
	}
}

// NoteRequestDtoToUpdateData converts a NoteRequest DTO and ID into an UpdateData structure for note updates.
func NoteRequestDtoToUpdateData(id uuid.UUID, request *dto.NoteRequest) *note.UpdateData {
	return &note.UpdateData{
		ID:   id,
		Name: request.Name,
		Text: request.Text,
	}
}

// NoteToResponseDto converts a note.Note model to a dto.NoteResponse transferring specific fields.
func NoteToResponseDto(note *note.Note) *dto.NoteResponse {
	return &dto.NoteResponse{
		ID:        note.ID,
		Name:      note.Name,
		Text:      note.Text,
		UserID:    note.UserId,
		CreatedAt: note.CreatedAt,
		UpdatedAt: time.Time(note.UpdatedAt),
	}
}

// NoteSearchToResponseDto converts a search result containing notes into a NoteSearchResponse DTO.
// It iterates over the rows in the search result, converting each note into a NoteResponse DTO using NoteToResponseDto.
// Returns a NoteSearchResponse with the total rows and the converted rows.
func NoteSearchToResponseDto(res *search.Result[note.Note]) *dto.NoteSearchResponse {
	rows := make([]*dto.NoteResponse, len(res.Rows))
	for i := range res.Rows {
		rows[i] = NoteToResponseDto(res.Rows[i])
	}

	return &dto.NoteSearchResponse{
		TotalRows: res.TotalRows,
		Rows:      rows,
	}
}
