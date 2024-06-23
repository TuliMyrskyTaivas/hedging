package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"time"

	"gonum.org/v1/gonum/stat"
)

type MoexHistoryRange []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Dates []struct {
		From string `json:"from"`
		Till string `json:"till"`
	} `json:"dates,omitempty"`
}

type MoexHistory []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	History []struct {
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
	} `json:"history,omitempty"`
}

type MoexAssetInfo []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Securities []struct {
		Secid            string  `json:"SECID"`
		Boardid          string  `json:"BOARDID"`
		Shortname        string  `json:"SHORTNAME"`
		Secname          string  `json:"SECNAME"`
		Prevsettleprice  float64 `json:"PREVSETTLEPRICE"`
		Decimals         int     `json:"DECIMALS"`
		Minstep          float64 `json:"MINSTEP"`
		Lasttradedate    string  `json:"LASTTRADEDATE"`
		Lastdeldate      string  `json:"LASTDELDATE"`
		Sectype          string  `json:"SECTYPE"`
		Latname          string  `json:"LATNAME"`
		Assetcode        string  `json:"ASSETCODE"`
		Prevopenposition int     `json:"PREVOPENPOSITION"`
		Lotvolume        int     `json:"LOTVOLUME"`
		Initialmargin    float64 `json:"INITIALMARGIN"`
		Highlimit        float64 `json:"HIGHLIMIT"`
		Lowlimit         float64 `json:"LOWLIMIT"`
		Stepprice        float64 `json:"STEPPRICE"`
		Lastsettleprice  float64 `json:"LASTSETTLEPRICE"`
		Prevprice        float64 `json:"PREVPRICE"`
		Imtime           string  `json:"IMTIME"`
		Buysellfee       float64 `json:"BUYSELLFEE"`
		Scalperfee       float64 `json:"SCALPERFEE"`
		Negotiatedfee    float64 `json:"NEGOTIATEDFEE"`
		Exercisefee      float64 `json:"EXERCISEFEE"`
	} `json:"securities,omitempty"`
}

type MoexAssetDescription []struct {
	Charsetinfo struct {
		Name string `json:"name"`
	} `json:"charsetinfo,omitempty"`
	Boards []struct {
		Secid        string `json:"secid"`
		Boardid      string `json:"boardid"`
		Title        string `json:"title"`
		BoardGroupID int    `json:"board_group_id"`
		MarketID     int    `json:"market_id"`
		Market       string `json:"market"`
		EngineID     int    `json:"engine_id"`
		Engine       string `json:"engine"`
		IsTraded     int    `json:"is_traded"`
		Decimals     int    `json:"decimals"`
		HistoryFrom  string `json:"history_from"`
		HistoryTill  string `json:"history_till"`
		ListedFrom   string `json:"listed_from"`
		ListedTill   string `json:"listed_till"`
		IsPrimary    int    `json:"is_primary"`
		Currencyid   any    `json:"currencyid"`
	} `json:"boards,omitempty"`
}

// ///////////////////////////////////////////////////////////////////
// Extract array of price changes from MoexHistory
// ///////////////////////////////////////////////////////////////////
func extractPriceChanges(history MoexHistory) []float64 {
	var changes []float64
	var prevClose float64 = 0
	for _, item := range history[1].History {
		changes = append(changes, item.Close-prevClose)
		prevClose = item.Close
	}
	return changes
}

// ///////////////////////////////////////////////////////////////////
func parseMoexTime(moexTime string) time.Time {
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
func moexQuery[T any](url string) (T, error) {
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

// ///////////////////////////////////////////////////////////////////
// Query MOEX on engine, market and primary board for the specified asset
// ///////////////////////////////////////////////////////////////////
func getEngineMarketBoard(asset string) (string, string, string) {
	slog.Debug(fmt.Sprintf("Quering MOEX on engine/market for %s", asset))
	url := fmt.Sprintf("https://iss.moex.com/iss/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=boards", asset)
	assetDescription, err := moexQuery[MoexAssetDescription](url)
	if err != nil {
		log.Fatal("Failed to get asset description from MOEX")
	}

	info := assetDescription[1].Boards[0]
	if info.IsPrimary != 1 {
		log.Fatal("First board is not primary!")
	}

	slog.Debug("For %s on MOEX engine is %s, market is %s, primary board is %s", asset, info.Engine, info.Market, info.Boardid)
	return info.Engine, info.Market, info.Boardid
}

// ///////////////////////////////////////////////////////////////////
// Query MOEX on future's underlying asset code
// ///////////////////////////////////////////////////////////////////
func getFutureUnderlyingAsset(future string) string {
	slog.Debug(fmt.Sprintf("Quering MOEX on %s future", future))

	url := fmt.Sprintf("https://iss.moex.com/iss/engines/futures/markets/forts/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=securities",
		future)
	assetInfo, err := moexQuery[MoexAssetInfo](url)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to query MOEX: %w", err))
	}

	asset := assetInfo[1].Securities[0].Assetcode
	// For GLDRUBF future MOEX returns GLDRUBTOM instead of GLDRUB_TOM
	var re = regexp.MustCompile("(TOM)$")
	return re.ReplaceAllString(asset, "_TOM")
}

// ///////////////////////////////////////////////////////////////////
// Query MOEX on the dates for which history is available for the specified asset
// ///////////////////////////////////////////////////////////////////
func getHistoryRange(engine string, market string, board string, asset string) (time.Time, time.Time) {
	slog.Debug(fmt.Sprintf("Quering MOEX on history range for %s", asset))
	url := fmt.Sprintf("https://iss.moex.com/iss/history/engines/%s/markets/%s/boards/%s/securities/%s/dates.json?iss.json=extended&iss.meta=off&marketprice_board=1",
		engine, market, board, asset)

	historyRange, err := moexQuery[MoexHistoryRange](url)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to query MOEX: %w", err))
	}

	from := historyRange[1].Dates[0].From
	till := historyRange[1].Dates[0].Till

	slog.Debug(fmt.Sprintf("MOEX history for %s is available from %s till %s", asset, from, till))
	return parseMoexTime(from), parseMoexTime(till)
}

// ///////////////////////////////////////////////////////////////////
// Get MOEX asset history
// ///////////////////////////////////////////////////////////////////
func getHistory(engine string, market string, board string, asset string, from time.Time, to time.Time) MoexHistory {
	const timeFormat string = "2006-01-02"
	timeFrom := from.Format(timeFormat)
	timeTo := to.Format(timeFormat)

	slog.Debug(fmt.Sprintf("Quering MOEX history on %s from %s to %s", asset, timeFrom, timeTo))

	url := fmt.Sprintf("https://iss.moex.com/iss/history/engines/%s/markets/%s/boards/%s/securities/%s.json?iss.json=extended&iss.meta=off&iss.only=history&from=%s&till=%s&marketprice_board=1",
		engine, market, board, asset, timeFrom, timeTo)

	history, err := moexQuery[MoexHistory](url)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to query MOEX: %w", err))
	}

	slog.Debug(fmt.Sprintf("MOEX history of %s contains %d items", asset, len(history[1].History)))
	return history
}

// ///////////////////////////////////////////////////////////////////
// Setup a global logger
// ///////////////////////////////////////////////////////////////////
func setupLogger(verbose bool) {
	var log *slog.Logger

	if verbose {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	slog.SetDefault(log)
}

// ///////////////////////////////////////////////////////////////////
// Entry point
// ///////////////////////////////////////////////////////////////////
func main() {
	var verbose bool
	var help bool
	var futureTicker string
	var baseTicker string

	flag.StringVar(&baseTicker, "b", "", "base ticker")
	flag.StringVar(&futureTicker, "f", "", "future ticker")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&help, "h", false, "show help")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	setupLogger(verbose)

	buildInfo, _ := debug.ReadBuildInfo()
	slog.Debug(fmt.Sprintf("Built by %s", buildInfo.GoVersion))

	underlyingAsset := getFutureUnderlyingAsset(futureTicker)
	assetEngine, assetMarket, assetBoard := getEngineMarketBoard(underlyingAsset)
	fmt.Printf("Underlying asset for %s is %s\n", futureTicker, underlyingAsset)

	// Query range: last year
	historyTo := time.Now()
	historyFrom := historyTo.AddDate(-1, 0, 0)

	// Adjust range on availability of data on MOEX
	futureHistoryFrom, _ := getHistoryRange("futures", "forts", "RFUD", futureTicker)
	assetHistoryFrom, _ := getHistoryRange(assetEngine, assetMarket, assetBoard, underlyingAsset)
	if futureHistoryFrom.After(historyFrom) {
		historyFrom = futureHistoryFrom
	}
	if assetHistoryFrom.After(historyFrom) {
		historyFrom = assetHistoryFrom
	}

	futureHistory := getHistory("futures", "forts", "RFUD", futureTicker, historyFrom, historyTo)
	futureChanges := extractPriceChanges(futureHistory)
	futureStdDev := stat.StdDev(futureChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", futureTicker, futureStdDev)

	assetHistory := getHistory(assetEngine, assetMarket, assetBoard, underlyingAsset, historyFrom, historyTo)
	assetChanges := extractPriceChanges(assetHistory)
	assetStdDev := stat.StdDev(assetChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", underlyingAsset, assetStdDev)

	correlation := stat.Correlation(futureChanges, assetChanges, nil)
	fmt.Printf("Correlation between price changes of %s and %s: %f\n", futureTicker, underlyingAsset, correlation)
}
