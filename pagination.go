// Pagination
// We're using two popular kinds of pagination methods:
// 1. Offset/Limit pagination --> OffsetPagination (i. Paginator)
// 2. Seek/Keyset/Cursor pagination --> SeekPagination (i. SeekPaginator)
//
// Reference:
// - https://blog.jooq.org/2013/10/26/faster-sql-paging-with-jooq-using-the-seek-method/
package pg

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// OffsetPagination holds paging info in offset pagination method.
type OffsetPagination struct {
	Page         int64 `json:"page" in:"query=page" `
	PerPage      int64 `json:"per_page" in:"query=per_page"`
	CountPages   int64 `json:"count_pages"`
	CountRecords int64 `json:"count_records"`

	defaultPerPage int64
}

// NewOffsetPagination creates a new `Pagination` with a default page size.
func NewOffsetPagination(defaultPerPage int64) *OffsetPagination {
	p := &OffsetPagination{
		defaultPerPage: defaultPerPage,
	}
	p.normalize()
	return p
}

// SetDefaultPerPage sets the default page size.
func (p *OffsetPagination) SetDefaultPerPage(defaultPerPage int64) int64 {
	p.defaultPerPage = defaultPerPage
	p.normalize()
	return p.defaultPerPage
}

// Limit returns the page size.
func (p *OffsetPagination) Limit() int64 {
	p.normalize()
	return p.PerPage
}

// Offset returns the size of skipped items of current page.
func (p *OffsetPagination) Offset() int64 {
	p.normalize()
	return (p.Page - 1) * p.PerPage
}

// CurrentPage returns the current page index.
func (p *OffsetPagination) CurrentPage() int64 {
	p.normalize()
	return p.Page
}

// PageSize returns the page size.
func (p *OffsetPagination) PageSize() int64 {
	p.normalize()
	return p.PerPage
}

// SetCountRecords update the `Records` field.
func (p *OffsetPagination) SetCountRecords(total int64) {
	p.CountRecords = total
	p.normalize()
}

func (p *OffsetPagination) normalize() {
	if p.defaultPerPage <= 0 {
		p.defaultPerPage = 20
	}

	if p.Page <= 0 {
		p.Page = 1
	}

	if p.PerPage <= 0 {
		p.PerPage = p.defaultPerPage
	}

	if p.CountRecords <= 0 {
		p.CountRecords = 0
	}

	p.CountPages = int64(math.Ceil(float64(p.CountRecords) / float64(p.PerPage)))
}

// LinkHeader compose a Link Header for the HTTP response.
// See: https://www.w3.org/wiki/LinkHeader
// e.g. Link: <https://api.example.com/users?page=1>; rel="first", <https://api.example.com/users?page=2>; rel="next"
func (p *OffsetPagination) LinkHeader(theURL *url.URL) string {
	var linkHeaders []string
	firstLink := theURL.Query()
	firstLink.Set("page", "1")
	linkHeaders = append(linkHeaders, fmt.Sprintf(`<%s?%s>; rel="first"`, theURL.Path, firstLink.Encode()))

	if p.Page > 1 {
		prevLink := theURL.Query()
		prevLink.Set("page", strconv.FormatInt(p.Page-1, 10))
		linkHeaders = append(linkHeaders, fmt.Sprintf(`<%s?%s>; rel="prev"`, theURL.Path, prevLink.Encode()))
	}

	if p.Page+1 < p.CountPages {
		nextLink := theURL.Query()
		nextLink.Set("page", strconv.FormatInt(p.Page+1, 10))
		linkHeaders = append(linkHeaders, fmt.Sprintf(`<%s?%s>; rel="next"`, theURL.Path, nextLink.Encode()))
	}

	lastLink := theURL.Query()
	lastLink.Set("page", strconv.FormatInt(p.CountPages, 10))
	linkHeaders = append(linkHeaders, fmt.Sprintf(`<%s?%s>; rel="last"`, theURL.Path, lastLink.Encode()))

	return strings.Join(linkHeaders, ", ")
}

// XPaginationHeader compose a text in format: {Page},{Size},{CountPages},{CountRecords}
// providing the information of this pagination.
// e.g. X-Pagination: 1,20,10,200
func (p *OffsetPagination) XPaginationHeader() string {
	return strings.Join([]string{
		strconv.FormatInt(p.Page, 10),
		strconv.FormatInt(p.PerPage, 10),
		strconv.FormatInt(p.CountPages, 10),
		strconv.FormatInt(p.CountRecords, 10),
	}, ",")
}

func (p *OffsetPagination) String() string {
	return "OffsetPagination#" + p.XPaginationHeader()
}

// SetResponseHeaders write paging info headers to the HTTP response.
func (p *OffsetPagination) SetResponseHeaders(rw http.ResponseWriter, r *http.Request) {
	// Add Link header for pagination info.
	rw.Header().Set("Link", p.LinkHeader(r.URL))
	rw.Header().Set("X-Pagination", p.XPaginationHeader())
}

// SeekPagination holds paging info in seek pagination method.
type SeekPagination struct {
	limit  int64
	cursor string

	defaultLimit int64
}

// NewSeekPagination creates a new SeekPagination with default limit value.
func NewSeekPagination(defaultLimit int64) *SeekPagination {
	if defaultLimit <= 0 {
		defaultLimit = 10
	}
	return &SeekPagination{
		defaultLimit: defaultLimit,
	}
}

// SetLimit updates limit to a new value and returns the new value.
func (p *SeekPagination) SetLimit(newLimit int64) int64 {
	p.limit = newLimit
	p.normalize()
	return p.limit
}

// Limit returns a valid limit (>0) number.
func (p *SeekPagination) Limit() int64 {
	return p.limit
}

// SetCursor updates cursor to a new value and returns the new value.
func (p *SeekPagination) SetCursor(newCursor string) string {
	p.cursor = newCursor
	return p.cursor
}

// Cursor returns the cursor string.
func (p *SeekPagination) Cursor() string {
	return p.cursor
}

func (p *SeekPagination) normalize() {
	if p.limit <= 0 {
		p.limit = p.defaultLimit
	}
}

// LinkHeader compose a Link Header for the HTTP response.
// See: https://www.w3.org/wiki/LinkHeader
func (p *SeekPagination) LinkHeader(theURL *url.URL) string {
	var linkHeaders []string

	nextLink := theURL.Query()
	nextLink.Set("limit", strconv.FormatInt(p.Limit(), 10))
	nextLink.Set("cursor", p.Cursor())
	linkHeaders = append(linkHeaders, fmt.Sprintf(`<%s?%s>; rel="next"`, theURL.Path, nextLink.Encode()))

	return strings.Join(linkHeaders, ", ")
}

// XPaginationHeader compose a text in format: {Cursor},{Limit}
// providing the information of this pagination.
// e.g. X-Pagination: dXNlcjoxMCwz,20
func (p *SeekPagination) XPaginationHeader() string {
	return strings.Join([]string{
		p.Cursor(),
		strconv.FormatInt(p.Limit(), 10),
	}, ",")
}

// SetResponseHeaders write paging info headers to the HTTP response.
func (p *SeekPagination) SetResponseHeaders(rw http.ResponseWriter, r *http.Request) {
	// Add Link header for pagination info.
	rw.Header().Set("Link", p.LinkHeader(r.URL))
	rw.Header().Set("X-Pagination", p.XPaginationHeader())
}
