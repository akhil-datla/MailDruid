package handlers

// Request types for JSON body binding with validation.

type CreateUserRequest struct {
	Name           string `json:"name" validate:"required,min=2,max=100"`
	Email          string `json:"email" validate:"required,email"`
	ReceivingEmail string `json:"receivingEmail" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	Domain         string `json:"domain" validate:"required"`
	Port           int    `json:"port" validate:"required,min=1,max=65535"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name           *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email          *string `json:"email,omitempty" validate:"omitempty,email"`
	ReceivingEmail *string `json:"receivingEmail,omitempty" validate:"omitempty,email"`
	OldPassword    *string `json:"oldPassword,omitempty"`
	NewPassword    *string `json:"newPassword,omitempty" validate:"omitempty,min=6"`
	Domain         *string `json:"domain,omitempty"`
	Port           *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Folder         *string `json:"folder,omitempty"`
}

type UpdateTagsRequest struct {
	Tags []string `json:"tags" validate:"required,min=1,dive,required"`
}

type UpdateBlacklistRequest struct {
	Senders []string `json:"senders" validate:"required"`
}

type UpdateStartTimeRequest struct {
	StartTime string `json:"startTime" validate:"required"`
}

type UpdateSummaryCountRequest struct {
	Count int `json:"count" validate:"required,min=1,max=100"`
}

type UpdateFolderRequest struct {
	Folder string `json:"folder" validate:"required"`
}

type ScheduleTaskRequest struct {
	Interval string `json:"interval" validate:"required"`
}

type UpdateTaskRequest struct {
	OldInterval string `json:"oldInterval" validate:"required"`
	NewInterval string `json:"newInterval" validate:"required"`
}
