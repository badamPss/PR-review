package v1

import (
	"net/http"
	"pr-review/internal/handlers"
	"pr-review/internal/handlers/v1/dto"

	"github.com/labstack/echo/v4"
)

func (a *API) registerTeamHandlers(group *echo.Group) {
	group.POST("/team/add", a.createTeam)
	group.GET("/team/get", a.getTeam)
	group.POST("/team/deactivateMembers", a.deactivateTeamMembers)
}

func (a *API) createTeam(c echo.Context) error {
	var req dto.CreateTeamRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	team, users, err := a.service.CreateTeamWithMembers(ctx, req.TeamName, req.Members)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "create team")
	}

	teamMembers := dto.ToTeamMembers(users)

	return c.JSON(http.StatusCreated, map[string]any{
		"team": dto.TeamResponse{
			TeamName: team.Name,
			Members:  teamMembers,
		},
	})
}

func (a *API) getTeam(c echo.Context) error {
	teamName := c.QueryParam("team_name")
	if teamName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "team_name is required")
	}

	ctx := c.Request().Context()
	team, members, err := a.service.GetTeamByName(ctx, teamName)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "get team")
	}

	teamMembers := dto.ToTeamMembers(members)

	return c.JSON(http.StatusOK, dto.TeamResponse{
		TeamName: team.Name,
		Members:  teamMembers,
	})
}

func (a *API) deactivateTeamMembers(c echo.Context) error {
	var req dto.DeactivateTeamRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.TeamName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "team_name is required")
	}

	ctx := c.Request().Context()
	updated, err := a.service.DeactivateTeamAndReassign(ctx, req.TeamName)
	if err != nil {
		return handlers.ConvertDomainError(c, err, "deactivate team members")
	}

	return c.JSON(http.StatusOK, dto.DeactivateTeamResponse{
		TeamName:         req.TeamName,
		ReassignedPRsCnt: updated,
	})
}
