package event

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/slimnate/laser-beam/data"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS events(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		application TEXT,
		message TEXT,
		time INTEGER NOT NULL,
		organization_id INTEGER NOT NULL,
		FOREIGN KEY(organization_id) REFERENCES organizations(id)
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(event Event, orgID int64) (*Event, error) {
	t := time.Now()
	tUnix := t.Unix()
	query := "INSERT INTO events(type, name, application, message, time, organization_id) values(?, ?, ?, ?, ?, ?)"
	res, err := r.db.Exec(query, event.Type, event.Name, event.Application, event.Message, tUnix, orgID)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, data.ErrDuplicate
			}
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintForeignKey) {
				return nil, data.ErrForeignKey
			}
		}

		fmt.Printf("err: %v\n", err)
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	event.ID = id
	event.Time = t
	event.OrganizationID = orgID

	return &event, nil
}

func (r *SQLiteRepository) All() ([]Event, error) {
	rows, err := r.db.Query("SELECT id, type, name, application, message, time, organization_id from events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Event
	for rows.Next() {
		var e Event
		var timestamp int64
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &timestamp, &e.OrganizationID); err != nil {
			return nil, err
		}
		e.Time = time.Unix(timestamp, 0)
		all = append(all, e)
	}
	return all, nil
}

func (r *SQLiteRepository) AllForOrganization(orgID int64) ([]Event, error) {
	rows, err := r.db.Query("SELECT id, type, name, application, message, time, organization_id from events WHERE organization_id = ?", orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Event
	for rows.Next() {
		var e Event
		var timestamp int64
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &timestamp, &e.OrganizationID); err != nil {
			return nil, err
		}
		e.Time = time.Unix(timestamp, 0)
		all = append(all, e)
	}
	return all, nil
}

func (r *SQLiteRepository) GetByID(id int64) (*Event, error) {
	row := r.db.QueryRow("SELECT id, type, name, application, message, time, organization_id FROM events WHERE id = ?", id)

	var e Event
	var timestamp int64
	if err := row.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &timestamp, &e.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	e.Time = time.Unix(timestamp, 0)
	return &e, nil
}

func (r *SQLiteRepository) GetByIDAndOrg(id int64, orgID int64) (*Event, error) {
	row := r.db.QueryRow("SELECT id, type, name, application, message, time, organization_id FROM events WHERE id = ? AND organization_id = ?", id, orgID)

	var e Event
	var timestamp int64
	if err := row.Scan(&e.ID, &e.Type, &e.Name, &e.Application, &e.Message, &timestamp, &e.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	e.Time = time.Unix(timestamp, 0)
	return &e, nil
}

func (r *SQLiteRepository) Update(id int64, newEvent Event) (*Event, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE events SET name = ?, type = ?, message = ?, application = ? WHERE id = ?"
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

func (r *SQLiteRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM events WHERE id = ?", id)
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
