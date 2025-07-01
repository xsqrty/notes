package dtoadapter

import (
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/dto"
	"time"
)

func NoteRequestDtoToCreateData(request *dto.NoteRequest) *note.CreateData {
	return &note.CreateData{
		Name: request.Name,
		Text: request.Text,
	}
}

func NoteRequestDtoToUpdateData(id uuid.UUID, request *dto.NoteRequest) *note.UpdateData {
	return &note.UpdateData{
		ID:   id,
		Name: request.Name,
		Text: request.Text,
	}
}

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
