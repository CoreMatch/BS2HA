package main

import (
	"fmt"
	"os"

	"github.com/lnb/bs2ha/internal/app"
)

func main() {
	const version = "1.3.0"
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("BS2HA %s\n", version)
		return
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "bs2ha: %v\n", err)
		os.Exit(1)
	}
}
