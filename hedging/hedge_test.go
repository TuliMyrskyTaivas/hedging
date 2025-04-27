package hedging

import (
	"testing"

	"github.com/TuliMyrskyTaivas/hedging/moex"
	"github.com/stretchr/testify/assert"
)

func TestNewHedgeCalculator(t *testing.T) {
	calculator, err := newHedgeCalculator()
	assert.NoError(t, err)
	assert.NotNil(t, calculator)
}

func TestExecuteWithMissingHedge(t *testing.T) {
	calculator := &hedgeCalculator{}
	command := Command{
		Asset:        "SBER",
		Hedge:        "",
		HistoryDepth: 6,
	}
	err := calculator.Execute(command)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hedge asset was not specified")
}

func TestExecuteWithInvalidHedge(t *testing.T) {
	calculator := &hedgeCalculator{}
	command := Command{
		Asset:        "SBER",
		Hedge:        "INVALID_HEDGE",
		HistoryDepth: 6,
	}

	err := calculator.Execute(command)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "asset INVALID_HEDGE not found on MOEX")
}

func TestExtractPriceChanges(t *testing.T) {
	history := []moex.HistoryItem{
		{Close: 100},
		{Close: 105},
		{Close: 110},
	}
	expectedChanges := []float64{100, 5, 5}
	changes := extractPriceChanges(history)
	assert.Equal(t, expectedChanges, changes)
}

func TestExecuteWithValidData(t *testing.T) {
	calculator := &hedgeCalculator{}
	command := Command{
		Asset:        "SBER",
		Hedge:        "GAZP",
		HistoryDepth: 6,
	}

	err := calculator.Execute(command)
	assert.NoError(t, err)
}
