package organization

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// Errors
var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row does not exist")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

// Repository
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
	CREATE TABLE IF NOT EXISTS organizations(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		key TEXT NOT NULL
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(org Organization) (*Organization, error) {
	query := "INSERT INTO organizations(name, key) values(?, ?)"
	res, err := r.db.Exec(query, org.Name, org.Key)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		fmt.Printf("err: %v\n", err)
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	org.ID = id

	return &org, nil
}

func (r *SQLiteRepository) All() ([]Organization, error) {
	rows, err := r.db.Query("SELECT * from organizations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Key); err != nil {
			return nil, err
		}
		all = append(all, org)
	}
	return all, nil
}

func (r *SQLiteRepository) GetByID(id int64) (*Organization, error) {
	row := r.db.QueryRow("SELECT id, key, name FROM organizations WHERE id = ?", id)

	var org Organization
	if err := row.Scan(&org.ID, &org.Key, &org.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &org, nil
}

func (r *SQLiteRepository) GetByKey(key string) (*Organization, error) {
	row := r.db.QueryRow("SELECT id, key, name FROM organizations WHERE key = ?", key)

	var org Organization
	if err := row.Scan(&org.ID, &org.Key, &org.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &org, nil
}

func (r *SQLiteRepository) Update(id int64, updated Organization) (*Organization, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE organizations SET name = ?, key = ? WHERE id = ?"
	res, err := r.db.Exec(query, updated.Name, updated.Key, updated.ID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM organizations WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
