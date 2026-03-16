package main

import (
	"flag"
	"fmt"
	"github.com/mogglemoss/lazytailscale/model"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time by GoReleaser via ldflags.
var version = "dev"

func main() {
	demo := flag.Bool("demo", false, "run with fictional demo data (no tailscaled required)")
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

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
