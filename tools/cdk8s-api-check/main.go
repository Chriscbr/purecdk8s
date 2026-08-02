// Command cdk8s-api-check compares purecdk8s core with upstream cdk8s.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	apicheck "github.com/Chriscbr/purecdk8s/tools/api-check-common"
)

const (
	upstreamPackage = "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	localPackage    = "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

var replacements = map[string]apicheck.Replacement{
	"github.com/Chriscbr/purecdk8s/constructs/v10": {
		Path: "github.com/aws/constructs-go/constructs/v10",
		Name: "constructs",
	},
	"github.com/Chriscbr/purecdk8s/jsii": {
		Path: "github.com/aws/jsii-runtime-go",
		Name: "jsii",
	},
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <source-dir>")
		os.Exit(2)
	}

	sourceDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		fatal(err)
	}
	report, err := apicheck.Check(apicheck.Options{
		UpstreamPackage: upstreamPackage,
		LocalPackage:    localPackage,
		SourceDir:       sourceDir,
		Replacements:    replacements,
	})
	if err != nil {
		fatal(err)
	}
	if !apicheck.HasIncompatibleChanges(report) {
		fmt.Printf("cdk8s API is compatible with %s.\n", upstreamPackage)
		return
	}

	fmt.Fprintf(os.Stderr, "cdk8s API differs from %s:\n", upstreamPackage)
	if err := report.TextIncompatible(os.Stderr, false); err != nil {
		fatal(err)
	}
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
