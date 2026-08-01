package main

import (
	"os"

	"github.com/purecdk8s/purecdk8s/internal/cli"
)

var version = cli.DefaultVersion

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.Options{
		Name:    "purecdk8s",
		Version: version,
	}))
}
