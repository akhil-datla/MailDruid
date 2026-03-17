package user

import (
	"context"
	"sync"
)

// MemoryRepository is an in-memory user repository for testing.
type MemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{users: make(map[string]*User)}
}

func (r *MemoryRepository) Create(_ context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *MemoryRepository) FindByEmail(_ context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) Update(_ context.Context, u *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *MemoryRepository) ListAll(_ context.Context) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		cp := *u
		result = append(result, &cp)
	}
	return result, nil
}
