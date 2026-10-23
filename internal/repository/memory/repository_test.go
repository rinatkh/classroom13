package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/rinatkh/classroom13/internal/note"
)

func TestNoteRepositoryConcurrentCreate(t *testing.T) {
	repository := NewNoteRepository()
	const workers = 50
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := repository.Create(context.Background(), note.Note{OwnerID: 1, Title: "note"}); err != nil {
				t.Errorf("Create() error = %v", err)
			}
		}()
	}
	wg.Wait()
	items, err := repository.ListByOwner(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListByOwner() error = %v", err)
	}
	if len(items) != workers {
		t.Fatalf("len(items) = %d, want %d", len(items), workers)
	}
}
