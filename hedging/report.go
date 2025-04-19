package hedging

import "database/sql"

type Report struct {
	db *sql.DB
}

func NewReport(filename string) (*Report, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	const createReportTable string = `
		CREATE TABLE IF NOT EXISTS report (
			ticker STRING NOT NULL,
			index_name STRING NOT NULL,
			beta REAL NOT NULL,
			date DATETIME NOT NULL,
			primary key (ticker, index_name)
		)`

	if _, err = db.Exec(createReportTable); err != nil {
		return nil, err
	}

	return &Report{db: db}, nil
}

func (report *Report) AddReport(ticker string, index string, beta float64, date string) error {
	_, err := report.db.Exec("INSERT OR REPLACE INTO report (ticker, index_name, beta, date) VALUES (?, ?, ?, ?)", ticker, index, beta, date)
	return err
}

func (report *Report) Close() error {
	return report.db.Close()
}
