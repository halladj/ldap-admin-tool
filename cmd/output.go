package cmd

import (
	"fmt"
	"strings"
)

const bannerWidth = 45

// printBanner prints a formatted banner with a title and key-value pairs
func printBanner(title string, rows ...string) {
	fmt.Printf("\n%s\n  %s\n", strings.Repeat("=", bannerWidth), title)
	for i := 0; i+1 < len(rows); i += 2 {
		fmt.Printf("  %-10s : %s\n", rows[i], rows[i+1])
	}
	fmt.Printf("%s\n", strings.Repeat("=", bannerWidth))
}

// printProgress prints a banner with just a message
func printProgress(action string) {
	fmt.Printf("\n%s\n  %s\n", strings.Repeat("=", bannerWidth), action)
}

// printDone prints the bottom border of a banner
func printDone() {
	fmt.Printf("%s\n", strings.Repeat("=", bannerWidth))
}
