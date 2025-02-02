package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/chavacava/dfence/internal/infra"
	"github.com/chavacava/dfence/internal/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var policyFile string

var cmdEnforce = &cobra.Command{
	Use:   "enforce",
	Short: "Enforce policy on given packages",
	Long:  "Check if the packages respect the dependencies policy.",
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		logger, ok := viper.Get("logger").(infra.Logger)
		if !ok {
			log.Fatal("Unable to retrieve the logger.") // revive:disable-line:deep-exit
		}

		var err error
		stream, err := os.Open(policyFile)
		if err != nil {
			logger.Fatalf("Unable to open policy file %s: %+v", policyFile, err)
		}

		policy, err := policy.NewPolicyFromJSON(stream)
		if err != nil {
			logger.Fatalf("Unable to load policy : %v", err) // revive:disable-line:deep-exit
		}

		const pkgSelector = "./..."
		logger.Infof("Retrieving packages...")
		pkgs, err := retrievePackages(pkgSelector)
		if err != nil {
			logger.Fatalf("Unable to retrieve packages using the selector '%s': %v", pkgSelector, err)
		}
		logger.Infof("Will work with %d package(s).", len(pkgs))

		err = check(policy, pkgs, logger)
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(1) // revive:disable-line:deep-exit
		}
	},
}

func init() {
	cmdPolicy.AddCommand(cmdEnforce)
	cmdEnforce.Flags().StringVar(&policyFile, "policy", "", "path to dependencies policy file ")
	cmdEnforce.MarkFlagRequired("policy")
}

func check(p policy.Policy, pkgs []string, logger infra.Logger) error {
	checker, err := policy.NewChecker(p, logger)
	if err != nil {
		logger.Fatalf("Unable to run the checker: %v", err)
	}

	pkgCount := len(pkgs)
	errCount := 0
	out := make(chan policy.CheckResult, pkgCount)
	for _, pkg := range pkgs {
		go checker.CheckPkg(pkg, out)
	}

	logger.Infof("Checking...")

	for i := 0; i < pkgCount; i++ {
		result := <-out
		for _, w := range result.Warns {
			logger.Warningf(w.Error())
		}
		for _, e := range result.Errs {
			logger.Errorf(e.Error())
		}

		errCount += len(result.Errs)
	}

	logger.Infof("Check done")

	if errCount > 0 {
		return fmt.Errorf("found %d error(s)", errCount)
	}

	return nil
}

// retrievePackages yields the all packages matching the given selector
func retrievePackages(pkgSelector string) ([]string, error) {
	r := []string{}
	cmd := exec.Command("go", "list", pkgSelector)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return r, errors.New(errStr)
	}

	r = strings.Split(outStr, "\n")

	return r[:len(r)-1], nil
}
