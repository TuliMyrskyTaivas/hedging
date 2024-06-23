package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"gonum.org/v1/gonum/stat"
)

// ///////////////////////////////////////////////////////////////////
// Extract array of price changes from MoexHistory
// ///////////////////////////////////////////////////////////////////
func extractPriceChanges(history moex.History) []float64 {
	var changes []float64
	var prevClose float64 = 0
	for _, item := range history[1].History {
		changes = append(changes, item.Close-prevClose)
		prevClose = item.Close
	}
	return changes
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
	var asset string

	flag.StringVar(&asset, "a", "", "base asset")
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

	// Use future's underlying asset if base asset is not explicitly defined
	if len(asset) == 0 {
		asset = moex.GetFutureUnderlyingAsset(futureTicker)
		fmt.Printf("Underlying asset for %s is %s\n", futureTicker, asset)
	}
	assetEngine, assetMarket, assetBoard := moex.GetEngineMarketBoard(asset)

	// Query range: last year
	historyTo := time.Now()
	historyFrom := historyTo.AddDate(-1, 0, 0)

	// Adjust range on availability of data on MOEX
	futureHistoryFrom, _ := moex.GetHistoryRange("futures", "forts", "RFUD", futureTicker)
	assetHistoryFrom, _ := moex.GetHistoryRange(assetEngine, assetMarket, assetBoard, asset)
	if futureHistoryFrom.After(historyFrom) {
		historyFrom = futureHistoryFrom
	}
	if assetHistoryFrom.After(historyFrom) {
		historyFrom = assetHistoryFrom
	}

	futureHistory := moex.GetHistory("futures", "forts", "RFUD", futureTicker, historyFrom, historyTo)
	futureChanges := extractPriceChanges(futureHistory)
	futureStdDev := stat.StdDev(futureChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", futureTicker, futureStdDev)

	assetHistory := moex.GetHistory(assetEngine, assetMarket, assetBoard, asset, historyFrom, historyTo)
	assetChanges := extractPriceChanges(assetHistory)
	assetStdDev := stat.StdDev(assetChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", asset, assetStdDev)

	correlation := stat.Correlation(futureChanges, assetChanges, nil)
	fmt.Printf("Correlation between price changes of %s and %s: %f\n", futureTicker, asset, correlation)
}
