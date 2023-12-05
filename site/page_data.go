package site

import (
	"github.com/slimnate/laser-beam/data"
	"github.com/slimnate/laser-beam/data/event"
	"github.com/slimnate/laser-beam/data/organization"
	"github.com/slimnate/laser-beam/data/user"
)

type PageData struct {
	User         *user.User
	Organization *organization.Organization
	Events       *data.PaginationResponseData[[]event.Event]
	Route        string
	Errors       map[string]string
	Toasts       []string
}

func (d PageData) HasError(name string) bool {
	if d.Errors != nil {
		_, ok := d.Errors[name]
		return ok
	}
	return false
}

func (d *PageData) AddToast(s string) {
	d.Toasts = append(d.Toasts, s)
}
