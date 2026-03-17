package handlers

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/akhil-datla/maildruid/internal/domain/summary"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/server/middleware"
	"github.com/labstack/echo/v4"
)

// SummaryHandler handles email summarization endpoints.
type SummaryHandler struct {
	userSvc    *user.Service
	summarySvc *summary.Service
	logger     *slog.Logger
}

// NewSummaryHandler creates a new summary handler.
func NewSummaryHandler(userSvc *user.Service, summarySvc *summary.Service, logger *slog.Logger) *SummaryHandler {
	return &SummaryHandler{userSvc: userSvc, summarySvc: summarySvc, logger: logger}
}

// Generate creates a summary and word cloud on demand.
// POST /api/v1/summaries/generate
func (h *SummaryHandler) Generate(c echo.Context) error {
	id := middleware.GetUserID(c)
	u, err := h.userSvc.GetByID(c.Request().Context(), id)
	if errors.Is(err, user.ErrNotFound) {
		return c.JSON(http.StatusNotFound, errResp("user not found"))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to get user"))
	}

	result, err := h.summarySvc.Generate(c.Request().Context(), u)
	if errors.Is(err, user.ErrNoTags) {
		return c.JSON(http.StatusBadRequest, errResp("configure tags before generating a summary"))
	}
	if err != nil {
		if strings.Contains(err.Error(), "no emails found") {
			return c.JSON(http.StatusNotFound, errResp("no emails to summarize"))
		}
		h.logger.Error("summary generation failed", "error", err, "user_id", id)
		return c.JSON(http.StatusInternalServerError, errResp("summary generation failed"))
	}

	resp := SummaryResponse{Summary: result.Summary}

	if result.WordCloudPath != "" {
		data, err := os.ReadFile(result.WordCloudPath)
		if err == nil {
			resp.Image = base64.StdEncoding.EncodeToString(data)
		}
		os.Remove(result.WordCloudPath)
	}

	return c.JSON(http.StatusOK, resp)
}
