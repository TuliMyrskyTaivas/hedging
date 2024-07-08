package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/TuliMyrskyTaivas/hedging/hedging"
)

var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

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
	var command hedging.Command

	flag.StringVar(&command.Asset, "a", "", "base asset")
	flag.StringVar(&command.Hedge, "e", "", "hedge asset")
	flag.IntVar(&command.HistoryDepth, "d", 365, "history request depth")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&help, "h", false, "show help")
	flag.Parse()

	if help {
		fmt.Printf("Usage: %s [OPTIONS] command\n", os.Args[0])
		fmt.Printf("\tpossible commands: beta, hedge\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		log.Fatal("Command not speficied. Run with -h for the help")
	}

	setupLogger(verbose)

	buildInfo, _ := debug.ReadBuildInfo()
	slog.Debug(fmt.Sprintf("Built by %s at %s (SHA1=%s)", buildInfo.GoVersion, buildTime, sha1ver))

	executor, error := hedging.CreateCommand(flag.Arg(0))
	if error != nil {
		log.Fatal(error)
	}

	error = executor.Execute(command)
	if error != nil {
		log.Fatal(error)
	}
}
