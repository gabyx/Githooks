//go:build coverage

package main

import (
	"github.com/gabyx/githooks/githooks/apps/coverage"
	cm "github.com/gabyx/githooks/githooks/common"

	"testing"
)

func TestCoverage(t *testing.T) {
	if coverage.Setup("githooks-cli") {
		// Careful if you print to much stuff, certain tests might fail
		// fmt.Printf("Forward args: %q\n", os.Args)

		// Run the main binary...
		var cleanUpX cm.InterruptContext
		exitCode := mainRun(&cleanUpX)
		cleanUpX.RunHandlers()

		if exitCode != 0 {
			t.Fatal()
		}
	}
}
