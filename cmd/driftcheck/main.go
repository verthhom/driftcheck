package main

import (
	"flag"
	"fmt"
	"os"

	"driftcheck/internal/config"
	"driftcheck/internal/drift"
	"driftcheck/internal/snapshot"
)

func main() {
	var (
		configDir = flag.String("config", "./configs", "directory containing service config files")
		output    = flag.String("output", "text", "output format: text or json")
	)
	flag.Parse()

	loader := config.NewLoader()
	configs, err := loader.LoadAll(*configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading configs: %v\n", err)
		os.Exit(1)
	}

	if len(configs) == 0 {
		fmt.Fprintln(os.Stderr, "no config files found in", *configDir)
		os.Exit(1)
	}

	detector := drift.NewDetector()
	reporter := drift.NewReporter(os.Stdout, *output)

	var hasAnyDrift bool

	for _, cfg := range configs {
		fetcher := snapshot.NewEnvFetcher(cfg.ServiceName)
		snap, err := snapshot.New(cfg, fetcher)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating snapshot for %s: %v\n", cfg.ServiceName, err)
			os.Exit(1)
		}

		result, err := detector.Check(cfg, snap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error checking drift for %s: %v\n", cfg.ServiceName, err)
			os.Exit(1)
		}

		if err := reporter.Report(result); err != nil {
			fmt.Fprintf(os.Stderr, "error reporting result: %v\n", err)
			os.Exit(1)
		}

		if result.HasDrift() {
			hasAnyDrift = true
		}
	}

	if hasAnyDrift {
		os.Exit(2)
	}
}
