// +build coverage

package main

import (
	"github.com/gabyx/githooks/githooks/apps/coverage"

	"testing"
)

func TestCoverage(t *testing.T) {
	if coverage.Setup("runner") {
		// Careful if you print to much stuff, certain tests might fail
		// fmt.Printf("Forward args: %q\n", os.Args)

		// Run the main binary...
		if mainRun() != 0 {
			t.Fatal()
		}
	}
}
