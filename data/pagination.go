package data

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationRequestOptions struct {
	Offset  int64
	Limit   int64
	Filter  *FilterOptions
	OrderBy string
}

type FilterOptions struct {
	Key   string
	Value string
}

type PaginationResponseData[T any] struct {
	Data         T
	Request      *PaginationRequestOptions
	PreviousPage *PaginationRequestOptions
	NextPage     *PaginationRequestOptions
	Start        int64
	End          int64
	Total        int64
}

// Returns a string representation fof the object for logging
func (p *PaginationRequestOptions) ToString() string {
	return fmt.Sprintf("offset: %d Limit: %d Filter: %s OrderBy: %s", p.Offset, p.Limit, p.Filter, p.OrderBy)
}

// returns a set of query params that can be used to link to this PaginationRequestOptions object
func (p *PaginationRequestOptions) QueryParams() string {
	return fmt.Sprintf("?offset=%d&limit=%d&filter=%s&order_by=%s", p.Offset, p.Limit, p.Filter, p.OrderBy)
}

// returns the PaginationRequestOptions object that would correspond to the next page of items
func (p *PaginationRequestOptions) Next(total int64) *PaginationRequestOptions {
	res := p.Copy()
	newOffset := res.Offset + res.Limit

	// return nil if there are no pages beyond this one
	if newOffset >= total {
		return nil
	}
	res.Offset = newOffset

	return res
}

// returns the PaginationRequestOptions object that would correspond to the previous page of items
func (p *PaginationRequestOptions) Previous(total int64) *PaginationRequestOptions {
	res := p.Copy()
	newOffset := res.Offset - res.Limit

	if newOffset < 0 {
		return nil
	}

	res.Offset = newOffset

	return res
}

// returns a pointer to a copy of the current PaginationRequestOptions struct
func (p *PaginationRequestOptions) Copy() *PaginationRequestOptions {
	return &PaginationRequestOptions{
		Offset:  p.Offset,
		Limit:   p.Limit,
		Filter:  p.Filter,
		OrderBy: p.OrderBy,
	}
}

// Returns the system defaults for PaginationRequestOptions
func DefaultPaginationRequestOptions() *PaginationRequestOptions {
	return &PaginationRequestOptions{
		Offset:  0,
		Limit:   10,
		Filter:  nil,
		OrderBy: "id",
	}
}

// Parse pagination parameters from the query params of an incoming request, using the system defaults for any values not supplied
func ParsePaginationRequestOptions(ctx *gin.Context) (*PaginationRequestOptions, error) {
	return ParsePaginationRequestOptionsCustomDefault(ctx, DefaultPaginationRequestOptions())
}

// Parse pagination parameters from the query params of an incoming request, using custom defaults for any values not supplied
func ParsePaginationRequestOptionsCustomDefault(ctx *gin.Context, defaultOptions *PaginationRequestOptions) (*PaginationRequestOptions, error) {
	offset, offsetExists := ctx.GetQuery("offset")
	limit, limitExists := ctx.GetQuery("limit")
	_, filterExists := ctx.GetQuery("filter")
	order, orderExists := ctx.GetQuery("order_by")

	if offsetExists {
		offsetInt, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, err
		}
		defaultOptions.Offset = offsetInt
	}

	if limitExists {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, err
		}
		defaultOptions.Limit = limitInt
	}

	if filterExists {
		// TODO: parse filter values from string
	}

	// add custom or default order depending if it exists
	if orderExists {
		defaultOptions.OrderBy = order
	}

	log.Println(defaultOptions.ToString())

	return defaultOptions, nil
}
