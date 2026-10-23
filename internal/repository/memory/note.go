package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/rinatkh/classroom13/internal/note"
)

type NoteRepository struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]note.Note
}

func NewNoteRepository() *NoteRepository {
	return &NoteRepository{nextID: 1, items: make(map[int64]note.Note)}
}

func (r *NoteRepository) Create(_ context.Context, value note.Note) (note.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value.ID = r.nextID
	r.nextID++
	r.items[value.ID] = value
	return value, nil
}

func (r *NoteRepository) ListByOwner(_ context.Context, ownerID int64) ([]note.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]note.Note, 0)
	for _, value := range r.items {
		if value.OwnerID == ownerID {
			result = append(result, value)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *NoteRepository) Get(_ context.Context, id int64) (note.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.items[id]
	if !ok {
		return note.Note{}, note.ErrNotFound
	}
	return value, nil
}

func (r *NoteRepository) Update(_ context.Context, value note.Note) (note.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[value.ID]; !ok {
		return note.Note{}, note.ErrNotFound
	}
	r.items[value.ID] = value
	return value, nil
}

func (r *NoteRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return note.ErrNotFound
	}
	delete(r.items, id)
	return nil
}
