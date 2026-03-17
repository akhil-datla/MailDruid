package handlers

import (
	"net/http"

	"github.com/akhil-datla/maildruid/internal/scheduler"
	"github.com/akhil-datla/maildruid/internal/server/middleware"
	"github.com/labstack/echo/v4"
)

// ScheduleHandler handles task scheduling endpoints.
type ScheduleHandler struct {
	scheduler *scheduler.Scheduler
}

// NewScheduleHandler creates a new schedule handler.
func NewScheduleHandler(sched *scheduler.Scheduler) *ScheduleHandler {
	return &ScheduleHandler{scheduler: sched}
}

// Create schedules a new periodic task.
// POST /api/v1/schedules
func (h *ScheduleHandler) Create(c echo.Context) error {
	var req ScheduleTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.scheduler.AddTask(id, req.Interval); err != nil {
		return c.JSON(http.StatusBadRequest, errResp(err.Error()))
	}
	return c.JSON(http.StatusCreated, msgOK("task scheduled"))
}

// Update reschedules a task with a new interval.
// PATCH /api/v1/schedules
func (h *ScheduleHandler) Update(c echo.Context) error {
	var req UpdateTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.scheduler.UpdateTask(id, req.OldInterval, req.NewInterval); err != nil {
		return c.JSON(http.StatusBadRequest, errResp(err.Error()))
	}
	return c.JSON(http.StatusOK, msgOK("task updated"))
}

// Delete removes a scheduled task.
// DELETE /api/v1/schedules
func (h *ScheduleHandler) Delete(c echo.Context) error {
	var req ScheduleTaskRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.scheduler.RemoveTask(id, req.Interval); err != nil {
		return c.JSON(http.StatusBadRequest, errResp(err.Error()))
	}
	return c.JSON(http.StatusOK, msgOK("task deleted"))
}

// List returns all scheduled tasks.
// GET /api/v1/schedules
func (h *ScheduleHandler) List(c echo.Context) error {
	tasks := h.scheduler.ListTasks()
	return c.JSON(http.StatusOK, tasks)
}
