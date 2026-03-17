package user

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"

	"github.com/akhil-datla/maildruid/internal/infrastructure/encryption"
	"github.com/gofrs/uuid"
)

// Service contains user business logic.
type Service struct {
	repo      Repository
	encryptor *encryption.Service
	logger    *slog.Logger
}

// NewService creates a new user service.
func NewService(repo Repository, enc *encryption.Service, logger *slog.Logger) *Service {
	return &Service{repo: repo, encryptor: enc, logger: logger}
}

// CreateInput holds data needed to register a new user.
type CreateInput struct {
	Name           string
	Email          string
	ReceivingEmail string
	Password       string
	Domain         string
	Port           int
}

// Create registers a new user with an encrypted password.
func (s *Service) Create(ctx context.Context, in CreateInput) error {
	_, err := s.repo.FindByEmail(ctx, in.Email)
	if err == nil {
		return ErrAlreadyExists
	}
	if !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("checking existing user: %w", err)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generating UUID: %w", err)
	}

	encrypted, err := s.encryptor.Encrypt(in.Password)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}

	u := &User{
		ID:             id.String(),
		Name:           in.Name,
		Email:          in.Email,
		ReceivingEmail: in.ReceivingEmail,
		Password:       base64.RawStdEncoding.EncodeToString(encrypted),
		Domain:         in.Domain,
		Port:           in.Port,
		SummaryCount:   5,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	s.logger.Info("user created", "id", u.ID, "email", u.Email)
	return nil
}

// Authenticate validates credentials and returns the user ID.
func (s *Service) Authenticate(ctx context.Context, email, password string) (string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrInvalidPassword
		}
		return "", err
	}

	rawPass, err := base64.RawStdEncoding.DecodeString(u.Password)
	if err != nil {
		return "", fmt.Errorf("decoding password: %w", err)
	}

	decrypted, err := s.encryptor.Decrypt(rawPass)
	if err != nil {
		return "", fmt.Errorf("decrypting password: %w", err)
	}

	if decrypted != password {
		return "", ErrInvalidPassword
	}

	return u.ID, nil
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateInput holds optional fields for updating a user.
type UpdateInput struct {
	Name           *string
	Email          *string
	ReceivingEmail *string
	OldPassword    *string
	NewPassword    *string
	Domain         *string
	Port           *int
	Folder         *string
}

// Update modifies user fields. Password change requires old password verification.
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if in.Name != nil && *in.Name != "" {
		u.Name = *in.Name
	}
	if in.Email != nil && *in.Email != "" {
		u.Email = *in.Email
	}
	if in.ReceivingEmail != nil && *in.ReceivingEmail != "" {
		u.ReceivingEmail = *in.ReceivingEmail
	}
	if in.Domain != nil && *in.Domain != "" {
		u.Domain = *in.Domain
	}
	if in.Port != nil && *in.Port != 0 {
		u.Port = *in.Port
	}
	if in.Folder != nil && *in.Folder != "" {
		u.Folder = *in.Folder
	}

	if in.OldPassword != nil && in.NewPassword != nil && *in.OldPassword != "" && *in.NewPassword != "" {
		rawPass, err := base64.RawStdEncoding.DecodeString(u.Password)
		if err != nil {
			return fmt.Errorf("decoding password: %w", err)
		}
		decrypted, err := s.encryptor.Decrypt(rawPass)
		if err != nil {
			return fmt.Errorf("decrypting password: %w", err)
		}
		if decrypted != *in.OldPassword {
			return ErrInvalidPassword
		}
		encrypted, err := s.encryptor.Encrypt(*in.NewPassword)
		if err != nil {
			return fmt.Errorf("encrypting new password: %w", err)
		}
		u.Password = base64.RawStdEncoding.EncodeToString(encrypted)
	}

	return s.repo.Update(ctx, u)
}

// Delete removes a user by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// UpdateTags sets the email filter tags for a user.
func (s *Service) UpdateTags(ctx context.Context, id string, tags []string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.Tags = tags
	return s.repo.Update(ctx, u)
}

// UpdateBlackListSenders sets the sender blacklist for a user.
func (s *Service) UpdateBlackListSenders(ctx context.Context, id string, senders []string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.BlackListSenders = senders
	return s.repo.Update(ctx, u)
}

// UpdateStartTime sets the email processing start time.
func (s *Service) UpdateStartTime(ctx context.Context, id string, startTime string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if startTime != "" {
		parsed, err := parseTime(startTime)
		if err != nil {
			return fmt.Errorf("parsing start time: %w", err)
		}
		u.StartTime = parsed
	}

	return s.repo.Update(ctx, u)
}

// UpdateSummaryCount sets the number of sentences in summaries.
func (s *Service) UpdateSummaryCount(ctx context.Context, id string, count int) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.SummaryCount = count
	return s.repo.Update(ctx, u)
}

// UpdateFolder sets the IMAP folder to scan.
func (s *Service) UpdateFolder(ctx context.Context, id string, folder string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.Folder = folder
	return s.repo.Update(ctx, u)
}

// UpdateInterval sets the scheduling interval for a user.
func (s *Service) UpdateInterval(ctx context.Context, id string, interval string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	u.UpdateInterval = interval
	return s.repo.Update(ctx, u)
}

// DecryptPassword decrypts and returns the user's IMAP password.
func (s *Service) DecryptPassword(u *User) (string, error) {
	rawPass, err := base64.RawStdEncoding.DecodeString(u.Password)
	if err != nil {
		return "", fmt.Errorf("decoding password: %w", err)
	}
	return s.encryptor.Decrypt(rawPass)
}

// ListAll returns all users.
func (s *Service) ListAll(ctx context.Context) ([]*User, error) {
	return s.repo.ListAll(ctx)
}

// SaveLastUID persists the last processed UID map.
func (s *Service) SaveLastUID(ctx context.Context, u *User, lastUID string) error {
	u.LastUID = lastUID
	return s.repo.Update(ctx, u)
}
