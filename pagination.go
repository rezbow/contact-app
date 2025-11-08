package contactapp

import (
	"net/url"
	"strconv"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

// extracts from query
func extractPaginationData(r url.Values) (int, int) {
	var (
		page  = defaultPage
		limit = defaultLimit
	)
	if p, err := strconv.Atoi(r.Get("page")); err == nil && p > 0 {
		page = p
	}

	if l, err := strconv.Atoi(r.Get("limit")); err == nil && l < maxLimit {
		limit = l
	}
	return page, limit
}
