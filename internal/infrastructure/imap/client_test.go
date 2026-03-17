package imap

import (
	"testing"
	"time"
)

func TestFilterEmailsMatchesTags(t *testing.T) {
	emails := []Email{
		{UID: 1, Subject: "Weekly Report from Finance", From: "finance@co.com", Text: "Report body"},
		{UID: 2, Subject: "Lunch plans", From: "friend@co.com", Text: "Let's eat"},
		{UID: 3, Subject: "Q4 Report Summary", From: "ceo@co.com", Text: "Great quarter"},
	}

	filtered := FilterEmails(emails, []string{"report"}, nil, time.Time{})
	if len(filtered) != 2 {
		t.Errorf("expected 2 matching emails, got %d", len(filtered))
	}
}

func TestFilterEmailsExcludesBlacklist(t *testing.T) {
	emails := []Email{
		{UID: 1, Subject: "Weekly Report", From: "spam@co.com", Text: "Spam"},
		{UID: 2, Subject: "Weekly Report", From: "boss@co.com", Text: "Real report"},
	}

	filtered := FilterEmails(emails, []string{"report"}, []string{"spam@co.com"}, time.Time{})
	if len(filtered) != 1 {
		t.Errorf("expected 1, got %d", len(filtered))
	}
	if filtered[0].From != "boss@co.com" {
		t.Errorf("expected boss@co.com, got %s", filtered[0].From)
	}
}

func TestFilterEmailsRespectsStartTime(t *testing.T) {
	cutoff := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	emails := []Email{
		{UID: 1, Subject: "Old Report", From: "a@co.com", Sent: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
		{UID: 2, Subject: "New Report", From: "b@co.com", Sent: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	filtered := FilterEmails(emails, []string{"report"}, nil, cutoff)
	if len(filtered) != 1 {
		t.Errorf("expected 1, got %d", len(filtered))
	}
	if filtered[0].UID != 2 {
		t.Errorf("expected UID 2, got %d", filtered[0].UID)
	}
}

func TestFilterEmailsNoTags(t *testing.T) {
	emails := []Email{
		{UID: 1, Subject: "Hello", From: "a@co.com", Text: "Hi"},
	}

	filtered := FilterEmails(emails, nil, nil, time.Time{})
	if len(filtered) != 0 {
		t.Errorf("expected 0 with nil tags, got %d", len(filtered))
	}
}

func TestFilterEmailsCaseInsensitive(t *testing.T) {
	emails := []Email{
		{UID: 1, Subject: "URGENT REPORT", From: "a@co.com", Text: "Important"},
	}

	filtered := FilterEmails(emails, []string{"urgent"}, nil, time.Time{})
	if len(filtered) != 1 {
		t.Errorf("expected 1 case-insensitive match, got %d", len(filtered))
	}
}

func TestAggregateBody(t *testing.T) {
	emails := []Email{
		{Text: "First email"},
		{Text: "Second email"},
		{Text: ""},
	}

	body := AggregateBody(emails)
	expected := "First email. Second email. "
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestAggregateBodyEmpty(t *testing.T) {
	body := AggregateBody(nil)
	if body != "" {
		t.Errorf("expected empty string, got %q", body)
	}
}
