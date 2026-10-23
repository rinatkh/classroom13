package note

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrNotFound   = errors.New("note not found")
	ErrEmptyTitle = errors.New("title must not be empty")
)

type Note struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type Repository interface {
	Create(ctx context.Context, value Note) (Note, error)
	List(ctx context.Context) ([]Note, error)
}

type UseCase struct{ repository Repository }

func NewUseCase(repository Repository) *UseCase { return &UseCase{repository: repository} }

func (uc *UseCase) Create(ctx context.Context, title, text string) (Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Note{}, ErrEmptyTitle
	}
	return uc.repository.Create(ctx, Note{Title: title, Text: strings.TrimSpace(text)})
}

func (uc *UseCase) List(ctx context.Context) ([]Note, error) {
	return uc.repository.List(ctx)
}
