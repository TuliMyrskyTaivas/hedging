package hedging

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cache struct {
	db *sql.DB
}

func NewCache(filename string) (*Cache, error) {
	if filename == "" {
		filename = "cache.db"
	}
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	const createProfitsTable string = `
		CREATE TABLE IF NOT EXISTS profits (
			ticker STRING NOT NULL,
			date DATETIME NOT NULL,
			profit REAL,
			PRIMARY KEY (ticker, date)
		)`

	if _, err = db.Exec(createProfitsTable); err != nil {
		return nil, err
	}

	return &Cache{db: db}, nil
}

func (cache *Cache) GetAvailableRange(asset string) (time.Time, time.Time, error) {
	result := cache.db.QueryRow("SELECT min(date), max(date) FROM profits WHERE ticker=?", asset)

	var from, till string
	err := result.Scan(&from, &till)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	const TimeFormat = "2006-01-02"
	fromTime, err := time.Parse(TimeFormat, from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	tillTime, err := time.Parse(TimeFormat, till)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return fromTime, tillTime, nil
}

func (cache *Cache) PrintStats() error {
	result := cache.db.QueryRow("SELECT COUNT(date) FROM profits")

	var count int
	err := result.Scan(&count)
	if err == nil {
		slog.Debug(fmt.Sprintf("cache contains %d entries", count))
	}

	return err
}

func (cache *Cache) AddProfits(ticker string, dates []string, profits []float64) error {
	valueStrings := make([]string, 0, len(dates))
	valueArgs := make([]interface{}, 0, len(dates)*3)

	slog.Debug(fmt.Sprintf("insert %d profit records for %s", len(dates), ticker))
	if len(dates) != len(profits) {
		panic("the sizes of the sequences of dates and profits do not match")
	}

	for idx, date := range dates {
		valueStrings = append(valueStrings, "(?, ?, ?)")
		valueArgs = append(valueArgs, ticker)
		valueArgs = append(valueArgs, date)
		valueArgs = append(valueArgs, profits[idx])
	}

	statement := fmt.Sprintf("INSERT INTO profits (ticker, date, profit) VALUES %s", strings.Join(valueStrings, ","))
	result, err := cache.db.Exec(statement, valueArgs...)

	if err != nil {
		return err
	}

	inserted, err := result.RowsAffected()
	if err == nil {
		slog.Debug(fmt.Sprintf("%d profit records inserted for %s", inserted, ticker))
	}
	return err
}
