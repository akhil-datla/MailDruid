package handlers

import (
	"errors"
	"net/http"

	"github.com/akhil-datla/maildruid/internal/config"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/infrastructure/imap"
	"github.com/akhil-datla/maildruid/internal/server/middleware"
	"github.com/labstack/echo/v4"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	userSvc *user.Service
	authCfg config.AuthConfig
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userSvc *user.Service, authCfg config.AuthConfig) *UserHandler {
	return &UserHandler{userSvc: userSvc, authCfg: authCfg}
}

// Create registers a new user.
// POST /api/v1/users
func (h *UserHandler) Create(c echo.Context) error {
	var req CreateUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	err := h.userSvc.Create(c.Request().Context(), user.CreateInput{
		Name:           req.Name,
		Email:          req.Email,
		ReceivingEmail: req.ReceivingEmail,
		Password:       req.Password,
		Domain:         req.Domain,
		Port:           req.Port,
	})

	if errors.Is(err, user.ErrAlreadyExists) {
		return c.JSON(http.StatusConflict, errResp("user already exists"))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to create user"))
	}

	return c.JSON(http.StatusCreated, msgOK("user created successfully"))
}

// Login authenticates a user and returns a JWT.
// POST /api/v1/auth/login
func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id, err := h.userSvc.Authenticate(c.Request().Context(), req.Email, req.Password)
	if errors.Is(err, user.ErrInvalidPassword) {
		return c.JSON(http.StatusUnauthorized, errResp("invalid credentials"))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("authentication failed"))
	}

	token, err := middleware.GenerateToken(id, []byte(h.authCfg.SigningKey), h.authCfg.TokenExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("could not generate token"))
	}

	return c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// GetProfile returns the authenticated user's profile.
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c echo.Context) error {
	id := middleware.GetUserID(c)
	u, err := h.userSvc.GetByID(c.Request().Context(), id)
	if errors.Is(err, user.ErrNotFound) {
		return c.JSON(http.StatusNotFound, errResp("user not found"))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to get user"))
	}

	return c.JSON(http.StatusOK, u)
}

// Update modifies the authenticated user's profile.
// PATCH /api/v1/users/me
func (h *UserHandler) Update(c echo.Context) error {
	var req UpdateUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	err := h.userSvc.Update(c.Request().Context(), id, user.UpdateInput{
		Name:           req.Name,
		Email:          req.Email,
		ReceivingEmail: req.ReceivingEmail,
		OldPassword:    req.OldPassword,
		NewPassword:    req.NewPassword,
		Domain:         req.Domain,
		Port:           req.Port,
		Folder:         req.Folder,
	})

	if errors.Is(err, user.ErrNotFound) {
		return c.JSON(http.StatusNotFound, errResp("user not found"))
	}
	if errors.Is(err, user.ErrInvalidPassword) {
		return c.JSON(http.StatusBadRequest, errResp("incorrect old password"))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to update user"))
	}

	return c.JSON(http.StatusOK, msgOK("user updated successfully"))
}

// Delete removes the authenticated user.
// DELETE /api/v1/users/me
func (h *UserHandler) Delete(c echo.Context) error {
	id := middleware.GetUserID(c)
	if err := h.userSvc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to delete user"))
	}
	return c.JSON(http.StatusOK, msgOK("user deleted successfully"))
}

// GetFolders lists IMAP folders for the authenticated user.
// GET /api/v1/users/me/folders
func (h *UserHandler) GetFolders(c echo.Context) error {
	id := middleware.GetUserID(c)
	u, err := h.userSvc.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to get user"))
	}

	password, err := h.userSvc.DecryptPassword(u)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to decrypt credentials"))
	}

	im, err := imap.New(u.Email, password, u.Domain, u.Port)
	if err != nil {
		return c.JSON(http.StatusBadGateway, errResp("failed to connect to email server"))
	}

	folders, err := im.GetFolders()
	if err != nil {
		return c.JSON(http.StatusBadGateway, errResp("failed to list folders"))
	}

	return c.JSON(http.StatusOK, folders)
}

// UpdateFolder sets the IMAP folder.
// PATCH /api/v1/users/me/folder
func (h *UserHandler) UpdateFolder(c echo.Context) error {
	var req UpdateFolderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.userSvc.UpdateFolder(c.Request().Context(), id, req.Folder); err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to update folder"))
	}
	return c.JSON(http.StatusOK, msgOK("folder updated"))
}

// UpdateTags sets the email filter tags.
// PUT /api/v1/users/me/tags
func (h *UserHandler) UpdateTags(c echo.Context) error {
	var req UpdateTagsRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.userSvc.UpdateTags(c.Request().Context(), id, req.Tags); err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to update tags"))
	}
	return c.JSON(http.StatusOK, msgOK("tags updated"))
}

// UpdateBlacklist sets the sender blacklist.
// PUT /api/v1/users/me/blacklist
func (h *UserHandler) UpdateBlacklist(c echo.Context) error {
	var req UpdateBlacklistRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.userSvc.UpdateBlackListSenders(c.Request().Context(), id, req.Senders); err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to update blacklist"))
	}
	return c.JSON(http.StatusOK, msgOK("blacklist updated"))
}

// UpdateStartTime sets the email processing start time.
// PATCH /api/v1/users/me/start-time
func (h *UserHandler) UpdateStartTime(c echo.Context) error {
	var req UpdateStartTimeRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.userSvc.UpdateStartTime(c.Request().Context(), id, req.StartTime); err != nil {
		return c.JSON(http.StatusBadRequest, errResp(err.Error()))
	}
	return c.JSON(http.StatusOK, msgOK("start time updated"))
}

// UpdateSummaryCount sets the number of summary sentences.
// PATCH /api/v1/users/me/summary-count
func (h *UserHandler) UpdateSummaryCount(c echo.Context) error {
	var req UpdateSummaryCountRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	id := middleware.GetUserID(c)
	if err := h.userSvc.UpdateSummaryCount(c.Request().Context(), id, req.Count); err != nil {
		return c.JSON(http.StatusInternalServerError, errResp("failed to update summary count"))
	}
	return c.JSON(http.StatusOK, msgOK("summary count updated"))
}
