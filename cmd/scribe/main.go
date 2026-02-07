// Package main is the entry point for the scribe CLI.
package main

import (
	"os"

	"github.com/devrimcavusoglu/scribe/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
