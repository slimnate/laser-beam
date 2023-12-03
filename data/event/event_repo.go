package event

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/slimnate/laser-beam/data"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

func (r *EventRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS events(
		id SERIAL UNIQUE PRIMARY KEY,
		type VARCHAR(50) NOT NULL,
		name VARCHAR(250) NOT NULL,
		application VARCHAR(50),
		message VARCHAR(1000),
		time TIMESTAMP NOT NULL,
		organization_id INTEGER NOT NULL,
		FOREIGN KEY(organization_id) REFERENCES organizations(id)
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *EventRepository) Create(event Event, orgID int64) (*Event, error) {
	var lastInsertId int64
	query := "INSERT INTO events(type, name, application, message, time, organization_id) values($1, $2, $3, $4, $5, $6) RETURNING id"
	err := r.db.QueryRow(query, event.Type, event.Name, event.Application, event.Message, event.Time, orgID).Scan(&lastInsertId)

	if err != nil {
		return nil, err
	}

	event.ID = lastInsertId
	event.OrganizationID = orgID

	return &event, nil
}

func (r *EventRepository) All(pag *data.PaginationRequestOptions) ([]Event, error) {
	rows, err := r.db.Query("SELECT id, type, name, application, message, time, organization_id from events WHERE $1 = $2 ORDER BY $3 LIMIT $4 OFFSET $5", pag.Filter.Key, pag.Filter.Value, pag.OrderBy, pag.Limit, pag.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &e.Time, &e.OrganizationID); err != nil {
			return nil, err
		}
		all = append(all, e)
	}
	return all, nil
}

func (r *EventRepository) AllForOrganization(orgID int64, pag *data.PaginationRequestOptions) (*data.PaginationResponseData[[]Event], error) {
	if pag == nil {
		log.Printf("EventRepository.AllForOrganization: No pagination options supplied, using defaults")
		pag = data.DefaultPaginationRequestOptions()
	}

	var rows *sql.Rows
	var err error
	log.Printf("%+v\n", pag)
	if pag.Filter != nil {
		query := fmt.Sprintf("SELECT id, type, name, application, message, time, organization_id from events WHERE organization_id = $1 AND %s = $2 ORDER BY %s %s LIMIT $3 OFFSET $4", pag.Filter.Key, pag.OrderBy.Column, pag.OrderBy.Direction)
		// Query with filter
		rows, err = r.db.Query(query, orgID, pag.Filter.Value, pag.Limit, pag.Offset)
	} else {
		// Query without filter
		query := fmt.Sprintf("SELECT id, type, name, application, message, time, organization_id from events WHERE organization_id = $1 ORDER BY %s %s LIMIT $2 OFFSET $3", pag.OrderBy.Column, pag.OrderBy.Direction)
		rows, err = r.db.Query(query, orgID, pag.Limit, pag.Offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &e.Time, &e.OrganizationID); err != nil {
			return nil, err
		}
		results = append(results, e)
	}

	total, err := r.Count(&data.FilterOption{
		Key:   "organization_id",
		Value: fmt.Sprint(orgID),
	}, pag.Filter)

	if err != nil {
		return nil, err
	}

	typeValues, err := r.GetDistinctValues("type")
	if err != nil {
		return nil, err
	}
	typeOptionsList := data.FilterOptionsList{
		PropertyName: "type",
		Values:       typeValues,
	}
	if pag.Filter != nil && pag.Filter.Key == "type" {
		typeOptionsList.SelectedValue = pag.Filter.Value
	}

	appValues, err := r.GetDistinctValues("application")
	if err != nil {
		return nil, err
	}
	appOptionsList := data.FilterOptionsList{
		PropertyName: "application",
		Values:       appValues,
	}
	if pag.Filter != nil && pag.Filter.Key == "application" {
		appOptionsList.SelectedValue = pag.Filter.Value
	}

	pagRes := &data.PaginationResponseData[[]Event]{
		Data:          results,
		Request:       pag,
		PreviousPage:  pag.Previous(total),
		NextPage:      pag.Next(total),
		FilterOptions: []data.FilterOptionsList{typeOptionsList, appOptionsList},
		Total:         total,
		Start:         pag.Offset + 1,
		End:           min(pag.Offset+pag.Limit, total),
	}

	return pagRes, nil
}

func (r *EventRepository) GetByID(id int64) (*Event, error) {
	row := r.db.QueryRow("SELECT id, type, name, application, message, time, organization_id FROM events WHERE id = $1", id)

	var e Event
	if err := row.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &e.Time, &e.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &e, nil
}

func (r *EventRepository) GetByIDAndOrg(id int64, orgID int64) (*Event, error) {
	row := r.db.QueryRow("SELECT id, type, name, application, message, time, organization_id FROM events WHERE id = $1 AND organization_id = $2", id, orgID)

	var e Event
	if err := row.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &e.Time, &e.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &e, nil
}

func (r *EventRepository) Update(id int64, newEvent Event) (*Event, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE events SET name = $1, type = $2, message = $3, application = $4 WHERE id = $5"
	res, err := r.db.Exec(query, newEvent.Name, newEvent.Type, newEvent.Message, newEvent.Application, id)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, data.ErrUpdateFailed
	}

	updated, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *EventRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return data.ErrDeleteFailed
	}

	return err
}

func (r *EventRepository) Count(filters ...*data.FilterOption) (int64, error) {
	var output string
	var filterStrs []string
	var args []any

	for i, f := range filters {
		if f != nil {
			s := fmt.Sprintf("%s = $%d", f.Key, i+1)
			filterStrs = append(filterStrs, s)
			args = append(args, f.Value)
		}
	}

	q := fmt.Sprintf("SELECT COUNT(id) FROM events WHERE %s", strings.Join(filterStrs, " AND "))

	query, err := r.db.Prepare(q)
	if err != nil {
		return -1, err
	}

	err = query.QueryRow(args...).Scan(&output)
	if err != nil {
		return -1, err
	}

	count, err := strconv.ParseInt(output, 10, 64)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (r *EventRepository) GetDistinctValues(columnName string) ([]string, error) {
	var values []string

	q := fmt.Sprintf("SELECT DISTINCT %s FROM events", columnName)

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var v string
		err := rows.Scan(&v)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}
