// Command cdk8splus-api-check compares a purecdk8s+ package with the
// corresponding upstream cdk8s-plus-go API.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	apicheck "github.com/Chriscbr/purecdk8s/tools/api-check-common"
)

const (
	upstreamPackage = "github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2"
	localModule     = "github.com/Chriscbr/purecdk8s"
)

var replacements = map[string]apicheck.Replacement{
	"github.com/Chriscbr/purecdk8s/cdk8s/v2": {
		Path: "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2",
		Name: "cdk8s",
	},
	"github.com/Chriscbr/purecdk8s/constructs/v10": {
		Path: "github.com/aws/constructs-go/constructs/v10",
		Name: "constructs",
	},
	"github.com/Chriscbr/purecdk8s/jsii": {
		Path: "github.com/aws/jsii-runtime-go",
		Name: "jsii",
	},
	"github.com/Chriscbr/purecdk8s/cdk8splus34/v2": {
		Path: upstreamPackage,
		Name: "cdk8splus34",
	},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run . <kubernetes-minor-version> <source-dir>")
		os.Exit(2)
	}

	minorVersion := os.Args[1]
	if _, err := strconv.Atoi(minorVersion); err != nil {
		fmt.Fprintln(os.Stderr, "kubernetes-minor-version must be numeric")
		os.Exit(2)
	}

	sourceDir, err := filepath.Abs(os.Args[2])
	if err != nil {
		fatal(err)
	}
	report, err := apicheck.Check(apicheck.Options{
		UpstreamPackage: upstreamPackage,
		LocalPackage:    fmt.Sprintf("%s/cdk8splus%s/v2", localModule, minorVersion),
		SourceDir:       sourceDir,
		Replacements:    replacements,
	})
	if err != nil {
		fatal(err)
	}
	if !apicheck.HasIncompatibleChanges(report) {
		fmt.Printf("cdk8splus%s API is compatible with %s.\n", minorVersion, upstreamPackage)
		return
	}

	fmt.Fprintf(os.Stderr, "cdk8splus%s API differs from %s:\n", minorVersion, upstreamPackage)
	if err := report.TextIncompatible(os.Stderr, false); err != nil {
		fatal(err)
	}
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
