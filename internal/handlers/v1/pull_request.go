package v1

import (
	"net/http"
	"pr-review/internal/handlers"
	"pr-review/internal/handlers/v1/dto"

	"github.com/labstack/echo/v4"
)

func (a *API) createPullRequest(c echo.Context) error {
	var req dto.CreatePullRequestRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	pr, err := a.service.CreatePullRequest(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "create pull request")
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"pr": dto.FromModelPullRequest(pr),
	})
}

func (a *API) mergePullRequest(c echo.Context) error {
	var req dto.MergePullRequestRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	pr, err := a.service.MergePullRequest(ctx, req.PullRequestID)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "merge pull request")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"pr": dto.FromModelPullRequest(pr),
	})
}

func (a *API) reassignPullRequest(c echo.Context) error {
	var req dto.ReassignPullRequestRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	pr, newReviewerID, err := a.service.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "reassign reviewer")
	}

	return c.JSON(http.StatusOK, dto.ReassignPullRequestResponse{
		PR:         dto.FromModelPullRequest(pr),
		ReplacedBy: newReviewerID,
	})
}

func (a *API) registerPullRequestHandlers(group *echo.Group) {
	group.POST("/pullRequest/create", a.createPullRequest)
	group.POST("/pullRequest/merge", a.mergePullRequest)
	group.POST("/pullRequest/reassign", a.reassignPullRequest)
}
