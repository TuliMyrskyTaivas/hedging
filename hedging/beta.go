package hedging

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"gonum.org/v1/gonum/stat"
)

type betaCalculator struct{}

func newBetaCalculator() Executor {
	return &betaCalculator{}
}

func (calculator *betaCalculator) Execute(command Command) error {
	fmt.Printf("Calculate beta coefficient for %s using %s as market index\n", command.Asset, command.Hedge)

	if len(command.Hedge) == 0 {
		return fmt.Errorf("index was not specified. Run with -h for the help")
	}

	if len(command.Asset) == 0 {
		return fmt.Errorf("asset was not specified. Run with -h for the help")
	}

	index, err := moex.GetAsset(command.Hedge)
	if err != nil {
		return err
	}
	asset, err := moex.GetAsset(command.Asset)
	if err != nil {
		return err
	}

	historyTo := time.Now()
	historyFrom := historyTo.AddDate(0, -command.HistoryDepth, 0)
	indexHistoryBegin := moex.ParseTime(index.HistoryFrom)
	assetHistoryBegin := moex.ParseTime(asset.HistoryFrom)

	// Adjust range on availability of data on MOEX
	if indexHistoryBegin.After(historyFrom) {
		historyFrom = indexHistoryBegin
	}
	if assetHistoryBegin.After(historyFrom) {
		historyFrom = assetHistoryBegin
	}

	indexHistory, err := index.GetHistory(historyFrom, historyTo)
	if err != nil {
		return err
	}

	assetHistory, err := asset.GetHistory(historyFrom, historyTo)
	if err != nil {
		return err
	}

	shortestLength := min(len(indexHistory), len(assetHistory))
	indexProfits := getProfits(indexHistory, shortestLength)
	assetProfits := getProfits(assetHistory, shortestLength)

	if profitsCalculatedWrong(indexProfits) {
		slog.Debug("seems that MOEX reports same open and close prices for the index, recalculating profits...")
		indexProfits = getOvernightProfits(indexHistory, shortestLength)
	}

	indexStdDev := stat.StdDev(indexProfits, nil)
	beta := stat.Covariance(indexProfits, assetProfits, nil) / (indexStdDev * indexStdDev)

	fmt.Printf("Beta coefficient for %s on %s is %f\n", asset.Secid, index.Secid, beta)
	return nil
}

func getProfits(history []moex.HistoryItem, length int) []float64 {
	var profits []float64
	for idx, item := range history {
		if idx == length {
			break
		}
		if item.Open == 0 {
			profits = append(profits, 0)
		} else {
			profits = append(profits, (item.Close-item.Open)/item.Open)
		}
	}
	return profits
}

func getOvernightProfits(history []moex.HistoryItem, length int) []float64 {
	var profits []float64
	var prevClose float64 = 0
	for idx, item := range history {
		if idx == length {
			break
		}
		if prevClose == 0 {
			profits = append(profits, 0)
		} else {
			profits = append(profits, (item.Close-prevClose)/prevClose)
		}
		prevClose = item.Close
	}
	return profits
}

func profitsCalculatedWrong(profits []float64) bool {
	var sum float64
	for _, item := range profits {
		sum += item
	}

	return sum == 0
}
