package main

import (
	"os"

	"github.com/dev-pt-bai/cataloging/internal/pkg/runner"
)

func main() {
	os.Exit(runner.Run())
}
