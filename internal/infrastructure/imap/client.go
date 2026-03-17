package imap

import (
	"fmt"
	"strings"
	"time"

	goiMAP "github.com/BrianLeishman/go-imap"
)

// Client wraps IMAP operations.
type Client struct {
	dialer *goiMAP.Dialer
}

// New creates a new IMAP client connection.
func New(email, password, domain string, port int) (*Client, error) {
	im, err := goiMAP.New(email, password, domain, port)
	if err != nil {
		return nil, fmt.Errorf("connecting to IMAP: %w", err)
	}
	return &Client{dialer: im}, nil
}

// GetFolders lists all available IMAP folders.
func (c *Client) GetFolders() ([]string, error) {
	folders, err := c.dialer.GetFolders()
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}
	return folders, nil
}

// SelectFolder switches to the specified folder.
func (c *Client) SelectFolder(folder string) error {
	if err := c.dialer.SelectFolder(folder); err != nil {
		return fmt.Errorf("selecting folder %q: %w", folder, err)
	}
	return nil
}

// GetUIDs returns UIDs matching the given range (e.g., "1:*").
func (c *Client) GetUIDs(uidRange string) ([]int, error) {
	uids, err := c.dialer.GetUIDs(uidRange)
	if err != nil {
		return nil, fmt.Errorf("getting UIDs: %w", err)
	}
	return uids, nil
}

// Email represents a simplified email message.
type Email struct {
	UID     int
	Subject string
	From    string
	Text    string
	Sent    time.Time
}

// GetEmails retrieves emails starting from the given UID.
func (c *Client) GetEmails(folder string, fromUID int) ([]Email, []int, error) {
	uids, err := c.dialer.GetUIDs(fmt.Sprintf("%d:*", fromUID))
	if err != nil {
		return nil, nil, fmt.Errorf("getting UIDs: %w", err)
	}

	if len(uids) == 0 {
		return nil, nil, nil
	}

	rawEmails, err := c.dialer.GetEmails(uids...)
	if err != nil {
		return nil, nil, fmt.Errorf("getting emails: %w", err)
	}

	emails := make([]Email, 0, len(rawEmails))
	for _, raw := range rawEmails {
		sender := extractSender(raw.From)
		emails = append(emails, Email{
			UID:     raw.UID,
			Subject: raw.Subject,
			From:    sender,
			Text:    raw.Text,
			Sent:    raw.Sent,
		})
	}

	return emails, uids, nil
}

// FilterEmails filters emails by tags, blacklisted senders, and start time.
func FilterEmails(emails []Email, tags, blacklist []string, startTime time.Time) []Email {
	var filtered []Email
	for _, e := range emails {
		if !matchesTags(e.Subject, tags) {
			continue
		}
		if isBlacklisted(e.From, blacklist) {
			continue
		}
		if !startTime.IsZero() && e.Sent.Before(startTime) {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

// AggregateBody concatenates email bodies into a single string.
func AggregateBody(emails []Email) string {
	var b strings.Builder
	for _, e := range emails {
		if e.Text != "" {
			b.WriteString(e.Text)
			b.WriteString(". ")
		}
	}
	return b.String()
}

func matchesTags(subject string, tags []string) bool {
	lower := strings.ToLower(subject)
	for _, tag := range tags {
		if strings.Contains(lower, strings.ToLower(tag)) {
			return true
		}
	}
	return false
}

func isBlacklisted(sender string, blacklist []string) bool {
	for _, blocked := range blacklist {
		if sender == blocked {
			return true
		}
	}
	return false
}

func extractSender(addresses goiMAP.EmailAddresses) string {
	for addr := range addresses {
		return addr
	}
	return ""
}
