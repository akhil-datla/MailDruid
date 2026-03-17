package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/akhil-datla/maildruid/internal/domain/user"
	imapClient "github.com/akhil-datla/maildruid/internal/infrastructure/imap"
	"github.com/akhil-datla/maildruid/internal/infrastructure/wordcloud"
)

// Result holds the output of an email summarization.
type Result struct {
	Summary       string
	WordCloudPath string
}

// Service orchestrates the email summarization pipeline.
type Service struct {
	userSvc   *user.Service
	generator *wordcloud.Generator
	logger    *slog.Logger
}

// NewService creates a new summary service.
func NewService(userSvc *user.Service, gen *wordcloud.Generator, logger *slog.Logger) *Service {
	return &Service{userSvc: userSvc, generator: gen, logger: logger}
}

// Generate runs the full summarization pipeline for a user.
func (s *Service) Generate(ctx context.Context, u *user.User) (*Result, error) {
	if len(u.Tags) == 0 {
		return nil, user.ErrNoTags
	}

	password, err := s.userSvc.DecryptPassword(u)
	if err != nil {
		return nil, fmt.Errorf("decrypting password: %w", err)
	}

	im, err := imapClient.New(u.Email, password, u.Domain, u.Port)
	if err != nil {
		return nil, fmt.Errorf("connecting to IMAP: %w", err)
	}

	if u.Folder != "" {
		if err := im.SelectFolder(u.Folder); err != nil {
			return nil, fmt.Errorf("selecting folder: %w", err)
		}
	}

	// Determine starting UID from last processed state
	uidMap := make(map[string]int)
	if u.LastUID != "" {
		if err := json.Unmarshal([]byte(u.LastUID), &uidMap); err != nil {
			if !strings.Contains(err.Error(), "unexpected end of JSON input") {
				return nil, fmt.Errorf("parsing UID map: %w", err)
			}
		}
	}

	for _, tag := range u.Tags {
		if uidMap[tag] == 0 {
			uidMap[tag] = 1
		}
	}

	lowestUID := uidMap[u.Tags[0]]
	for _, tag := range u.Tags {
		if uidMap[tag] < lowestUID {
			lowestUID = uidMap[tag]
		}
	}

	// Check for latest UID to avoid overshooting
	latestUIDs, err := im.GetUIDs("999999999:*")
	if err != nil {
		return nil, fmt.Errorf("getting latest UIDs: %w", err)
	}
	if len(latestUIDs) > 0 && latestUIDs[len(latestUIDs)-1] < lowestUID {
		lowestUID = latestUIDs[len(latestUIDs)-1]
	}

	emails, uidList, err := im.GetEmails(u.Folder, lowestUID)
	if err != nil {
		return nil, fmt.Errorf("fetching emails: %w", err)
	}

	if len(emails) == 0 {
		return nil, fmt.Errorf("no emails found")
	}

	// Update UID tracking
	if len(uidList) > 0 {
		newUID := uidList[len(uidList)-1]
		for _, tag := range u.Tags {
			uidMap[tag] = newUID
		}
		uidBytes, err := json.Marshal(uidMap)
		if err == nil {
			_ = s.userSvc.SaveLastUID(ctx, u, string(uidBytes))
		}
	}

	filtered := imapClient.FilterEmails(emails, u.Tags, u.BlackListSenders, u.StartTime)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no emails found with tags: %v", u.Tags)
	}

	body := imapClient.AggregateBody(filtered)
	if body == "" {
		return nil, fmt.Errorf("no email content to summarize")
	}

	summarized := s.generator.Summarize(body, u.SummaryCount)
	keywords := s.generator.ExtractKeywords(summarized)

	wordCloudPath, err := s.generator.GenerateWordCloud(keywords)
	if err != nil {
		s.logger.Warn("word cloud generation failed", "error", err)
		// Non-fatal: return summary without word cloud
	}

	s.logger.Info("summary generated", "user", u.ID, "emails_processed", len(filtered))

	return &Result{
		Summary:       summarized,
		WordCloudPath: wordCloudPath,
	}, nil
}
