package hedging

import "fmt"

type betaCalculator struct{}

func (beta *betaCalculator) Execute(command Command) error {
	fmt.Printf("Calculate beta coefficient for %s using %s as market index\n", command.Asset, command.Hedge)
	return nil
}

func newBetaCalculator() Executor {
	return &betaCalculator{}
}