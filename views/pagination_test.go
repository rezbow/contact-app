package views

import (
	"net/url"
	"testing"
)

func TestPagination(t *testing.T) {
	t.Run("has next and prev page", func(t *testing.T) {
		someUrl, _ := url.Parse("https://localhost.com/contacts/?page=1&limit=5")
		pagination := NewPagination(5, 10, someUrl)
		if !pagination.HasNext() {
			t.Errorf("expected next page but didn't have any")
		}
		if !pagination.HasPrev() {
			t.Errorf("expected prev page but didn't have any")
		}
	})

	t.Run("no next page and has prev page", func(t *testing.T) {
		someUrl, _ := url.Parse("https://localhost.com/contacts/?page=1&limit=5")
		pagination := NewPagination(10, 10, someUrl)
		if pagination.HasNext() {
			t.Errorf("expected no next page but there is")
		}
		if !pagination.HasPrev() {
			t.Errorf("expected prev page but didn't have any")
		}
	})

	t.Run("has next page and no prev page", func(t *testing.T) {
		someUrl, _ := url.Parse("https://localhost.com/contacts/?page=1&limit=5")
		pagination := NewPagination(1, 10, someUrl)
		if !pagination.HasNext() {
			t.Errorf("expected next page but there is not")
		}
		if pagination.HasPrev() {
			t.Errorf("expected no prev page but there is")
		}
	})

	t.Run("no next or prev page (outside boundary)", func(t *testing.T) {
		someUrl, _ := url.Parse("https://localhost.com/contacts/?page=1&limit=5")
		pagination := NewPagination(100, 10, someUrl)
		if pagination.HasNext() {
			t.Errorf("expected no next page but there is")
		}
		if pagination.HasPrev() {
			t.Errorf("expected no prev page but there is")
		}
		pagination = NewPagination(-1, 10, someUrl)
		if pagination.HasNext() {
			t.Errorf("expected no next page but there is")
		}
		if pagination.HasPrev() {
			t.Errorf("expected no prev page but there is")
		}
	})

}
