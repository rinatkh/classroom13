package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/rinatkh/classroom13/before/internal/note"
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

func (r *NoteRepository) List(_ context.Context) ([]note.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]note.Note, 0, len(r.items))
	for _, value := range r.items {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
