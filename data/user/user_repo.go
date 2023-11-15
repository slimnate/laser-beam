package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/slimnate/laser-beam/data"
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
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		admin_status INTEGER NOT NULL,
		organization_id INTEGER NOT NULL,
		FOREIGN KEY(organization_id) REFERENCES organizations(id)
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(user UserSecret) (*User, error) {
	query := "INSERT INTO users(username, password, first_name, last_name, admin_status, organization_id) values (?, ?, ?, ?, ?, ?)"
	res, err := r.db.Exec(query, user.Username, user.Password, user.FirstName, user.LastName, user.AdminStatus, user.OrganizationID)

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

	user.ID = id

	return &user.User, nil
}

func (r *SQLiteRepository) AllForOrganization(orgID int64) ([]User, error) {
	rows, err := r.db.Query("SELECT id, username, first_name, last_name, admin_status, organization_id FROM users WHERE organization_id = ?", orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.AdminStatus, &u.OrganizationID); err != nil {
			return nil, err
		}
		all = append(all, u)
	}
	return all, nil
}

func (r *SQLiteRepository) GetByID(id int64) (*User, error) {
	row := r.db.QueryRow("SELECT id, username, first_name, last_name, admin_status, organization_id FROM users WHERE id = ?", id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.AdminStatus, &u.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteRepository) GetByUsername(username string) (*UserSecret, error) {
	row := r.db.QueryRow("SELECT id, username, password, first_name, last_name, admin_status, organization_id FROM users WHERE username = ?", username)

	var u UserSecret
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.FirstName, &u.LastName, &u.AdminStatus, &u.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteRepository) UpdateUserInfo(id int64, new User) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE events SET first_name = ?, last_name = ? WHERE id = ?"
	res, err := r.db.Exec(query, new.FirstName, new.LastName, id)

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

func (r *SQLiteRepository) UpdateLoginInfo(id int64, new UserSecret) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE events SET username = ?, password = ? WHERE id = ?"
	res, err := r.db.Exec(query, new.Username, new.Password, id)

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
