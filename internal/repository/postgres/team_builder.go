package postgres

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

type teamSelectBuilder struct {
	b squirrel.SelectBuilder
}

func newTeamSelectBuilder() *teamSelectBuilder {
	b := newQueryBuilder().
		Select("id", "name").
		From("pr_review.team")

	return &teamSelectBuilder{b: b}
}

func (t *teamSelectBuilder) WhereIDs(ids []int64) *teamSelectBuilder {
	if len(ids) > 0 {
		t.b = t.b.Where(squirrel.Eq{"id": ids})
	}
	return t
}

func (t *teamSelectBuilder) WhereName(name string) *teamSelectBuilder {
	if name != "" {
		t.b = t.b.Where(squirrel.Like{"name": "%" + name + "%"})
	}
	return t
}

func (t *teamSelectBuilder) Limit(limit int) *teamSelectBuilder {
	if limit > 0 {
		t.b = t.b.Limit(uint64(limit))
	}
	return t
}

func (t *teamSelectBuilder) Offset(offset int) *teamSelectBuilder {
	if offset > 0 {
		t.b = t.b.Offset(uint64(offset))
	}
	return t
}

func (t *teamSelectBuilder) OrderBy(field, direction string) *teamSelectBuilder {
	if field == "" {
		return t
	}
	if direction == "" {
		direction = "ASC"
	}
	t.b = t.b.OrderBy(fmt.Sprintf("%s %s", field, direction))
	return t
}

func (t *teamSelectBuilder) Build() (string, []interface{}, error) {
	return t.b.ToSql()
}
