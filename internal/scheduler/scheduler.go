package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/akhil-datla/maildruid/internal/domain/summary"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/infrastructure/smtp"
)

// TaskInfo represents a scheduled task visible in the API.
type TaskInfo struct {
	Interval string   `json:"interval"`
	UserIDs  []string `json:"userIds"`
}

// Scheduler manages periodic email processing tasks.
type Scheduler struct {
	mu         sync.RWMutex
	tasks      map[string][]string            // interval -> []userID
	stopChans  map[string]chan struct{}         // interval -> stop channel
	userSvc    *user.Service
	summarySvc *summary.Service
	mailer     *smtp.Sender
	logger     *slog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// New creates a new scheduler.
func New(userSvc *user.Service, summarySvc *summary.Service, mailer *smtp.Sender, logger *slog.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:      make(map[string][]string),
		stopChans:  make(map[string]chan struct{}),
		userSvc:    userSvc,
		summarySvc: summarySvc,
		mailer:     mailer,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// LoadExisting loads tasks from the database on startup.
func (s *Scheduler) LoadExisting(ctx context.Context) error {
	users, err := s.userSvc.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("loading users for scheduling: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range users {
		if u.UpdateInterval == "" || u.UpdateInterval == "0" {
			continue
		}
		s.tasks[u.UpdateInterval] = append(s.tasks[u.UpdateInterval], u.ID)
		if _, exists := s.stopChans[u.UpdateInterval]; !exists {
			s.startWorker(u.UpdateInterval)
		}
	}

	s.logger.Info("loaded scheduled tasks", "intervals", len(s.tasks))
	return nil
}

// AddTask schedules a new periodic task for a user.
func (s *Scheduler) AddTask(userID, interval string) error {
	if _, err := strconv.Atoi(interval); err != nil {
		return fmt.Errorf("interval must be a number (minutes): %w", err)
	}

	u, err := s.userSvc.GetByID(s.ctx, userID)
	if err != nil {
		return err
	}

	if u.UpdateInterval != "" && u.UpdateInterval != "0" {
		return fmt.Errorf("user already has a scheduled task (interval: %s)", u.UpdateInterval)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[interval] = append(s.tasks[interval], userID)

	if _, exists := s.stopChans[interval]; !exists {
		s.startWorker(interval)
	}

	if err := s.userSvc.UpdateInterval(s.ctx, userID, interval); err != nil {
		return fmt.Errorf("saving interval: %w", err)
	}

	s.logger.Info("task scheduled", "user_id", userID, "interval", interval)
	return nil
}

// RemoveTask removes a user's scheduled task.
func (s *Scheduler) RemoveTask(userID, interval string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userIDs, ok := s.tasks[interval]
	if !ok {
		return fmt.Errorf("no tasks with interval %s", interval)
	}

	s.tasks[interval] = removeFromSlice(userIDs, userID)

	if len(s.tasks[interval]) == 0 {
		s.stopWorker(interval)
		delete(s.tasks, interval)
	}

	if err := s.userSvc.UpdateInterval(s.ctx, userID, "0"); err != nil {
		return fmt.Errorf("clearing interval: %w", err)
	}

	s.logger.Info("task removed", "user_id", userID, "interval", interval)
	return nil
}

// UpdateTask reschedules a task with a new interval.
func (s *Scheduler) UpdateTask(userID, oldInterval, newInterval string) error {
	if _, err := strconv.Atoi(newInterval); err != nil {
		return fmt.Errorf("interval must be a number (minutes): %w", err)
	}

	s.mu.Lock()

	// Remove from old interval
	if userIDs, ok := s.tasks[oldInterval]; ok {
		s.tasks[oldInterval] = removeFromSlice(userIDs, userID)
		if len(s.tasks[oldInterval]) == 0 {
			s.stopWorker(oldInterval)
			delete(s.tasks, oldInterval)
		}
	}

	// Add to new interval
	s.tasks[newInterval] = append(s.tasks[newInterval], userID)
	if _, exists := s.stopChans[newInterval]; !exists {
		s.startWorker(newInterval)
	}

	s.mu.Unlock()

	if err := s.userSvc.UpdateInterval(s.ctx, userID, newInterval); err != nil {
		return fmt.Errorf("updating interval: %w", err)
	}

	s.logger.Info("task updated", "user_id", userID, "old", oldInterval, "new", newInterval)
	return nil
}

// RemoveAllForUser removes all tasks for a given user.
func (s *Scheduler) RemoveAllForUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for interval, userIDs := range s.tasks {
		s.tasks[interval] = removeFromSlice(userIDs, userID)
		if len(s.tasks[interval]) == 0 {
			s.stopWorker(interval)
			delete(s.tasks, interval)
		}
	}
}

// ListTasks returns all currently scheduled tasks.
func (s *Scheduler) ListTasks() []TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]TaskInfo, 0, len(s.tasks))
	for interval, userIDs := range s.tasks {
		ids := make([]string, len(userIDs))
		copy(ids, userIDs)
		tasks = append(tasks, TaskInfo{Interval: interval, UserIDs: ids})
	}
	return tasks
}

// Stop shuts down all scheduled workers.
func (s *Scheduler) Stop() {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	for interval := range s.stopChans {
		s.stopWorker(interval)
	}
	s.logger.Info("scheduler stopped")
}

// startWorker launches a goroutine for the given interval. Must be called with mu held.
func (s *Scheduler) startWorker(interval string) {
	minutes, err := strconv.Atoi(interval)
	if err != nil {
		s.logger.Error("invalid interval", "interval", interval, "error", err)
		return
	}

	stopCh := make(chan struct{})
	s.stopChans[interval] = stopCh

	go func(interval string, duration time.Duration, stop <-chan struct{}) {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				s.processTick(interval)
			}
		}
	}(interval, time.Duration(minutes)*time.Minute, stopCh)
}

// stopWorker signals a worker to stop. Must be called with mu held.
func (s *Scheduler) stopWorker(interval string) {
	if ch, ok := s.stopChans[interval]; ok {
		close(ch)
		delete(s.stopChans, interval)
	}
}

func (s *Scheduler) processTick(interval string) {
	s.mu.RLock()
	userIDs := make([]string, len(s.tasks[interval]))
	copy(userIDs, s.tasks[interval])
	s.mu.RUnlock()

	for _, userID := range userIDs {
		go s.processUser(userID)
	}
}

func (s *Scheduler) processUser(userID string) {
	u, err := s.userSvc.GetByID(s.ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user for processing", "user_id", userID, "error", err)
		return
	}

	result, err := s.summarySvc.Generate(s.ctx, u)
	if err != nil {
		s.logger.Warn("summary generation failed", "user_id", userID, "error", err)
		_ = s.mailer.SendSummary(u.ReceivingEmail, u.Name, u.Tags, "", "", fmt.Sprintf("Summary generation error: %s", err.Error()))
		return
	}

	sendErr := s.mailer.SendSummary(u.ReceivingEmail, u.Name, u.Tags, result.Summary, result.WordCloudPath, "")
	if sendErr != nil {
		s.logger.Error("failed to send summary email", "user_id", userID, "error", sendErr)
	}

	if result.WordCloudPath != "" {
		os.Remove(result.WordCloudPath)
	}

	s.logger.Info("periodic summary sent", "user_id", userID)
}

func removeFromSlice(s []string, item string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}
