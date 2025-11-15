package postgres

import (
	"github.com/Masterminds/squirrel"
)

type userSelectBuilder struct {
	b squirrel.SelectBuilder
}

func newUserSelectBuilder() *userSelectBuilder {
	b := newQueryBuilder().
		Select("id", "name", "team_id", "is_active").
		From("pr_review.users")

	return &userSelectBuilder{b: b}
}

func (u *userSelectBuilder) WhereIDs(ids []int64) *userSelectBuilder {
	if len(ids) > 0 {
		u.b = u.b.Where(squirrel.Eq{"id": ids})
	}
	return u
}

func (u *userSelectBuilder) WhereTeamID(teamID *int64) *userSelectBuilder {
	if teamID != nil && *teamID > 0 {
		u.b = u.b.Where(squirrel.Eq{"team_id": *teamID})
	}
	return u
}

func (u *userSelectBuilder) WhereIsActive(isActive *bool) *userSelectBuilder {
	if isActive != nil {
		u.b = u.b.Where(squirrel.Eq{"is_active": *isActive})
	}
	return u
}

func (u *userSelectBuilder) Limit(limit int) *userSelectBuilder {
	if limit > 0 {
		u.b = u.b.Limit(uint64(limit))
	}
	return u
}

func (u *userSelectBuilder) Offset(offset int) *userSelectBuilder {
	if offset > 0 {
		u.b = u.b.Offset(uint64(offset))
	}
	return u
}

func (u *userSelectBuilder) Build() (string, []interface{}, error) {
	return u.b.ToSql()
}
