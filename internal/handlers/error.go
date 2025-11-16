package handlers

import (
	"errors"
	"fmt"
	"net/http"
	domainerrors "pr-review/internal/errors"
	"pr-review/internal/handlers/v1/dto"
	"strings"

	"github.com/labstack/echo/v4"
)

func ConvertDomainError(c echo.Context, err error, description string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domainerrors.NotFoundError):
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "NOT_FOUND",
				Message: fmt.Sprintf("%s: %v", description, err),
			},
		})

	case errors.Is(err, domainerrors.AlreadyExistsError):
		msg := err.Error()
		var code string
		if strings.Contains(msg, "PR id already exists") {
			code = "PR_EXISTS"
		} else if strings.Contains(msg, "team_name already exists") {
			code = "TEAM_EXISTS"
		} else {
			code = "ALREADY_EXISTS"
		}

		statusCode := http.StatusConflict
		if code == "TEAM_EXISTS" {
			statusCode = http.StatusBadRequest
		}

		return c.JSON(statusCode, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    code,
				Message: extractMessage(err),
			},
		})

	case errors.Is(err, domainerrors.BusinessLogicError):
		msg := err.Error()
		var code string
		if strings.Contains(msg, "cannot reassign on merged PR") {
			code = "PR_MERGED"
		} else if strings.Contains(msg, "reviewer is not assigned") {
			code = "NOT_ASSIGNED"
		} else if strings.Contains(msg, "no active replacement candidate") {
			code = "NO_CANDIDATE"
		} else {
			code = "BUSINESS_LOGIC_ERROR"
		}

		return c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    code,
				Message: extractMessage(err),
			},
		})

	default:
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "NOT_FOUND",
				Message: "resource not found",
			},
		})
	}
}

func extractMessage(err error) string {
	msg := err.Error()
	if idx := len("business logic error: "); len(msg) > idx && msg[:idx] == "business logic error: " {
		return msg[idx:]
	}
	if idx := len("already exists: "); len(msg) > idx && msg[:idx] == "already exists: " {
		return msg[idx:]
	}
	if idx := len("not found: "); len(msg) > idx && msg[:idx] == "not found: " {
		return msg[idx:]
	}
	return msg
}
