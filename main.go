package main

import (
	"flag"
	"fmt"
	"github.com/mogglemoss/lazytailscale/model"
	"github.com/mogglemoss/lazytailscale/server"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time by GoReleaser via ldflags.
var version = "dev"

func main() {
	demo := flag.Bool("demo", false, "run with fictional demo data (no tailscaled required)")
	ver := flag.Bool("version", false, "print version and exit")
	serve := flag.Bool("serve", false, "serve the TUI over SSH using Wish")
	port := flag.Int("port", 23234, "SSH server port (used with --serve)")
	host := flag.String("host", "0.0.0.0", "SSH server bind address (used with --serve)")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	if *serve {
		if err := server.Start(*host, *port); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
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
