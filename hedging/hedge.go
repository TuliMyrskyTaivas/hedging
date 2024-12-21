package hedging

import (
	"fmt"
	"time"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"gonum.org/v1/gonum/stat"
)

type hedgeCalculator struct {
	cache *Cache
}

func newHedgeCalculator() (Executor, error) {
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}
	return &hedgeCalculator{cache: cache}, nil
}

func (calculator *hedgeCalculator) Execute(command Command) error {
	fmt.Printf("Calculate hedging coefficient for %s and %s\n", command.Asset, command.Hedge)

	if len(command.Hedge) == 0 {
		return fmt.Errorf("hedge asset was not specified. Run with -h for the help")
	}

	hedge, err := moex.GetAsset(command.Hedge)
	if err != nil {
		return err
	}

	// Use future's underlying asset if base asset is not explicitly defined
	var asset moex.Asset
	if len(command.Asset) == 0 {
		asset, err = hedge.GetFutureUnderlyingAsset()
		fmt.Printf("Underlying asset for %s is %s\n", hedge.Secid, asset.Secid)
	} else {
		asset, err = moex.GetAsset(command.Asset)
	}

	if err != nil {
		return err
	}

	historyTo := time.Now()
	historyFrom := historyTo.AddDate(0, -command.HistoryDepth, 0)
	hedgeHistoryBegin := moex.ParseTime(hedge.HistoryFrom)
	assetHistoryBegin := moex.ParseTime(asset.HistoryFrom)

	// Adjust range on availability of data on MOEX
	if hedgeHistoryBegin.After(historyFrom) {
		historyFrom = hedgeHistoryBegin
	}
	if assetHistoryBegin.After(historyFrom) {
		historyFrom = assetHistoryBegin
	}

	hedgeHistory, err := hedge.GetHistory(historyFrom, historyTo)
	if err != nil {
		return err
	}

	assetHistory, err := asset.GetHistory(historyFrom, historyTo)
	if err != nil {
		return err
	}

	hedgeChanges := extractPriceChanges(hedgeHistory)
	hedgeStdDev := stat.StdDev(hedgeChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", hedge.Secid, hedgeStdDev)

	assetChanges := extractPriceChanges(assetHistory)
	assetStdDev := stat.StdDev(assetChanges, nil)
	fmt.Printf("%s standard deviation: %f\n", asset.Secid, assetStdDev)

	correlation := stat.Correlation(hedgeChanges, assetChanges, nil)
	fmt.Printf("Correlation between price changes of %s and %s: %f\n", hedge.Secid, asset.Secid, correlation)

	optimalHedge := (assetStdDev / hedgeStdDev) * correlation
	hedgingEfficiency := correlation * correlation
	fmt.Printf("Optimal hedging coefficient is %f, hedging efficiency is %f\n", optimalHedge, hedgingEfficiency)

	return nil
}

// ///////////////////////////////////////////////////////////////////
// Extract array of price changes from MoexHistory
// ///////////////////////////////////////////////////////////////////
func extractPriceChanges(history []moex.HistoryItem) []float64 {
	var changes []float64
	var prevClose float64 = 0
	for _, item := range history {
		changes = append(changes, item.Close-prevClose)
		prevClose = item.Close
	}
	return changes
}
