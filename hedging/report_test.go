package hedging

import (
	"database/sql"
	"testing"
)

func TestNewReport(t *testing.T) {
	// Test creating a new report
	report, err := NewReport("test.db")
	if err != nil {
		t.Fatalf("Failed to create report: %v", err)
	}
	defer report.Close()

	// Test adding a report entry
	err = report.AddReport("AAPL", "S&P 500", 1.2, "2023-10-01")
	if err != nil {
		t.Fatalf("Failed to add report entry: %v", err)
	}

	// Test closing the report
	err = report.Close()
	if err != nil {
		t.Fatalf("Failed to close report: %v", err)
	}

	// Open database to verify the entry was added
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if the entry exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM report WHERE ticker = ? AND index_name = ? AND beta = ? AND date = ?", "AAPL", "S&P 500", 1.2, "2023-10-01").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 entry in the database, got %d", count)
	}
}
