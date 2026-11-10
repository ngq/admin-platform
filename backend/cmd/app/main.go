package main

import (
	"fmt"
	"os"

	svcsetup "github.com/ngq/admin-platform/internal/service/setup"
	gorp "github.com/ngq/gorp"
)

func main() {
	if handled, err := maybeHandleMigrateCLI(os.Args[1:]); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	projectRoot := svcsetup.ResolveProjectRoot()
	probe := probeDatabaseReadiness(projectRoot)
	if probe.needsSetup {
		completed, err := runSetupWizardServer(projectRoot, probe)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !completed {
			return
		}
	}

	if err := gorp.Run(gorp.WithMonoGovernance(), gorp.WithContractHTTPMode(), gorp.WithMigrate(migrate), gorp.WithSetup(setup)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
