package moex

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"
)

// ///////////////////////////////////////////////////////////////////
func ParseTime(moexTime string) time.Time {
	const timeFormat string = "2006-01-02"
	time, err := time.Parse(timeFormat, moexTime)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse date: %w", err))
	}
	return time
}

// ///////////////////////////////////////////////////////////////////
func parseJSON[T any](s []byte) (T, error) {
	var r T
	if err := json.Unmarshal(s, &r); err != nil {
		slog.Error(fmt.Sprintf("failed to unmarshal JSON response: %s", err.Error()))
		return r, err
	}
	return r, nil
}

// ///////////////////////////////////////////////////////////////////
func query[T any](url string) (T, error) {
	var result T

	slog.Debug(fmt.Sprintf("Query MOEX: %s", url))
	res, err := http.Get(url)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to query MOEX: %s", err.Error()))
		return result, err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to read response from MOEX: %s", err.Error()))
		return result, err
	}
	return parseJSON[T](body)
}
