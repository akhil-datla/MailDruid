package user

import (
	"errors"
	"time"

	"github.com/lib/pq"
)

// Domain errors.
var (
	ErrNotFound        = errors.New("user not found")
	ErrAlreadyExists   = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid credentials")
	ErrNoTags          = errors.New("no tags configured")
)

// User represents a registered MailDruid user.
type User struct {
	ID               string         `json:"id" gorm:"primaryKey"`
	Name             string         `json:"name"`
	Email            string         `json:"email" gorm:"uniqueIndex"`
	ReceivingEmail   string         `json:"receivingEmail"`
	Password         string         `json:"-"`
	Domain           string         `json:"domain"`
	Port             int            `json:"port"`
	Folder           string         `json:"folder"`
	Tags             pq.StringArray `json:"tags" gorm:"type:text[]"`
	BlackListSenders pq.StringArray `json:"blackListSenders" gorm:"type:text[]"`
	StartTime        time.Time      `json:"startTime"`
	SummaryCount     int            `json:"summaryCount"`
	LastUID          string         `json:"-"`
	UpdateInterval   string         `json:"updateInterval"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}
