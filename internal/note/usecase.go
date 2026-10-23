package note

import (
	"context"
	"strings"
)

type UseCase struct {
	repository Repository
}

func NewUseCase(repository Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (uc *UseCase) Create(ctx context.Context, userID int64, input CreateInput) (Note, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Note{}, ErrEmptyTitle
	}
	return uc.repository.Create(ctx, Note{
		OwnerID: userID,
		Title:   title,
		Text:    strings.TrimSpace(input.Text),
	})
}

func (uc *UseCase) List(ctx context.Context, userID int64) ([]Note, error) {
	return uc.repository.ListByOwner(ctx, userID)
}

func (uc *UseCase) Get(ctx context.Context, userID, noteID int64) (Note, error) {
	value, err := uc.repository.Get(ctx, noteID)
	if err != nil {
		return Note{}, err
	}
	if value.OwnerID != userID {
		return Note{}, ErrForbidden
	}
	return value, nil
}

func (uc *UseCase) Update(ctx context.Context, userID, noteID int64, input UpdateInput) (Note, error) {
	value, err := uc.repository.Get(ctx, noteID)
	if err != nil {
		return Note{}, err
	}
	if value.OwnerID != userID {
		return Note{}, ErrForbidden
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Note{}, ErrEmptyTitle
	}
	value.Title = title
	value.Text = strings.TrimSpace(input.Text)
	return uc.repository.Update(ctx, value)
}

func (uc *UseCase) Delete(ctx context.Context, userID, noteID int64) error {
	value, err := uc.repository.Get(ctx, noteID)
	if err != nil {
		return err
	}
	if value.OwnerID != userID {
		return ErrForbidden
	}
	return uc.repository.Delete(ctx, noteID)
}
