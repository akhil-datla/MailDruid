package user

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/akhil-datla/maildruid/internal/infrastructure/encryption"
)

func setupTestService(t *testing.T) (*Service, *MemoryRepository) {
	t.Helper()
	repo := NewMemoryRepository()
	enc, err := encryption.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("encryption.New: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewService(repo, enc, logger)
	return svc, repo
}

func TestCreateUser(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	err := svc.Create(ctx, CreateInput{
		Name:           "Test User",
		Email:          "test@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "secret123",
		Domain:         "imap.example.com",
		Port:           993,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify user exists in repo
	users, _ := repo.ListAll(ctx)
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Test User" {
		t.Errorf("expected name 'Test User', got %q", users[0].Name)
	}
	if users[0].Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", users[0].Email)
	}
	if users[0].Password == "" {
		t.Error("password should be encrypted, not empty")
	}
	if users[0].Password == "secret123" {
		t.Error("password should be encrypted, not plaintext")
	}
	if users[0].SummaryCount != 5 {
		t.Errorf("expected default summary count 5, got %d", users[0].SummaryCount)
	}
}

func TestCreateDuplicateUser(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := CreateInput{
		Name:           "Test",
		Email:          "dup@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "secret123",
		Domain:         "imap.example.com",
		Port:           993,
	}

	if err := svc.Create(ctx, input); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	err := svc.Create(ctx, input)
	if err != ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAuthenticateSuccess(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Auth User",
		Email:          "auth@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "mypassword",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, err := svc.Authenticate(ctx, "auth@example.com", "mypassword")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Auth User",
		Email:          "auth2@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "correctpass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	_, err := svc.Authenticate(ctx, "auth2@example.com", "wrongpass")
	if err != ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestAuthenticateNonexistentUser(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "noone@example.com", "pass")
	if err != ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestGetByID(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Get User",
		Email:          "get@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "pass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, _ := svc.Authenticate(ctx, "get@example.com", "pass")
	u, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if u.Name != "Get User" {
		t.Errorf("expected 'Get User', got %q", u.Name)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, "nonexistent-id")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Old Name",
		Email:          "update@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "pass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, _ := svc.Authenticate(ctx, "update@example.com", "pass")

	newName := "New Name"
	newRecv := "new-recv@example.com"
	err := svc.Update(ctx, id, UpdateInput{
		Name:           &newName,
		ReceivingEmail: &newRecv,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if u.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", u.Name)
	}
	if u.ReceivingEmail != "new-recv@example.com" {
		t.Errorf("expected 'new-recv@example.com', got %q", u.ReceivingEmail)
	}
}

func TestUpdatePassword(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Pass User",
		Email:          "passchange@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "oldpass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, _ := svc.Authenticate(ctx, "passchange@example.com", "oldpass")

	oldP := "oldpass"
	newP := "newpass"
	err := svc.Update(ctx, id, UpdateInput{
		OldPassword: &oldP,
		NewPassword: &newP,
	})
	if err != nil {
		t.Fatalf("Update password: %v", err)
	}

	// Old password should no longer work
	_, err = svc.Authenticate(ctx, "passchange@example.com", "oldpass")
	if err != ErrInvalidPassword {
		t.Error("old password should no longer work")
	}

	// New password should work
	_, err = svc.Authenticate(ctx, "passchange@example.com", "newpass")
	if err != nil {
		t.Fatalf("new password should work: %v", err)
	}
}

func TestUpdatePasswordWrongOld(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "User",
		Email:          "wrongold@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "realpass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, _ := svc.Authenticate(ctx, "wrongold@example.com", "realpass")

	oldP := "WRONG"
	newP := "newpass"
	err := svc.Update(ctx, id, UpdateInput{
		OldPassword: &oldP,
		NewPassword: &newP,
	})
	if err != ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	svc, repo := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name:           "Delete Me",
		Email:          "delete@example.com",
		ReceivingEmail: "recv@example.com",
		Password:       "pass",
		Domain:         "imap.example.com",
		Port:           993,
	})

	id, _ := svc.Authenticate(ctx, "delete@example.com", "pass")

	err := svc.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	users, _ := repo.ListAll(ctx)
	if len(users) != 0 {
		t.Errorf("expected 0 users after delete, got %d", len(users))
	}
}

func TestUpdateTags(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "Tag User", Email: "tags@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "tags@example.com", "p")

	err := svc.UpdateTags(ctx, id, []string{"report", "weekly"})
	if err != nil {
		t.Fatalf("UpdateTags: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if len(u.Tags) != 2 || u.Tags[0] != "report" || u.Tags[1] != "weekly" {
		t.Errorf("expected [report weekly], got %v", u.Tags)
	}
}

func TestUpdateBlackListSenders(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "BL User", Email: "bl@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "bl@example.com", "p")

	err := svc.UpdateBlackListSenders(ctx, id, []string{"spam@co.com"})
	if err != nil {
		t.Fatalf("UpdateBlackListSenders: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if len(u.BlackListSenders) != 1 || u.BlackListSenders[0] != "spam@co.com" {
		t.Errorf("expected [spam@co.com], got %v", u.BlackListSenders)
	}
}

func TestUpdateFolder(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "Folder User", Email: "folder@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "folder@example.com", "p")

	err := svc.UpdateFolder(ctx, id, "INBOX")
	if err != nil {
		t.Fatalf("UpdateFolder: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if u.Folder != "INBOX" {
		t.Errorf("expected 'INBOX', got %q", u.Folder)
	}
}

func TestUpdateStartTime(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "ST User", Email: "st@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "st@example.com", "p")

	err := svc.UpdateStartTime(ctx, id, "2025-01-15T00:00:00Z")
	if err != nil {
		t.Fatalf("UpdateStartTime: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if u.StartTime.Year() != 2025 || u.StartTime.Month() != 1 || u.StartTime.Day() != 15 {
		t.Errorf("unexpected start time: %v", u.StartTime)
	}
}

func TestUpdateStartTimeInvalid(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "ST2", Email: "st2@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "st2@example.com", "p")

	err := svc.UpdateStartTime(ctx, id, "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestUpdateSummaryCount(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "SC User", Email: "sc@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "sc@example.com", "p")

	err := svc.UpdateSummaryCount(ctx, id, 10)
	if err != nil {
		t.Fatalf("UpdateSummaryCount: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if u.SummaryCount != 10 {
		t.Errorf("expected 10, got %d", u.SummaryCount)
	}
}

func TestUpdateInterval(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "Int User", Email: "int@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "int@example.com", "p")

	err := svc.UpdateInterval(ctx, id, "30")
	if err != nil {
		t.Fatalf("UpdateInterval: %v", err)
	}

	u, _ := svc.GetByID(ctx, id)
	if u.UpdateInterval != "30" {
		t.Errorf("expected '30', got %q", u.UpdateInterval)
	}
}

func TestDecryptPassword(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "Dec User", Email: "dec@example.com", ReceivingEmail: "r@ex.com",
		Password: "my-secret-pass", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "dec@example.com", "my-secret-pass")

	u, _ := svc.GetByID(ctx, id)
	pass, err := svc.DecryptPassword(u)
	if err != nil {
		t.Fatalf("DecryptPassword: %v", err)
	}
	if pass != "my-secret-pass" {
		t.Errorf("expected 'my-secret-pass', got %q", pass)
	}
}

func TestSaveLastUID(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateInput{
		Name: "UID User", Email: "uid@example.com", ReceivingEmail: "r@ex.com",
		Password: "p", Domain: "imap.ex.com", Port: 993,
	})
	id, _ := svc.Authenticate(ctx, "uid@example.com", "p")

	u, _ := svc.GetByID(ctx, id)
	err := svc.SaveLastUID(ctx, u, `{"report":500}`)
	if err != nil {
		t.Fatalf("SaveLastUID: %v", err)
	}

	u2, _ := svc.GetByID(ctx, id)
	if u2.LastUID != `{"report":500}` {
		t.Errorf("expected lastUID, got %q", u2.LastUID)
	}
}

func TestListAll(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	for i, email := range []string{"a@ex.com", "b@ex.com", "c@ex.com"} {
		svc.Create(ctx, CreateInput{
			Name: "User", Email: email, ReceivingEmail: "r@ex.com",
			Password: "p", Domain: "imap.ex.com", Port: 993 + i,
		})
	}

	users, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}
