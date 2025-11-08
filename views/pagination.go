package views

import (
	"fmt"
	"net/url"
)

type Pagination struct {
	CurrentPage, TotalPage int
	URL                    *url.URL
}

func (p *Pagination) HasNext() bool {
	return p.CurrentPage < p.TotalPage && p.CurrentPage > 0
}

func (p *Pagination) HasPrev() bool {
	return p.CurrentPage > 1 && p.CurrentPage <= p.TotalPage
}

func (p *Pagination) Next() string {
	return p.Page(p.CurrentPage + 1)
}

func (p *Pagination) Prev() string {
	return p.Page(p.CurrentPage - 1)
}

func (p *Pagination) Page(page int) string {
	if page < 1 {
		page = 1
	}
	if page > p.TotalPage {
		page = p.TotalPage
	}
	query := p.URL.Query()
	query.Set("page", fmt.Sprintf("%d", page))
	p.URL.RawQuery = query.Encode()
	return p.URL.String()
}

// create a pagiantion struct for paginated views
func NewPagination(currentPage, totalPage int, u *url.URL) *Pagination {
	p := &Pagination{
		CurrentPage: currentPage,
		TotalPage:   totalPage,
		URL:         u,
	}
	if currentPage > totalPage {
		currentPage = totalPage
	}
	return p
}
