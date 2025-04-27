package hedging

import "fmt"

type Command struct {
	Asset        string
	Hedge        string
	HistoryDepth int
	Report       string
}

type Executor interface {
	Execute(command Command) error
}

func CreateCommand(commandName string) (Executor, error) {
	if commandName == "beta" {
		return newBetaCalculator()
	}
	if commandName == "hedge" {
		return newHedgeCalculator()
	}
	return nil, fmt.Errorf("wrong command %s, run with -h for the help", commandName)
}
