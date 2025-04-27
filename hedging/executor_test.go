package hedging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCommandWithValidCommand(t *testing.T) {
	executor, err := CreateCommand("beta")
	assert.NoError(t, err)
	assert.NotNil(t, executor)

	executor, err = CreateCommand("hedge")
	assert.NoError(t, err)
	assert.NotNil(t, executor)
}

func TestCreateCommandWithInvalidCommand(t *testing.T) {
	executor, err := CreateCommand("invalid")
	assert.Error(t, err)
	assert.Nil(t, executor)
	assert.EqualError(t, err, "wrong command invalid, run with -h for the help")
}

func TestExecuteWithInvalidDataExecutor(t *testing.T) {
	calculator := &hedgeCalculator{}
	command := Command{
		Asset:        "",
		Hedge:        "",
		HistoryDepth: -1,
	}

	err := calculator.Execute(command)
	assert.Error(t, err)
}
