package note

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	items  map[int64]Note
	nextID int64
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: make(map[int64]Note), nextID: 1}
}

func (r *fakeRepository) Create(_ context.Context, value Note) (Note, error) {
	value.ID = r.nextID
	r.nextID++
	r.items[value.ID] = value
	return value, nil
}
func (r *fakeRepository) ListByOwner(_ context.Context, ownerID int64) ([]Note, error) {
	result := make([]Note, 0)
	for _, value := range r.items {
		if value.OwnerID == ownerID {
			result = append(result, value)
		}
	}
	return result, nil
}
func (r *fakeRepository) Get(_ context.Context, id int64) (Note, error) {
	value, ok := r.items[id]
	if !ok {
		return Note{}, ErrNotFound
	}
	return value, nil
}
func (r *fakeRepository) Update(_ context.Context, value Note) (Note, error) {
	r.items[value.ID] = value
	return value, nil
}
func (r *fakeRepository) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

func TestOwnerCanUpdateAndDelete(t *testing.T) {
	repository := newFakeRepository()
	repository.items[1] = Note{ID: 1, OwnerID: 7, Title: "before"}
	uc := NewUseCase(repository)

	updated, err := uc.Update(context.Background(), 7, 1, UpdateInput{Title: "after"})
	if err != nil || updated.Title != "after" {
		t.Fatalf("Update() = %+v, %v", updated, err)
	}
	if err := uc.Delete(context.Background(), 7, 1); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestAnotherUserGetsForbidden(t *testing.T) {
	repository := newFakeRepository()
	repository.items[1] = Note{ID: 1, OwnerID: 7, Title: "private"}
	uc := NewUseCase(repository)

	_, err := uc.Get(context.Background(), 8, 1)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Get() error = %v, want ErrForbidden", err)
	}
	if err := uc.Delete(context.Background(), 8, 1); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Delete() error = %v, want ErrForbidden", err)
	}
	if _, ok := repository.items[1]; !ok {
		t.Fatal("foreign note was deleted")
	}
}

func TestListContainsOnlyOwnerNotes(t *testing.T) {
	repository := newFakeRepository()
	repository.items[1] = Note{ID: 1, OwnerID: 7, Title: "mine"}
	repository.items[2] = Note{ID: 2, OwnerID: 8, Title: "foreign"}
	uc := NewUseCase(repository)

	result, err := uc.List(context.Background(), 7)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result) != 1 || result[0].OwnerID != 7 {
		t.Fatalf("List() = %+v", result)
	}
}
