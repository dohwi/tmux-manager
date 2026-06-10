package main

import (
	"os"

	"github.com/dohwi/tmux-manager/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
