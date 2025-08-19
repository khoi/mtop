package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command line flags
	jsonMode := flag.Bool("json", false, "Output system stats in JSON format instead of TUI")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mtop - System monitor for macOS\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s           Start interactive TUI mode\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json    Output current stats as JSON\n", os.Args[0])
	}
	flag.Parse()

	if *jsonMode {
		// JSON output mode
		stats, err := collectSystemStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting system stats: %v\n", err)
			os.Exit(1)
		}

		jsonData, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
		return
	}

	// TUI mode
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
