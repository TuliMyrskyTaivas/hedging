package hedging

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Cache struct {
	db *sql.DB
}

func NewCache() (*Cache, error) {
	const filename string = "cache.db"
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	const createCacheTable string = `
		CREATE TABLE IF NOT EXISTS assets (
			ticker STRING NOT NULL PRIMARY KEY,
			historyFrom DATETIME NOT NULL,
			historyTill DATETIME NOT NULL
		)`
	const createProfitsTable string = `
		CREATE TABLE IF NOT EXISTS profits (
			ticker STRING NOT NULL,
			date DATETIME NOT NULL,
			profit REAL,
			FOREIGN KEY (ticker) REFERENCES cache(ticker)
		)`

	if _, err = db.Exec(createCacheTable); err != nil {
		return nil, err
	}
	if _, err = db.Exec(createProfitsTable); err != nil {
		return nil, err
	}

	return &Cache{db: db}, nil
}

//func (cache *Cache) GetAvailableRange(asset string) (time.Time, time.Time, error) {
//	row := cache.db.QueryRow("SELECT historyFrom, historyFill FROM assets WHERE ticker=?", asset)
//
//	if err != nil {
//		return time.Time{}, time.Time{}, err
//	}
//
//}

func (cache *Cache) AddProfits(ticker string, profits []float64) {

}
