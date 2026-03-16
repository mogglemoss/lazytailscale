# lazytailscale

![License: MIT](https://img.shields.io/badge/license-MIT-pink.svg)
![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)
[![Built with Charm](https://img.shields.io/badge/built_with-Charm-ff69b4.svg)](https://charm.sh)

A terminal dashboard for your Tailscale network. Two-pane keyboard-driven TUI: peer list on the left, selected-peer detail on the right. Runs entirely from your local Tailscale socket — no API key, no cloud, no opinions about your network topology.

---

![lazytailscale](./assets/lazytailscale.png)

---

## Features

**Peer list**
- All nodes on your tailnet, sorted online-first then alphabetically
- Status dots: green (online), amber (seen < 5min), red (offline)
- Exit node and subnet router indicators
- Live node count and filter-aware paginator

**Per-peer detail**
- Tailscale IP and MagicDNS name
- Connection type: `◈ direct` (peer-to-peer) or `◌ relayed` (via DERP relay)
- OS, last contact, last WireGuard handshake
- Exit node status with one-key toggle
- Advertised subnet routes
- ACL tags
- Key expiry warning when ≤ 14 days remaining

**Latency**
- Pings the selected peer every 10 seconds via TSMP
- Sparkline of last 8 results with avg / min / max
- Color-coded: green < 10ms · amber < 50ms · red ≥ 50ms · `✕` for failed

**SSH**
- `enter` suspends the TUI, hands off the terminal, resumes on exit
- Username prompt pre-filled with your local username, remembers per-host across the session
- MagicDNS name used when available, IP as fallback

**Notifications**
- Status bar briefly notes when a peer connects or disconnects between polls

**Theming**
- Built-in Charm Native palette (hot pink · mint · soft purple)
- Automatically adopts your [Omarchy](https://omarchy.org) theme when detected — reads `~/.config/omarchy/themes/current/colors.toml`, no configuration required
- `AdaptiveColor` throughout for correct rendering in both light and dark terminals

**Demo mode**
- `--demo` runs with a fictional tailnet — no Tailscale installation required
- Useful for screenshots, testing, and kicking the tyres before committing

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install mogglemoss/tap/lazytailscale
```

### go install

```bash
go install github.com/mogglemoss/lazytailscale@latest
```

### From source

```bash
git clone https://github.com/mogglemoss/lazytailscale
cd lazytailscale
go build -o lazytailscale .
./lazytailscale
```

Requires `tailscaled` running locally. On Linux the process must have access to `/var/run/tailscale/tailscaled.sock` — run as the user who owns the Tailscale session, or with appropriate permissions.

---

## Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous node |
| `↓` / `j` | Next node |
| `enter` | SSH into selected node |
| `e` | Toggle exit node on / off |
| `p` | Ping selected node now |
| `r` | Toggle subnet routes |
| `c` | Copy address to clipboard (MagicDNS preferred) |
| `/` | Filter peer list |
| `R` | Refresh peer list |
| `?` | Toggle full help |
| `q` / `ctrl+c` | Quit |

Mouse: click to select · scroll wheel to navigate list or scroll detail pane.

---

## Technical Specifications

| Parameter | Value |
|-----------|-------|
| Data source | `tailscaled` LocalClient · no API key · no external requests |
| Poll interval | 5s peers · 10s ping |
| Ping type | TSMP |
| Ping history | 8 samples per node, ring buffer |
| SSH | `tea.ExecProcess` · clean terminal handoff · no pty management |
| Clipboard | `pbcopy` / `xclip` / `wl-copy` · detected at runtime |
| Theming | Omarchy auto-detected · built-in Charm Native fallback |
| Runtime dependencies | None |

---

## Not Affiliated

lazytailscale is not affiliated with or endorsed by Tailscale Inc. It reads from the local socket and means no harm.

---

## License

MIT. See [LICENSE](./LICENSE).
