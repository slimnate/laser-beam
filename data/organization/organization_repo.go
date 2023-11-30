package organization

import (
	"database/sql"
	"errors"

	"github.com/slimnate/laser-beam/data"
)

// Repository
type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

func (r *OrganizationRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS organizations(
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		key VARCHAR(100) NOT NULL
	)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *OrganizationRepository) Create(org Organization, key string) (*Organization, error) {
	var lastInsertId int64
	query := "INSERT INTO organizations (name, key) values ($1, $2) RETURNING id"
	err := r.db.QueryRow(query, org.Name, key).Scan(&lastInsertId)

	if err != nil {
		return nil, err
	}

	org.ID = lastInsertId

	return &org, nil
}

func (r *OrganizationRepository) All() ([]Organization, error) {
	rows, err := r.db.Query("SELECT id, name from organizations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name); err != nil {
			return nil, err
		}
		all = append(all, org)
	}
	return all, nil
}

func (r *OrganizationRepository) GetByID(id int64) (*Organization, error) {
	row := r.db.QueryRow("SELECT id, name FROM organizations WHERE id = $1", id)

	var org Organization
	if err := row.Scan(&org.ID, &org.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) GetByKey(key string) (*Organization, error) {
	row := r.db.QueryRow("SELECT id, name FROM organizations WHERE key = $1", key)

	var org Organization
	if err := row.Scan(&org.ID, &org.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) Update(id int64, updated Organization) (*Organization, error) {
	if id == 0 {
		return nil, errors.New("invalid ID to update")
	}
	query := "UPDATE organizations SET name = $1 WHERE id = $2"
	res, err := r.db.Exec(query, updated.Name, updated.ID)

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

	return &updated, nil
}

func (r *OrganizationRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM organizations WHERE id = $1", id)
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
