package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams holds parsed, bounds-checked pagination query params.
type PaginationParams struct {
	Page    int
	PerPage int
	Offset  int
}

// ParsePagination reads page/per_page from the query string, applying the
// defaults and caps required by the spec (default 10, max 50).
func ParsePagination(c *fiber.Ctx) PaginationParams {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil || perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
	}
}

// BuildMeta turns a total row count into the Meta object for a paginated response.
func BuildMeta(p PaginationParams, total int) Meta {
	lastPage := total / p.PerPage
	if total%p.PerPage != 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}

	return Meta{
		CurrentPage: p.Page,
		PerPage:     p.PerPage,
		Total:       total,
		LastPage:    lastPage,
	}
}
