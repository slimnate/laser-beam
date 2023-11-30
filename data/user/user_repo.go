package user

import (
	"database/sql"
	"errors"

	"github.com/slimnate/laser-beam/data"
)

// Repository
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(64) NOT NULL,
		first_name VARCHAR(64),
		last_name VARCHAR(64),
		email VARCHAR(128) NOT NULL,
		phone VARCHAR(20),
		admin_status INTEGER NOT NULL,
		organization_id INTEGER NOT NULL,
		FOREIGN KEY(organization_id) REFERENCES organizations(id)
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *UserRepository) Create(user UserSecret) (*User, error) {
	var lastInsertId int64
	query := "INSERT INTO users(username, password, first_name, last_name, email, phone, admin_status, organization_id) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	err := r.db.QueryRow(query, user.Username, user.Password, user.FirstName, user.LastName, user.Email, user.Phone, user.AdminStatus, user.OrganizationID).Scan(&lastInsertId)

	if err != nil {
		return nil, err
	}

	user.ID = lastInsertId

	return &user.User, nil
}

func (r *UserRepository) AllForOrganization(orgID int64) ([]User, error) {
	rows, err := r.db.Query("SELECT id, username, first_name, last_name, email, phone, admin_status, organization_id FROM users WHERE organization_id = $1", orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.AdminStatus, &u.OrganizationID); err != nil {
			return nil, err
		}
		all = append(all, u)
	}
	return all, nil
}

func (r *UserRepository) GetByID(id int64) (*User, error) {
	row := r.db.QueryRow("SELECT id, username, first_name, last_name, email, phone, admin_status, organization_id FROM users WHERE id = $1", id)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.AdminStatus, &u.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*UserSecret, error) {
	row := r.db.QueryRow("SELECT id, username, password, first_name, last_name, email, phone, admin_status, organization_id FROM users WHERE username = $1", username)

	var u UserSecret
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.AdminStatus, &u.OrganizationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateUserInfo(id int64, new User) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE users SET first_name = $1, last_name = $2, email = $3, phone = $4 WHERE id = $5"
	res, err := r.db.Exec(query, new.FirstName, new.LastName, new.Email, new.Phone, id)

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

func (r *UserRepository) UpdateLoginInfo(id int64, new UserSecret) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE users SET username = $1, password = $2 WHERE id = $3"
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

func (r *UserRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
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
