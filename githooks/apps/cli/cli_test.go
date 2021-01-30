package main

import (
	"flag"
	"gabyx/githooks/coverage"

	"testing"
)

var githooksCoverage *bool

func init() { //nolint: gochecknoinits
	githooksCoverage = flag.Bool("githooksCoverage", false, "Set to true when running coverage")
}

func TestCoverage(t *testing.T) {

	if *githooksCoverage {

		coverage.Setup("cli")

		// Run the main binary...
		mainRun()
	}
}
