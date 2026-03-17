package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps go-playground/validator for Echo.
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator.
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate validates a struct.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationErrors(err))
	}
	return nil
}

func formatValidationErrors(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok {
		msg := "validation failed: "
		for i, fe := range ve {
			if i > 0 {
				msg += "; "
			}
			msg += fe.Field() + " " + fe.Tag()
		}
		return msg
	}
	return err.Error()
}

func bindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, errResp("invalid request body"))
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	return nil
}
