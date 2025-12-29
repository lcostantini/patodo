package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	store, err := NewTaskStore()
	if err != nil {
		fmt.Printf("Error initializing task store: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(store), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
