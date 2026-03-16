# lazytailscale

![License: MIT](https://img.shields.io/badge/license-MIT-pink.svg)
![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)
![Built with Charm](https://raw.githubusercontent.com/charmbracelet/charm/main/title-dark.svg)

A terminal dashboard for your Tailscale network. Two-pane keyboard-driven TUI: peer list on the left, selected-peer detail on the right. Runs entirely from your local Tailscale socket — no API key, no cloud, no opinions about your network topology.

---

![lazytailscale](./screenshot.png)

## What It Does

- Displays all peers on your tailnet in a scrollable list, sorted by status then alphabetically
- Shows per-peer detail: Tailscale IP, MagicDNS name, OS, connection type (direct or relayed), last WireGuard handshake, advertised subnet routes, ACL tags
- Pings the selected peer every 10 seconds and renders the history as a sparkline — because a 3ms round trip deserves to be seen
- Launches SSH sessions cleanly via `tea.ExecProcess`: TUI suspends, terminal hands off, TUI resumes on exit
- Prompts for SSH username (pre-filled with your local user, remembers per-host across the session)
- Copies the selected peer's Tailscale IP to clipboard
- Filters the peer list. The filtered peers are not gone. They are merely not being looked at
- Refreshes peer data every 5 seconds

---

## Installation

### From source

```bash
git clone https://github.com/mogglemoss/lazytailscale
cd lazytailscale
go build -o lazytailscale .
./lazytailscale
```

### go install

```bash
go install github.com/mogglemoss/lazytailscale@latest
```

Requires `tailscaled` running locally. On Linux, the process must have access to `/var/run/tailscale/tailscaled.sock` — run as the user who owns the Tailscale session, or with appropriate permissions.

---

## Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous node |
| `↓` / `j` | Next node |
| `enter` | SSH into selected node |
| `p` | Ping selected node now |
| `r` | Toggle subnet routes |
| `c` | Copy Tailscale IP to clipboard |
| `/` | Filter peer list |
| `R` | Refresh peer list |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

Mouse supported: click to select, scroll to navigate.

---

## Technical Specifications

| Parameter | Value |
|-----------|-------|
| Data source | LocalClient · no API key · no external network requests |
| Poll interval | 5s peers · 10s ping |
| Ping type | TSMP |
| Ping history | 8 samples per node |
| SSH | `tea.ExecProcess` · clean terminal handoff |
| Clipboard | `pbcopy` / `xclip` / `wl-copy` · detected at runtime |
| Requires | `tailscaled` running locally |
| Runtime dependencies | None |

---

## Sparkline Color Semantics

| Color | Meaning |
|-------|---------|
| Green | avg < 10ms |
| Amber | avg < 50ms |
| Red | avg ≥ 50ms |
| `✕` | Ping failed |

---

## Connection Status

- `◈ direct` — WireGuard peer-to-peer
- `◌ relayed` — Traffic transiting a DERP relay
- `○ unknown` — Connection type undetermined

---

## Not Affiliated

lazytailscale is not affiliated with or endorsed by Tailscale Inc. It reads from the local socket and means no harm.

---

## License

MIT. See [LICENSE](./LICENSE).
