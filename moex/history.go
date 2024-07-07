package moex

import (
	"fmt"
	"log"
	"log/slog"
	"time"
)

type HistoryRange []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Dates []struct {
		From string `json:"from"`
		Till string `json:"till"`
	} `json:"dates,omitempty"`
}

type HistoryItem struct {
	Boardid           string  `json:"BOARDID"`
	Tradedate         string  `json:"TRADEDATE"`
	Secid             string  `json:"SECID"`
	Open              float64 `json:"OPEN"`
	Low               float64 `json:"LOW"`
	High              float64 `json:"HIGH"`
	Close             float64 `json:"CLOSE"`
	Openpositionvalue float64 `json:"OPENPOSITIONVALUE"`
	Value             float64 `json:"VALUE"`
	Volume            int     `json:"VOLUME"`
	Openposition      int     `json:"OPENPOSITION"`
	Settleprice       float64 `json:"SETTLEPRICE"`
	Swaprate          float64 `json:"SWAPRATE"`
	Waprice           float64 `json:"WAPRICE"`
	Settlepriceday    float64 `json:"SETTLEPRICEDAY"`
	Change            float64 `json:"CHANGE"`
	Qty               int     `json:"QTY"`
	Numtrades         int     `json:"NUMTRADES"`
}

type History []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	History       []HistoryItem `json:"history,omitempty"`
	HistoryCursor []struct {
		Index    int `json:"INDEX"`
		Total    int `json:"TOTAL"`
		Pagesize int `json:"PAGESIZE"`
	} `json:"history.cursor,omitempty"`
}

// ///////////////////////////////////////////////////////////////////
// Query MOEX on the dates for which history is available for the specified asset
// ///////////////////////////////////////////////////////////////////
func (asset *Asset) GetHistoryRange() (time.Time, time.Time) {
	slog.Debug(fmt.Sprintf("Quering MOEX on history range for %s", asset.Secid))
	url := fmt.Sprintf("https://iss.moex.com/iss/history/engines/%s/markets/%s/boards/%s/securities/%s/dates.json?iss.json=extended&iss.meta=off&marketprice_board=1",
		asset.Engine, asset.Market, asset.Boardid, asset.Secid)

	historyRange, err := query[HistoryRange](url)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to query MOEX: %w", err))
	}

	from := historyRange[1].Dates[0].From
	till := historyRange[1].Dates[0].Till

	slog.Debug(fmt.Sprintf("MOEX history for %s is available from %s till %s", asset.Secid, from, till))
	return ParseTime(from), ParseTime(till)
}

// ///////////////////////////////////////////////////////////////////
// Get MOEX asset history
// ///////////////////////////////////////////////////////////////////
func (asset *Asset) GetHistory(from time.Time, to time.Time) ([]HistoryItem, error) {
	const timeFormat string = "2006-01-02"
	timeFrom := from.Format(timeFormat)
	timeTo := to.Format(timeFormat)

	var result []HistoryItem
	start := 0

	for {
		slog.Debug(fmt.Sprintf("Quering MOEX history on %s from %s to %s (starting from %d)", asset.Secid, timeFrom, timeTo, start))

		url := fmt.Sprintf("https://iss.moex.com/iss/history/engines/%s/markets/%s/boards/%s/securities/%s.json?iss.json=extended&iss.meta=off&from=%s&till=%s&marketprice_board=1&start=%d",
			asset.Engine, asset.Market, asset.Boardid, asset.Secid, timeFrom, timeTo, start)

		history, err := query[History](url)
		if err != nil {
			return nil, err
		}

		result = append(result, history[1].History...)
		start = start + history[1].HistoryCursor[0].Pagesize
		if start > history[1].HistoryCursor[0].Total {
			break
		}

	}

	slog.Debug(fmt.Sprintf("MOEX history of %s contains %d items", asset.Secid, len(result)))
	return result, nil
}
