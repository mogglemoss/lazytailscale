# lazytailscale

![License: MIT](https://img.shields.io/badge/license-MIT-pink.svg)
![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)
[![Built with Charm](https://img.shields.io/badge/built_with-Charm-ff69b4.svg)](https://charm.sh)

A terminal dashboard for your Tailscale network. Two-pane keyboard-driven TUI: peer list on the left, selected-peer detail on the right. Runs entirely from your local Tailscale socket — no API key, no cloud, no opinions about your network topology.

---

![lazytailscale](./assets/lazytailscale.gif)

---

## Features

**Peer list**
- All nodes on your tailnet, sorted online-first then alphabetically
- Status dots: green (online) · amber (seen < 5 min) · red (offline)
- Exit node and subnet router indicators
- Live node count with filter-aware paginator

**Per-peer detail**
- Tailscale IP and MagicDNS name
- Connection type: `◈ direct` (peer-to-peer) or `◌ relayed` (via DERP relay)
- OS, last contact, last WireGuard handshake
- Exit node status with one-key toggle (`e`)
- Advertised subnet routes
- ACL tags
- Key expiry warning when ≤ 14 days remaining

**Latency**
- Pings the selected peer every 10 seconds via TSMP
- Sparkline of last 8 results with avg / min / max and trend arrow (↓ improving · ↑ degrading)
- Color-coded: green < 10ms · amber < 50ms · red ≥ 50ms · `✕` for failed

**Connect**
- `enter` on a peer opens a centered connection picker modal
- **SSH** — `enter` fast-paths directly if credentials are saved for that host; `s` opens the form to review or edit username and port; suspends the TUI and hands off the terminal cleanly, resuming with a welcome-back message on exit; connection failures surface a dismissable error panel rather than a silent flash
- **RDP** — opens the platform RDP client (`open rdp://` on macOS, `xfreerdp`/`remmina` on Linux, `mstsc` on Windows)
- **VNC** — opens the platform VNC viewer (`open vnc://` on macOS, `vncviewer`/`xdg-open` on Linux); credentials handled by the viewer app

**SSH server mode**
- `--serve` runs lazytailscale as a [Wish](https://github.com/charmbracelet/wish) SSH server — no client install needed
- `ssh yourhost -p 23234` gives any user on your tailnet the full interactive dashboard
- Each connection gets its own isolated session; the dashboard talks to the local `tailscaled` socket
- `--port` and `--host` flags control the bind address (defaults: `0.0.0.0:23234`)

**Connection control**
- Connect and disconnect Tailscale from within the TUI (`u`)
- Status bar reflects current node state: NODE NOMINAL · NODE UNREACHABLE · DISCONNECTED

**Notifications**
- Status bar briefly notes when a peer connects or disconnects between polls
- Exit node flash: `⬡` pulses in the status bar when an exit node is activated or deactivated
- Refresh heartbeat: `◦` appears in the status bar after each successful peer fetch

**Filter**
- `/` activates the peer list filter; matches against hostname, Tailscale IP, OS, and MagicDNS name
- Scroll indicators (▲ / ▼) appear when the list is taller than the viewport

**Theming**
- Built-in Charm Native palette (hot pink · mint · soft purple)
- Automatically adopts your [Omarchy](https://omarchy.org) theme when detected — reads `~/.config/omarchy/themes/current/colors.toml`, no configuration required
- `AdaptiveColor` throughout for correct rendering in both light and dark terminals

**Demo mode**
- `--demo` runs with a fictional tailnet — no Tailscale installation required
- Useful for screenshots, testing, or trying it out before committing

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install mogglemoss/tap/lazytailscale
```

### AUR (Arch Linux / Omarchy)

```bash
yay -S lazytailscale-bin
# or
paru -S lazytailscale-bin
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
| `enter` | Open connection picker for selected node |
| `e` | Toggle exit node on / off |
| `u` | Connect / disconnect Tailscale |
| `p` | Ping selected node now |
| `r` | Toggle subnet routes expanded view |
| `c` | Copy address to clipboard (MagicDNS preferred) |
| `/` | Filter peer list (hostname · IP · OS · DNS) |
| `R` | Refresh peer list |
| `?` | Toggle full help |
| `q` / `ctrl+c` | Quit |

**Inside the connection picker:**

| Key | Action |
|-----|--------|
| `enter` / `s` | SSH (fast-path if credentials saved; `s` always shows form) |
| `r` | RDP |
| `v` | VNC |
| `esc` | Cancel |

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
| RDP | `open rdp://` / `xfreerdp` / `remmina` / `mstsc` · detected at runtime |
| VNC | `open vnc://` / `vncviewer` / `xdg-open` · detected at runtime |
| SSH server | [Wish](https://github.com/charmbracelet/wish) · `--serve --port 23234` |
| Clipboard | `pbcopy` / `xclip` / `wl-copy` · detected at runtime |
| Theming | Omarchy auto-detected · built-in Charm Native fallback |
| Runtime dependencies | None |

---

## Not Affiliated

lazytailscale is not affiliated with or endorsed by Tailscale Inc. It reads from the local socket and means no harm.

---

## License

MIT. See [LICENSE](./LICENSE).
