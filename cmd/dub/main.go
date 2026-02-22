// cmd/dub/main.go
package main

import (
	"os"

	"github.com/salmonumbrella/dub-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		if cmd.IsUsageError(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
