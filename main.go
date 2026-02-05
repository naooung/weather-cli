package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	city := strings.Join(os.Args[1:], " ")

	if err := RunNow(city); err != nil {
		fail("failed: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  weather <city>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  weather seoul")
	fmt.Println(`  weather "new york"`)
}