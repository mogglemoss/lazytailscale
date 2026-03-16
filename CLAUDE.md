# lazytailscale

A terminal dashboard for your Tailscale network. Two-pane layout: peer list on
the left, selected-peer detail on the right. Keyboard-driven, SSH-launchable,
homelab-aware.

Built on the Charm stack: Bubbletea (Elm Architecture TUI framework), Bubbles
(component library), Lipgloss (styling DSL). Single Go binary, no runtime deps.

## Stack

- **Language:** Go 1.22+
- **TUI:** github.com/charmbracelet/bubbletea
- **Components:** github.com/charmbracelet/bubbles (list, viewport, textinput, spinner)
- **Styling:** github.com/charmbracelet/lipgloss
- **Tailscale data:** tailscale.com/client/tailscale (LocalClient — no API key needed)
- **Package manager:** Go modules (go mod)

## Project structure

```
lazytailscale/
├── CLAUDE.md
├── go.mod
├── go.sum
├── main.go               # Entry point, tea.NewProgram
├── model/
│   ├── model.go          # Root model: Init / Update / View
│   ├── keys.go           # Keybindings (key.Binding)
│   └── messages.go       # All tea.Msg types
├── ui/
│   ├── peerlist.go       # Left pane: peer list component
│   ├── detail.go         # Right pane: detail panel component
│   ├── statusbar.go      # Top bar: network name, self IP, online status
│   ├── helpbar.go        # Bottom bar: keybinding hints
│   └── styles.go         # All lipgloss styles, single source of truth
├── tailscale/
│   ├── client.go         # LocalClient wrapper, polling logic
│   └── types.go          # Internal peer/network types (mapped from tstype)
├── ping/
│   └── ping.go           # Async ping via tailscale ping, sparkline history
└── ssh/
    └── launch.go         # os/exec ssh into selected peer via Tailscale IP
```

## Architecture

lazytailscale uses the Elm Architecture strictly:

```
Init() → initial model + fetch peers Cmd
Update(msg) → new model + optional Cmd
View(model) → rendered string
```

All async work returns via `tea.Cmd`. Nothing blocks Update.

### Message flow

```
tickMsg (1s)     → triggers fetchPeersCmd
fetchPeersCmd    → returns peersLoadedMsg
peersLoadedMsg   → updates model.peers, triggers pingCmd for selected peer
pingCmd          → returns pingResultMsg
pingResultMsg    → updates sparkline ring buffer
windowSizeMsg    → recalculates layout dimensions
```

### Layout

```
╭─ statusbar (1 line) ──────────────────────────────────────────╮
│ ◈ lazytailscale   magpie-cherimoya.ts.net · 100.64.0.1 · ● online │
├─ peers (fixed 28 cols) ──┬─ detail (flex) ────────────────────┤
│ ● mollusk          NAS   │ mollusk                             │
│ ● mag-pi           Pi 5  │ 100.64.0.7 · mollusk.magpie-...    │
│ ● cloud-machine    Mac   │                                     │
│ ...                      │ routes / ports / ping sparkline     │
├──────────────────────────┴─────────────────────────────────────┤
│ helpbar (1 line): ↑↓ navigate  enter ssh  p ping  r routes ... │
╰────────────────────────────────────────────────────────────────╯
```

Left pane width: 28 columns fixed. Right pane: terminal width minus 28 minus
border. Heights recalculated on every windowSizeMsg.

## Data source

Use `tailscale.com/client/tailscale.LocalClient` — it talks to the local
tailscaled socket and requires no API key or network access. Key methods:

```go
lc := &tailscale.LocalClient{}
st, err := lc.Status(ctx)          // *ipnstate.Status — all peer info
pr, err := lc.Ping(ctx, ip, ping.TSMP)  // ping a peer
```

`st.Peer` is a `map[key.NodePublic]*ipnstate.PeerStatus`. Map over it to build
the internal peer list. Sort by: online first, then alphabetical by hostname.

Internal peer type (types.go):

```go
type Peer struct {
    Hostname     string
    TailscaleIP  string        // first IPv4 from TailscaleIPs
    DNSName      string        // trimmed trailing dot
    OS           string
    Online       bool
    LastSeen     time.Time
    AdvertisedRoutes []netip.Prefix
    AllowedIPs   []netip.Prefix
    TailscaleVersion string
    PingHistory  []time.Duration  // ring buffer, last 8 pings
    Tags         []string
}
```

## Peer list

- Built on `bubbles/list` with a custom `list.DefaultDelegate`
- Status dot: ● green = online, ● amber = idle (seen < 5min), ● red = offline,
  ● gray = unknown
- Show hostname (truncated to 18 chars) + OS tag right-aligned
- Filter mode: `/` activates list's built-in filtering

## Detail panel

Built as a `bubbles/viewport` so long content scrolls. Sections rendered top to
bottom with lipgloss borders between them:

1. **Header** — hostname, Tailscale IP, MagicDNS name
2. **Meta row** — OS, Tailscale version, last seen (relative: "2m ago")
3. **Routes** — advertised routes with approved/pending status
4. **Ping** — sparkline (8 bars, last 8 ping results) + avg/min/max
5. **Tags** — ACL tags if present

Sparkline bars: use braille block characters `▁▂▃▄▅▆▇█` scaled to min/max of
the window. Color: green if avg < 10ms, amber < 50ms, red ≥ 50ms.

## Key bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous peer |
| `↓` / `j` | Next peer |
| `enter` | SSH into selected peer (`ssh <tailscale-ip>`) |
| `p` | Force ping selected peer now |
| `r` | Toggle routes expanded view |
| `c` | Copy selected peer's Tailscale IP to clipboard |
| `/` | Filter peer list |
| `R` | Refresh peer list immediately |
| `?` | Toggle full help |
| `q` / `ctrl+c` | Quit |

SSH launch (ssh/launch.go): `exec.Command("ssh", ip).Run()` with the program
suspended via `tea.ExecProcess` so the terminal hands off cleanly and resumes
lazytailscale on exit.

Clipboard: use `golang.design/x/clipboard` or shell out to `pbcopy` / `xclip` /
`wl-copy` depending on platform.

## Styling (styles.go)

Single `styles.go` defines all lipgloss styles. No inline style calls elsewhere.
All colors live in a `Theme` struct so a `--theme` flag can be added later
without touching anything else.

### Theme struct

```go
type Theme struct {
    // Status colors
    Online  lipgloss.Color
    Idle    lipgloss.Color
    Offline lipgloss.Color
    Unknown lipgloss.AdaptiveColor

    // Chrome
    Accent       lipgloss.Color        // logo, section headers, ping sparkline (fast)
    AccentSubtle lipgloss.AdaptiveColor // secondary accent, borders
    Selected     lipgloss.AdaptiveColor // selected row background
    Border       lipgloss.AdaptiveColor // panel dividers
    TextPrimary  lipgloss.AdaptiveColor
    TextSecondary lipgloss.AdaptiveColor

    // Ping sparkline thresholds reuse Online / Idle / Offline
}

var DefaultTheme = Theme{ /* Charm Native — see below */ }
```

### Charm Native palette (default)

| Role | Dark terminal | Light terminal | Notes |
|---|---|---|---|
| Online | `#04B575` | `#028F5B` | Charm mint green |
| Idle | `#FFBF00` | `#B38600` | warm amber |
| Offline | `#FF5F87` | `#C4004F` | Charm hot pink (doubles as error) |
| Unknown | `#6C6C6C` / `#9A9A9A` | adaptive gray |
| Accent | `#FF5F87` | `#C4004F` | hot pink — logo, headings |
| AccentSubtle | `#7B61FF` / `#5B41DF` | soft purple — keybinding keys, tags |
| Selected bg | `#2D2B55` / `#DDD9FF` | purple-tinted row highlight |
| Border | `#3D3D3D` / `#CCCCCC` | panel dividers |
| TextPrimary | `#FFFDF5` / `#1A1A1A` | near-white / near-black |
| TextSecondary | `#B4B2A9` / `#6C6A62` | muted |

### Visual rules

- **Borders:** `lipgloss.RoundedBorder()` — `╭─╮` / `╰─╯` everywhere. No square corners.
- **Separators:** `·` (middle dot U+00B7) between fields in status bar, not `|`
- **Logo mark:** `◈` prefix in status bar: `◈ lazytailscale`
- **Status dots:** `●` for online/idle/offline, `○` for unknown
- **Ping sparkline:** braille block chars `▁▂▃▄▅▆▇█`, colored by avg latency
  - avg < 10ms → Online color (mint)
  - avg < 50ms → Idle color (amber)
  - avg ≥ 50ms → Offline color (pink/red)
  - failed ping → `✕` in accent color
- **Spinner:** Charm `dots` style, accent color, shown in status bar while pinging
- **Keybinding hints** in helpbar: key in AccentSubtle, description in TextSecondary
- **Section headers** in detail panel: uppercase, accent color, no border — just the color and spacing

Use `lipgloss.AdaptiveColor` for everything that must work in light terminals.
Hardcoded `lipgloss.Color` is fine for the vivid accent colors (they're designed
to pop on both backgrounds).

## Polling

Peers refresh every 5 seconds via `tea.Tick`. Ping for the selected peer runs
every 10 seconds and on demand (`p`). Store up to 8 ping results in a ring
buffer per peer. Do not ping all peers continuously — only the selected one.

## Error handling

- If LocalClient fails (tailscaled not running): show a centered error panel
  with message and retry hint, keep ticking
- Ping timeout: record as 0ms / failed, render as `✕` in sparkline
- SSH exec failure: show a one-line error in the status bar, clear after 3s

## Build & run

```bash
go build -o lazytailscale .
./lazytailscale

# or
go run .
```

Requires tailscaled running locally. On Linux needs access to
`/var/run/tailscale/tailscaled.sock` — run as the user who owns the Tailscale
session or with appropriate permissions.

## Future / out of scope for v1

- Network map / graph view (nodes + relay vs direct edges)
- DERP relay detection (direct vs relayed connection per peer)
- ACL viewer
- Exit node management
- Multi-tailnet support
