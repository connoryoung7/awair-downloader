package store

import (
	"database/sql"
	"fmt"

	"github.com/connoryoung/awair-downloader/internal/awair"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS readings (
		id          INTEGER PRIMARY KEY,
		timestamp   TEXT UNIQUE,
		score       REAL,
		temp        REAL,
		temp_index  REAL,
		humid       REAL,
		humid_index REAL,
		co2         REAL,
		co2_index   REAL,
		voc         REAL,
		voc_index   REAL,
		pm25        REAL,
		pm25_index  REAL
	)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func nullable(comp string, fn func(string) (float64, bool)) *float64 {
	if v, ok := fn(comp); ok {
		return &v
	}
	return nil
}

func (s *Store) Insert(r *awair.Reading) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO readings
			(timestamp, score, temp, temp_index, humid, humid_index, co2, co2_index, voc, voc_index, pm25, pm25_index)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		r.Score,
		nullable("temp", r.Sensor),
		nullable("temp", r.Index),
		nullable("humid", r.Sensor),
		nullable("humid", r.Index),
		nullable("co2", r.Sensor),
		nullable("co2", r.Index),
		nullable("voc", r.Sensor),
		nullable("voc", r.Index),
		nullable("pm25", r.Sensor),
		nullable("pm25", r.Index),
	)
	return err
}
