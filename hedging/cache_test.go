package hedging

import (
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// TestDBFile is the name of the test database file
	TestDBFile = "test_cache.db"
)

func setupTestDB(t *testing.T) *Cache {
	t.Helper()

	// Remove any existing test database
	_ = os.Remove(TestDBFile)

	// Create a new cache instance
	cache, err := NewCache(TestDBFile)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	return cache
}

func teardownTestDB(cache *Cache) {
	cache.db.Close()
	_ = os.Remove(TestDBFile)
}

func TestNewCache(t *testing.T) {
	cache := setupTestDB(t)
	defer teardownTestDB(cache)

	// Check if the database connection is valid
	if cache.db == nil {
		t.Fatal("expected a valid database connection, got nil")
	}

	// Check if the profits table exists
	_, err := cache.db.Exec("SELECT 1 FROM profits LIMIT 1")
	if err != nil {
		t.Fatalf("profits table does not exist: %v", err)
	}
}

func TestAddProfits(t *testing.T) {
	cache := setupTestDB(t)
	defer teardownTestDB(cache)

	ticker := "AAPL"
	dates := []string{"2023-01-01", "2023-01-02"}
	profits := []float64{100.5, 200.75}

	err := cache.AddProfits(ticker, dates, profits)
	if err != nil {
		t.Fatalf("failed to add profits: %v", err)
	}

	// Verify the inserted data
	rows, err := cache.db.Query("SELECT ticker, date, profit FROM profits WHERE ticker=?", ticker)
	if err != nil {
		t.Fatalf("failed to query profits: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	if count != len(dates) {
		t.Fatalf("expected %d rows, got %d", len(dates), count)
	}
}

func TestGetAvailableRange(t *testing.T) {
	cache := setupTestDB(t)
	defer teardownTestDB(cache)

	ticker := "AAPL"
	dates := []string{"2023-01-01", "2023-01-02", "2023-01-03"}
	profits := []float64{100.5, 200.75, 300.25}

	err := cache.AddProfits(ticker, dates, profits)
	if err != nil {
		t.Fatalf("failed to add profits: %v", err)
	}

	from, till, err := cache.GetAvailableRange(ticker)
	if err != nil {
		t.Fatalf("failed to get available range: %v", err)
	}

	expectedFrom, _ := time.Parse("2006-01-02", dates[0])
	expectedTill, _ := time.Parse("2006-01-02", dates[len(dates)-1])

	if !from.Equal(expectedFrom) || !till.Equal(expectedTill) {
		t.Fatalf("expected range (%v, %v), got (%v, %v)", expectedFrom, expectedTill, from, till)
	}
}

func TestPrintStats(t *testing.T) {
	cache := setupTestDB(t)
	defer teardownTestDB(cache)

	ticker := "AAPL"
	dates := []string{"2023-01-01", "2023-01-02"}
	profits := []float64{100.5, 200.75}

	err := cache.AddProfits(ticker, dates, profits)
	if err != nil {
		t.Fatalf("failed to add profits: %v", err)
	}

	err = cache.PrintStats()
	if err != nil {
		t.Fatalf("failed to print stats: %v", err)
	}
}
