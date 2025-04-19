package hedging

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"gonum.org/v1/gonum/stat"
)

type betaCalculator struct {
	cache *Cache
}

// ////////////////////////////////////////////////////////
// Constructor
// ////////////////////////////////////////////////////////
func newBetaCalculator() (Executor, error) {
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}

	err = cache.PrintStats()
	if err != nil {
		return nil, err
	}

	return &betaCalculator{cache: cache}, nil
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

	// Create report file if requested
	var report *Report = nil
	if len(command.Report) > 0 {
		var err error
		report, err = NewReport(command.Report)
		if err != nil {
			return fmt.Errorf("failed to create report file: %s", err)
		}
		defer report.Close()
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
		if len(assets) == len(assetNames) && index.Secid != "" {
			break
		}
	}

	// Calculate beta on assets
	betaResults := make(chan betaReport, len(assetNames))
	for _, asset := range assets {
		go calcBeta(asset, index, command.HistoryDepth, calculator.cache, betaResults, errors)
	}
	// Read the results of calculation or stop on first error
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

	// Print the results
	printer, err := GetPrinter()
	if err != nil {
		return err
	}
	for _, beta := range betas {
		if report != nil {
			err = report.AddReport(beta.asset, index.Secid, beta.beta, time.Now().Format("2006-01-02 15:04:05"))
			if err != nil {
				return fmt.Errorf("failed to add line to report: %s", err)
			}
		}
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

func calcBeta(asset moex.Asset, index moex.Asset, depthMonth int, cache *Cache, result chan betaReport, errResult chan error) {
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

	// Normalize histories: leave only items with the same date
	if len(indexHistory) > len(assetHistory) {
		indexHistory = filterHistory(assetHistory, indexHistory)
	} else if len(assetHistory) > len(indexHistory) {
		assetHistory = filterHistory(indexHistory, assetHistory)
	}

	slog.Debug(fmt.Sprintf("history of %s contains %d items, history of %s contains %d items",
		asset.Secid, len(assetHistory), index.Secid, len(indexHistory)))

	indexProfits := getProfits(indexHistory)
	assetProfits := getProfits(assetHistory)

	if profitsCalculatedWrong(indexProfits) {
		slog.Debug("seems that MOEX reports same open and close prices for the index, recalculating profits...")
		indexProfits = getOvernightProfits(indexHistory)
	}

	saveProfits(cache, asset.Secid, assetHistory, assetProfits)
	saveProfits(cache, index.Secid, indexHistory, indexProfits)

	indexStdDev := stat.StdDev(indexProfits, nil)
	beta := stat.Covariance(indexProfits, assetProfits, nil) / (indexStdDev * indexStdDev)
	result <- betaReport{asset.Secid, beta}
}

// ////////////////////////////////////////////////////////
// Get profits as difference of close and open prices
// ////////////////////////////////////////////////////////
func filterHistory(baseLine []moex.HistoryItem, input []moex.HistoryItem) []moex.HistoryItem {
	output := input[:0]
	for i, j := 0, 0; i < len(baseLine) && j < len(input); {
		baseDate := moex.ParseTime(baseLine[i].Tradedate)
		inputDate := moex.ParseTime(input[j].Tradedate)
		if baseDate == inputDate {
			output = append(output, input[i])
			i++
			j++
		} else if baseDate.Before(inputDate) {
			i++
		} else {
			j++
		}
	}

	// Validate
	for idx, item := range output {
		if baseLine[idx].Tradedate != item.Tradedate {
			slog.Error(fmt.Sprintf("Dates not equal at idx %d: %s != %s", idx, baseLine[idx].Tradedate, item.Tradedate))
		}
	}
	return output
}

// ////////////////////////////////////////////////////////
// Get profits as difference of close and open prices
// ////////////////////////////////////////////////////////
func getProfits(history []moex.HistoryItem) []float64 {
	var profits []float64
	for _, item := range history {
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
func getOvernightProfits(history []moex.HistoryItem) []float64 {
	var profits []float64
	var prevClose float64 = 0
	for _, item := range history {
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

func saveProfits(cache *Cache, asset string, history []moex.HistoryItem, profits []float64) {
	var dates []string
	for _, item := range history {
		dates = append(dates, item.Tradedate)
	}
	cache.AddProfits(asset, dates, profits)
}
