package mapper

import (
	"github.com/rinatkh/classroom13/internal/httpapi/dto"
	"github.com/rinatkh/classroom13/internal/note"
)

func NoteToResponse(value note.Note) dto.NoteResponse {
	return dto.NoteResponse{ID: value.ID, Title: value.Title, Text: value.Text}
}

func NotesToResponse(values []note.Note) []dto.NoteResponse {
	result := make([]dto.NoteResponse, 0, len(values))
	for _, value := range values {
		result = append(result, NoteToResponse(value))
	}
	return result
}
