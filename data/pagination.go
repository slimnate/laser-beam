package data

import (
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaginationRequestOptions struct {
	Offset  int64
	Limit   int64
	Filter  *FilterOption
	OrderBy *OrderOption
}

type FilterOption struct {
	Key   string
	Value string
}

type OrderOption struct {
	Column    string
	Direction string
}

type PaginationResponseData[T any] struct {
	Data          T
	Request       *PaginationRequestOptions
	PreviousPage  *PaginationRequestOptions
	NextPage      *PaginationRequestOptions
	FilterOptions []FilterOptionsList
	Start         int64
	End           int64
	Total         int64
}

type FilterOptionsList struct {
	PropertyName  string
	Values        []string
	SelectedValue string
}

// Returns a string representation of the object for logging
func (p *PaginationRequestOptions) ToString() string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("offset: %d Limit: %d Filter: %s OrderBy: %s", p.Offset, p.Limit, FilterOptionsToString(p.Filter), OrderOptionsToString(p.OrderBy))
}

// returns a set of query params that can be used to link to this PaginationRequestOptions object
func (p *PaginationRequestOptions) QueryParams() template.URL {
	q := p.OffsetLimitQueryParams()
	q = ApplyFilterOptionsToQueryParams(q, p.Filter)
	q = ApplyOrderOptionsToQueryParams(q, p.OrderBy)
	return template.URL(q)
}

func (p *PaginationRequestOptions) OffsetLimitQueryParams() string {
	return fmt.Sprintf("?offset=%d&limit=%d&", p.Offset, p.Limit)
}

func ApplyFilterOptionsToQueryParams(q string, f *FilterOption) string {
	if f != nil {
		return fmt.Sprintf("%s&filter=%s", q, FilterOptionsToString(f))
	}
	return q
}

func ApplyOrderOptionsToQueryParams(q string, o *OrderOption) string {
	if o != nil {
		return fmt.Sprintf("%s&order_by=%s", q, OrderOptionsToString(o))
	}
	return q
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

// Returns a string representation of the object for logging
func FilterOptionsToString(f *FilterOption) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", f.Key, f.Value)
}

// Returns a string representation of the object for logging
func OrderOptionsToString(o *OrderOption) string {
	if o == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", o.Column, o.Direction)
}

// Generate a FilterOptions from a string in the format of "<key>:<value>". Returns an error if the string is empty, or the format incorrect
func FilterOptionsFromString(s string) (*FilterOption, error) {
	arr := strings.Split(s, ":")
	if len(arr) == 2 {
		return &FilterOption{
			Key:   arr[0],
			Value: arr[1],
		}, nil
	}
	return nil, errors.New("invalid format for 'filter' param")
}

// Generate an OrderOptions from a string in the format of "<column>" or "<column>:<direction>" where direction is either "asc" or "desc". If the string is empty, returns the default order options(id:asc). If only a column is provided, the direction defaults to "asc". Returns error if string matches none of these formats
func OrderOptionsFromString(s string) (*OrderOption, error) {
	arr := strings.Split(s, ":")
	if len(arr) == 0 {
		return &OrderOption{
			Column:    "id",
			Direction: "asc",
		}, nil
	} else if len(arr) == 1 {
		return &OrderOption{
			Column:    arr[0],
			Direction: "asc",
		}, nil
	} else if len(arr) == 2 {
		return &OrderOption{
			Column:    arr[0],
			Direction: arr[1],
		}, nil
	}
	return nil, errors.New("invalid format for 'order' param")
}

// Returns the system defaults for PaginationRequestOptions
func DefaultPaginationRequestOptions() *PaginationRequestOptions {
	return &PaginationRequestOptions{
		Offset: 0,
		Limit:  10,
		Filter: nil,
		OrderBy: &OrderOption{
			Column:    "id",
			Direction: "asc",
		},
	}
}

// Parse pagination parameters from the query params of an incoming request, using the system defaults for any values not supplied
func ParsePaginationRequestOptions(ctx *gin.Context) (*PaginationRequestOptions, error) {
	return ParsePaginationRequestOptionsCustomDefault(ctx, DefaultPaginationRequestOptions())
}

// Parse pagination parameters from the query params of an incoming request, using custom defaults for any values not supplied
func ParsePaginationRequestOptionsCustomDefault(ctx *gin.Context, defaultOptions *PaginationRequestOptions) (*PaginationRequestOptions, error) {
	sysDefaults := DefaultPaginationRequestOptions()
	offset, offsetExists := ctx.GetQuery("offset")
	limit, limitExists := ctx.GetQuery("limit")
	filter, filterExists := ctx.GetQuery("filter")
	order, orderExists := ctx.GetQuery("order_by")

	if offsetExists {
		offsetInt, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			return nil, err
		}
		defaultOptions.Offset = offsetInt
	}
	if defaultOptions.Offset == 0 {
		defaultOptions.Offset = sysDefaults.Offset
	}

	if limitExists {
		limitInt, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			return nil, err
		}
		defaultOptions.Limit = limitInt
	}
	if defaultOptions.Limit == 0 { // no default provided, use system defaults
		defaultOptions.Limit = sysDefaults.Limit
	}

	if filterExists {
		filterOptions, err := FilterOptionsFromString(filter)
		if err != nil {
			return nil, err
		}
		defaultOptions.Filter = filterOptions
	}
	if defaultOptions.Filter == nil {
		defaultOptions.Filter = sysDefaults.Filter
	}

	if orderExists {
		orderOptions, err := OrderOptionsFromString(order)
		if err != nil {
			return nil, err
		}
		defaultOptions.OrderBy = orderOptions
	}
	if defaultOptions.OrderBy == nil {
		defaultOptions.OrderBy = sysDefaults.OrderBy
	}

	return defaultOptions, nil
}

func (p *PaginationRequestOptions) SortLinkForColumn(col string) template.URL {
	q := p.OffsetLimitQueryParams()
	q = ApplyFilterOptionsToQueryParams(q, p.Filter)
	newOrder := OrderOption{}
	if p.OrderBy == nil {
		newOrder.Column = "id"
		newOrder.Direction = "asc"
	} else {
		newOrder.Column = col
		if col == p.OrderBy.Column {
			// same column, so reverse directions
			if strings.ToLower(p.OrderBy.Direction) == "asc" {
				newOrder.Direction = "desc"
			} else {
				newOrder.Direction = "asc"
			}
		} else {
			// new column, so default to asc
			newOrder.Direction = "asc"
		}
	}
	q = ApplyOrderOptionsToQueryParams(q, &newOrder)

	return template.URL(q)
}

func (p *PaginationRequestOptions) FilterLink(col string, value string) template.URL {
	q := p.OffsetLimitQueryParams()
	q = ApplyOrderOptionsToQueryParams(q, p.OrderBy)

	// if col and value both provided, init newfilter. Otherwise it wil be nil
	// this is how we get a link to "unselect" a filter
	var newFilter *FilterOption
	if col != "" && value != "" {
		newFilter = &FilterOption{
			Key:   col,
			Value: value,
		}
	}

	q = ApplyFilterOptionsToQueryParams(q, newFilter)
	return template.URL(q)
}
