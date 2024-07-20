package hedging

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"gonum.org/v1/gonum/stat"
)

type betaCalculator struct{}

// ////////////////////////////////////////////////////////
// Constructor
// ////////////////////////////////////////////////////////
func newBetaCalculator() Executor {
	return &betaCalculator{}
}

type betaReport struct {
	asset string
	beta  float64
}

// ////////////////////////////////////////////////////////
// Command executor
// ////////////////////////////////////////////////////////
func (calculator *betaCalculator) Execute(command Command) error {
	fmt.Printf("Calculate beta coefficient for %s using %s as market index\n", command.Asset, command.Hedge)

	// Check inputs
	if len(command.Hedge) == 0 {
		return fmt.Errorf("index was not specified. Run with -h for the help")
	}

	if len(command.Asset) == 0 {
		return fmt.Errorf("asset was not specified. Run with -h for the help")
	}

	// Prepare channels
	assetNames := strings.Split(command.Asset, ",")
	assetResults := make(chan moex.Asset, len(assetNames)+1)
	errors := make(chan error, len(assetNames)+1)

	// Query info on index and assets
	go getAsset(command.Hedge, assetResults, errors)
	for _, asset := range assetNames {
		go getAsset(asset, assetResults, errors)
	}

	// Read results or stop if any error occurred
	var index moex.Asset
	var assets []moex.Asset
	for {
		select {
		case err := <-errors:
			return err
		case asset := <-assetResults:
			if asset.Secid == command.Hedge {
				index = asset
			} else {
				assets = append(assets, asset)
			}
		}
		if len(assets) == len(assetNames) {
			break
		}
	}

	printer, err := GetPrinter()
	if err != nil {
		return err
	}

	betaResults := make(chan betaReport, len(assetNames))
	for _, asset := range assets {
		go calcBeta(asset, index, command.HistoryDepth, betaResults, errors)
	}
	var betas []betaReport
	for {
		select {
		case err := <-errors:
			return err

		case beta := <-betaResults:
			betas = append(betas, beta)
		}
		if len(betas) == len(assetNames) {
			break
		}
	}

	for _, beta := range betas {
		printer.Printf("Beta coefficient for last %d month %s on %s is %f\n", command.HistoryDepth, beta.asset, index.Secid, beta.beta)
	}

	return nil
}

// ////////////////////////////////////////////////////////
// Get info on MOEX asset asynchronously
// ////////////////////////////////////////////////////////
func getAsset(assetName string, result chan moex.Asset, errResult chan error) {
	asset, err := moex.GetAsset(assetName)
	if err != nil {
		errResult <- err
	} else {
		result <- asset
	}
}

func calcBeta(asset moex.Asset, index moex.Asset, depthMonth int, result chan betaReport, errResult chan error) {
	// Adjust range on availability of data on MOEX
	historyTo := time.Now()
	historyFrom := historyTo.AddDate(0, -depthMonth, 0)
	indexHistoryBegin := moex.ParseTime(index.HistoryFrom)
	assetHistoryBegin := moex.ParseTime(asset.HistoryFrom)

	if indexHistoryBegin.After(historyFrom) {
		historyFrom = indexHistoryBegin
	}
	if assetHistoryBegin.After(historyFrom) {
		historyFrom = assetHistoryBegin
	}

	indexHistory, err := index.GetHistory(historyFrom, historyTo)
	if err != nil {
		errResult <- err
		return
	}

	assetHistory, err := asset.GetHistory(historyFrom, historyTo)
	if err != nil {
		errResult <- err
		return
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
	result <- betaReport{asset.Secid, beta}
}

// ////////////////////////////////////////////////////////
// Get profits as difference of close and open prices
// ////////////////////////////////////////////////////////
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

// ////////////////////////////////////////////////////////
// Get profits as difference of close prices overnight
// ////////////////////////////////////////////////////////
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

// ////////////////////////////////////////////////////////
// Check whether profits were calculated wrong (sum is zero)
// ////////////////////////////////////////////////////////
func profitsCalculatedWrong(profits []float64) bool {
	var sum float64
	for _, item := range profits {
		sum += item
	}

	return sum == 0
}
