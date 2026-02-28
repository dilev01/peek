package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dilev01/peek/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("peek v%s\n", version)
		return
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: peek <file.md>")
		os.Exit(1)
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	m := tui.NewModel(string(content))
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
