package v1

import (
	"net/http"

	"pr-review/internal/handlers"
	"pr-review/internal/handlers/v1/dto"

	"github.com/labstack/echo/v4"
)

func (a *API) registerUserHandlers(group *echo.Group) {
	group.POST("/users/setIsActive", a.setIsActive)
	group.GET("/users/getReview", a.getUserReviews)
}

func (a *API) setIsActive(c echo.Context) error {
	var req dto.SetIsActiveRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID := req.UserID

	ctx := c.Request().Context()
	user, err := a.service.SetUserIsActive(ctx, userID, req.IsActive)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "set user active")
	}

	teamName := ""
	if user.TeamID > 0 {
		team, err := a.service.GetTeamByID(ctx, user.TeamID)
		if err == nil && team != nil {
			teamName = team.Name
		}
	}

	resp := dto.UserResponse{
		UserID:   user.ID,
		Username: user.Name,
		TeamName: teamName,
		IsActive: user.IsActive,
	}

	return c.JSON(http.StatusOK, map[string]any{
		"user": resp,
	})
}

func (a *API) getUserReviews(c echo.Context) error {
	userIDStr := c.QueryParam("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	ctx := c.Request().Context()
	prs, err := a.service.ListUserReviews(ctx, userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := dto.GetReviewResponse{
		UserID:       userIDStr,
		PullRequests: make([]dto.PullRequestShortResponse, 0, len(prs)),
	}

	for _, pr := range prs {
		if pr == nil {
			continue
		}

		resp.PullRequests = append(resp.PullRequests, dto.PullRequestShortResponse{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.Title,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
