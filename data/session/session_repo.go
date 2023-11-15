package session

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
		CREATE TABLE IF NOT EXISTS sessions(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL,
			login_time INTEGER NOT NULL,
			last_seen_time INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(key string, userID int64) (*Session, error) {
	t := time.Now()
	tUnix := t.Unix()

	query := "INSERT INTO sessions(key, login_time, last_seen_time, user_id) values(?, ?, ?, ?)"

	res, err := r.db.Exec(query, key, tUnix, tUnix, userID)
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

	session := Session{
		ID:        id,
		Key:       key,
		LoginTime: t,
		LastSeen:  t,
		UserID:    userID,
	}

	return &session, nil

}

func (r *SQLiteRepository) GetByKey(key string) (*Session, error) {
	query := "SELECT id, key, login_time, last_seen_time, user_id FROM sessions WHERE key = ?"

	row := r.db.QueryRow(query, key)

	var s Session
	var loginTimestamp, lastSeenTimestamp int64
	if err := row.Scan(&s.ID, &s.Key, &loginTimestamp, &lastSeenTimestamp, &s.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, data.ErrNotExists
		}
		return nil, err
	}

	s.LoginTime = time.Unix(loginTimestamp, 0)
	s.LastSeen = time.Unix(lastSeenTimestamp, 0)

	return &s, nil
}

func (r *SQLiteRepository) DeleteByKey(key string) error {
	query := "DELETE from sessions where key = ?"

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
