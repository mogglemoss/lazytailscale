package main

import (
	"flag"
	"fmt"
	"math/rand"
	"github.com/mogglemoss/lazytailscale/model"
	"github.com/mogglemoss/lazytailscale/server"
	"github.com/mogglemoss/lazytailscale/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var farewells = [][2]string{
	{"lazytailscale has concluded its inquiry.", "the substrate persists without it."},
	{"signal lost. lazytailscale withdraws from the mesh.", "nodes continue their quiet communion."},
	{"observation terminated. the network remains indifferent.", "it was watched. briefly."},
	{"lazytailscale has logged its final packet and departed.", "connectivity endures."},
	{"session dissolved. the tailnet carries on.", "it always does."},
	{"lazytailscale blinks out. the mesh does not notice.", "this is fine."},
	{"inquiry suspended indefinitely.", "the peers hold their positions."},
	{"monitoring ceased. the substrate hums on without witness.", "as it prefers."},
	{"lazytailscale has exited the lattice.", "the nodes remember nothing."},
	{"dashboard offline. the network requires no dashboard to exist.", "it simply is."},
}

// version is set at build time by GoReleaser via ldflags.
var version = "dev"

func main() {
	demo := flag.Bool("demo", false, "run with fictional demo data (no tailscaled required)")
	ver := flag.Bool("version", false, "print version and exit")
	serve := flag.Bool("serve", false, "serve the TUI over SSH using Wish")
	port := flag.Int("port", 23234, "SSH server port (used with --serve)")
	host := flag.String("host", "0.0.0.0", "SSH server bind address (used with --serve)")
	theme := flag.String("theme", "", "color theme: default, catppuccin, dracula, nord")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	if *theme != "" {
		ui.SetTheme(*theme)
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

	msg := farewells[rand.Intn(len(farewells))]
	fmt.Println()
	fmt.Println("  ◈  " + msg[0])
	fmt.Println("     " + msg[1])
	fmt.Println()
}
