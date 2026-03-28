package pagination

import (
	"net/http"
	"strconv"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// Params holds parsed pagination parameters.
type Params struct {
	Page    int
	PerPage int
}

// Meta holds pagination metadata for API responses.
type Meta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

// Parse extracts page and per_page from query parameters with defaults and bounds.
func Parse(r *http.Request) Params {
	page := parseIntParam(r, "page", defaultPage)
	perPage := parseIntParam(r, "per_page", defaultPerPage)

	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return Params{
		Page:    page,
		PerPage: perPage,
	}
}

// Offset returns the SQL OFFSET value for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL LIMIT value (same as PerPage).
func (p Params) Limit() int {
	return p.PerPage
}

// NewMeta creates a Meta struct from the pagination params and a total count.
func (p Params) NewMeta(total int64) Meta {
	return Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	}
}

func parseIntParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
