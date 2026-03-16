# lazytailscale

![License: MIT](https://img.shields.io/badge/license-MIT-pink.svg)
![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)
![Built with Charm](https://raw.githubusercontent.com/charmbracelet/charm/main/title-dark.svg)

A terminal dashboard for your Tailscale network. LAZYTAILSCALE has no opinion on your network topology. LAZYTAILSCALE has opinions it has chosen not to share.

---

```
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║  LAZYTAILSCALE MAKES NO REPRESENTATIONS REGARDING THE        ║
║  ACCURACY OF PING LATENCY, THE MORAL CHARACTER OF YOUR       ║
║  PEERS, OR WHETHER THE NODE MARKED ONLINE IS ACTUALLY        ║
║  DOING ANYTHING USEFUL.                                      ║
║                                                              ║
║  LAZYTAILSCALE HAS NEVER SENT A PACKET.                      ║
║  LAZYTAILSCALE HAS NEVER RECEIVED ONE.                       ║
║  LAZYTAILSCALE CONSIDERS THESE FACTS ORTHOGONAL TO ITS       ║
║  MISSION.                                                    ║
║                                                              ║
║  YOUR PEERS ARE YOUR RESPONSIBILITY. LAZYTAILSCALE OBSERVES. ║
║  LAZYTAILSCALE DOCUMENTS. LAZYTAILSCALE DOES NOT INTERVENE.  ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
```

---

![lazytailscale](./screenshot.png)

## What It Does

LAZYTAILSCALE is a keyboard-driven TUI for inspecting your Tailscale network. It does the following things, which it considers sufficient:

- Displays all peers on your tailnet in a scrollable list, sorted by status and then alphabetically, because chaos is not a network topology
- Shows per-peer detail: Tailscale IP, MagicDNS name, OS, connection type (direct or relayed), last WireGuard handshake, advertised routes, ACL tags
- Pings the selected peer every 10 seconds and renders the history as a sparkline using braille block characters, because numbers alone are insufficient to convey the full emotional weight of a 3ms round trip
- Launches SSH sessions via `tea.ExecProcess`, which suspends the TUI cleanly, hands off the terminal, and resumes when you are done
- Copies the selected peer's Tailscale IP to the clipboard. LAZYTAILSCALE trusts you know what to do with it
- Filters the peer list. The filtered peers are not gone. They are merely not being looked at
- Refreshes the peer list every 5 seconds. What has changed is noted. LAZYTAILSCALE updates its records

---

## G.U.S.T.

When you select your own node in the peer list, you will encounter **G.U.S.T.** — the General Uptime Surveillance Terminal.

G.U.S.T. is a small animated ASCII entity that lives in the detail panel. It blinks. Its tail moves. It has been monitoring your tailnet and finds it adequate. It has filed the appropriate documentation. It does not sleep. It does not require sleep. It is fine.

G.U.S.T. rotates through eight remarks at a rate it has determined to be optimal. The remarks are accurate to the best of G.U.S.T.'s knowledge. G.U.S.T. acknowledges that its knowledge is limited to what can be inferred from a list of IP addresses.

G.U.S.T. is at peace with this.

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

Requires `tailscaled` running locally. On Linux, the process must have access to `/var/run/tailscale/tailscaled.sock`. Run as the user who owns the Tailscale session, or with appropriate permissions. LAZYTAILSCALE does not adjudicate permission disputes.

---

## Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous node |
| `↓` / `j` | Next node |
| `enter` | Initiate SSH contact |
| `p` | Ping selected node |
| `r` | Toggle claimed prefixes |
| `c` | Copy Tailscale address to clipboard |
| `/` | Filter node registry |
| `R` | Refresh node registry |
| `?` | Toggle help overlay |
| `q` / `ctrl+c` | Terminate process |

---

## Technical Specifications

| Parameter | Value |
|-----------|-------|
| Data source | LocalClient · no API key · no network request beyond your tailnet |
| Poll interval | 5s peers · 10s ping |
| Ping type | TSMP |
| Ping history | 8 samples per node |
| SSH | `tea.ExecProcess` · terminal handoff · no pty management required |
| Clipboard | `pbcopy` / `xclip` / `wl-copy` · platform detected at runtime |
| Requires | `tailscaled` running locally |
| Dependencies | None at runtime. Several at compile time. Go handles it. |

---

## Sparkline Color Semantics

| Color | Meaning |
|-------|---------|
| Green | avg < 10ms. Satisfactory. |
| Amber | avg < 50ms. Noted. |
| Red | avg ≥ 50ms. G.U.S.T. has recorded this. |
| `✕` | Ping failed. The node did not respond. LAZYTAILSCALE is not surprised. |

---

## Connection Status

The detail panel reports connection type for online peers:

- `◈ direct` — WireGuard peer-to-peer. LAZYTAILSCALE approves, though it did not ask.
- `◌ relayed` — Traffic is transiting a DERP relay. This is fine. This is normal. Everything is fine.
- `○ unknown` — Connection type could not be determined. LAZYTAILSCALE has noted this in its records and moved on.

---

## Not Affiliated

LAZYTAILSCALE is not affiliated with, endorsed by, or in communication with Tailscale Inc. in any capacity. LAZYTAILSCALE simply reads from the local socket. LAZYTAILSCALE means no harm.

---

## License

MIT. See [LICENSE](./LICENSE).

LAZYTAILSCALE is provided as-is. LAZYTAILSCALE makes no warranty, express or implied, regarding uptime, packet delivery, or the continued goodwill of your peers.

---

*— G.U.S.T., probably watching*
