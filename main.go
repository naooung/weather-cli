package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// 사용 예: weather now seoul
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	city := strings.Join(os.Args[2:], " ")

	switch cmd {
	case "now":
		if err := RunNow(city); err != nil {
			fail("failed: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  weather now <city>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println(`  weather now seoul`)
	fmt.Println(`  weather now "new york"`)
}