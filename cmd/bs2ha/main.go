package main

import (
	"fmt"
	"os"

	"github.com/lnb/bs2ha/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "bs2ha: %v\n", err)
		os.Exit(1)
	}
}
