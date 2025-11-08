package contactapp

import (
	"net/url"
	"testing"
)

func TestExtractPaginationData(t *testing.T) {
	t.Run("extract pagination data from URL", func(t *testing.T) {
		someQuery, _ := url.ParseQuery("page=4&limit=10")
		page, limit := extractPaginationData(someQuery)
		if page != 4 {
			t.Errorf("expected page be %d, got %d", 4, page)
		}
		if limit != 10 {
			t.Errorf("expected limit be %d, got %d", 10, limit)
		}
	})

	t.Run("data with default values from URL", func(t *testing.T) {
		someQuery, _ := url.ParseQuery("page=0&limit=ksk")
		page, limit := extractPaginationData(someQuery)
		if page != defaultPage {
			t.Errorf("expected page be %d, got %d", defaultPage, page)
		}
		if limit != defaultLimit {
			t.Errorf("expected limit be %d, got %d", defaultLimit, limit)
		}
	})

	t.Run("limit greater than 100 must be default to 10", func(t *testing.T) {
		someQuery, _ := url.ParseQuery("page=0&limit=1000")
		_, limit := extractPaginationData(someQuery)
		if limit != defaultLimit {
			t.Errorf("expected limit be %d, got %d", maxLimit, defaultLimit)
		}
	})
}
