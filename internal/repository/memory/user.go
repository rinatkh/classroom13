// Package memory содержит потокобезопасные in-memory реализации repository.
// Это infrastructure detail: usecase видит только интерфейсы user.Repository
// и note.Repository.
package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/rinatkh/classroom13/internal/user"
)

type UserRepository struct {
	mu      sync.RWMutex
	nextID  int64
	byEmail map[string]user.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{nextID: 1, byEmail: make(map[string]user.User)}
}

func (r *UserRepository) Create(_ context.Context, value user.User) (user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := strings.ToLower(value.Email)
	if _, exists := r.byEmail[key]; exists {
		return user.User{}, user.ErrAlreadyExists
	}
	value.ID = r.nextID
	r.nextID++
	r.byEmail[key] = value
	return value, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return value, nil
}
