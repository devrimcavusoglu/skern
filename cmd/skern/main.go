// Package main is the entry point for the skern CLI.
package main

import (
	"os"

	"github.com/devrimcavusoglu/skern/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
