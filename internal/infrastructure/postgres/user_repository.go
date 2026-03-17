package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/akhil-datla/maildruid/internal/domain/user"
	"gorm.io/gorm"
)

// UserRepository implements user.Repository with PostgreSQL.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db.GORM()}
}

func (r *UserRepository) Create(_ context.Context, u *user.User) error {
	if err := r.db.Create(u).Error; err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(_ context.Context, id string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	var u user.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Update(_ context.Context, u *user.User) error {
	if err := r.db.Save(u).Error; err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(_ context.Context, id string) error {
	if err := r.db.Where("id = ?", id).Delete(&user.User{}).Error; err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

func (r *UserRepository) ListAll(_ context.Context) ([]*user.User, error) {
	var users []*user.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return users, nil
}
