package session

import (
	"database/sql"
	"errors"
	"time"

	"github.com/slimnate/laser-beam/data"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS sessions(
			id SERIAL PRIMARY KEY,
			key CHAR(64) NOT NULL,
			login_time TIMESTAMP NOT NULL,
			last_seen_time TIMESTAMP NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SessionRepository) Create(key string, userID int64) (*Session, error) {
	var lastInsertId int64
	t := time.Now()

	query := "INSERT INTO sessions(key, login_time, last_seen_time, user_id) values($1, $2, $3, $4) RETURNING id"

	err := r.db.QueryRow(query, key, t, t, userID).Scan(&lastInsertId)

	if err != nil {
		return nil, err
	}

	session := Session{
		ID:        lastInsertId,
		Key:       key,
		LoginTime: t,
		LastSeen:  t,
		UserID:    userID,
	}

	return &session, nil
}

func (r *SessionRepository) GetByKey(key string) (*Session, error) {
	query := "SELECT id, key, login_time, last_seen_time, user_id FROM sessions WHERE key = $1"

	row := r.db.QueryRow(query, key)

	var s Session
	if err := row.Scan(&s.ID, &s.Key, &s.LoginTime, &s.LastSeen, &s.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}

	return &s, nil
}

func (r *SessionRepository) DeleteByKey(key string) error {
	query := "DELETE from sessions where key = $1"

	res, err := r.db.Exec(query, key)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected < 1 {
		return errors.New("not enough rows affected")
	} else if affected > 1 {
		return errors.New("too many rows affected")
	}
	return nil
}
