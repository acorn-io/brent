package main

import (
	"github.com/acorn-io/brent/pkg/server/cli"
	"github.com/acorn-io/cmd"
)

func main() {
	cmd.Main(cli.NewBrent())
}
