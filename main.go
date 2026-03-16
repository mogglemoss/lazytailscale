package main

import (
	"flag"
	"fmt"
	"github.com/mogglemoss/lazytailscale/model"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	demo := flag.Bool("demo", false, "run with fictional demo data (no tailscaled required)")
	flag.Parse()

	m := model.New(*demo)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
