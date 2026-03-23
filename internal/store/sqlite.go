package store

import (
	"database/sql"
	"fmt"

	"github.com/connoryoung/awair-downloader/internal/domain"
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
		temp_index  INTEGER,
		humid       REAL,
		humid_index INTEGER,
		co2         REAL,
		co2_index   INTEGER,
		voc         REAL,
		voc_index   INTEGER,
		pm25        REAL,
		pm25_index  INTEGER
	)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) InsertReading(r domain.Reading) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO readings
			(timestamp, score, temp, temp_index, humid, humid_index, co2, co2_index, voc, voc_index, pm25, pm25_index)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		r.Score,
		r.Temp.Value, r.Temp.Index,
		r.Humidity.Value, r.Humidity.Index,
		r.CO2.Value, r.CO2.Index,
		r.VOC.Value, r.VOC.Index,
		r.PM25.Value, r.PM25.Index,
	)
	return err
}
