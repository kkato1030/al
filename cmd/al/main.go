package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("al - Mac Management Tools")
		fmt.Println("Usage: al <command> [options]")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "version":
		fmt.Println("al version 0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
