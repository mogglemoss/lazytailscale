package main

import (
	"fmt"
	"github.com/mogglemoss/lazytailscale/model"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := model.New()
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
