package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	fmt.Print("not yet implemented")

	return nil
}
