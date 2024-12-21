package hedging

import "fmt"

type Command struct {
	Asset        string
	Hedge        string
	HistoryDepth int
}

type Executor interface {
	Execute(command Command) error
}

func CreateCommand(command string) (Executor, error) {
	if command == "beta" {
		return newBetaCalculator()
	}
	if command == "hedge" {
		return newHedgeCalculator()
	}
	return nil, fmt.Errorf("wrong command %s, run with -h for the help", command)
}
