# porthog

Cross-platform port management CLI — find what's hogging your ports.

## Features

- **List** listening ports with process info (PID, name, user)
- **Kill** processes by port with TOCTOU-safe termination
- **Find** free ports in any range
- **Watch** real-time port activity in a TUI dashboard
- Works on **Linux**, **macOS**, and **Windows**
- Single binary, zero runtime dependencies

## Install

**Go install:**
```bash
go install github.com/z1j1e/porthog/cmd/porthog@latest
```

**Shell script (Linux/macOS):**
```bash
curl -sSL https://raw.githubusercontent.com/z1j1e/porthog/main/install.sh | sh
```

**Scoop (Windows):**
```powershell
scoop install porthog
```

## Usage

```bash
porthog list                          # list all listening ports
porthog list 8080                     # show who's using port 8080
porthog list --json                   # JSON output for scripting
porthog list --tcp                    # TCP only
porthog list --sort pid               # sort by PID
porthog kill 8080                     # kill process on port 8080
porthog kill 8080 --dry-run           # preview without killing
porthog kill 8080 --force             # force kill (SIGKILL)
porthog free                          # find one free port
porthog free --range 8000-9000 --count 3  # find 3 free ports in range
porthog watch                         # real-time TUI monitor
porthog completion bash               # generate shell completions
```

## Comparison

| Feature | porthog | fkill-cli | killport |
|---------|---------|-----------|----------|
| Language | Go | Node.js | Node.js |
| Single binary | ✅ | ❌ | ❌ |
| Zero dependencies | ✅ | ❌ | ❌ |
| List ports | ✅ | ❌ | ❌ |
| Kill by port | ✅ | ✅ | ✅ |
| Find free port | ✅ | ❌ | ❌ |
| TUI watch mode | ✅ | ❌ | ❌ |
| JSON output | ✅ | ❌ | ❌ |
| TOCTOU protection | ✅ | ❌ | ❌ |
| Cross-platform | ✅ | ✅ | ✅ |

## Troubleshooting

### Permission errors

**Linux:** Run with `sudo` to see all ports and process info:
```bash
sudo porthog list
```

**macOS:** Some ports require root access. SIP may restrict process info.

**Windows:** Run as Administrator to see all process details.

### Port not found

If `porthog list <port>` returns nothing, the port may be bound to IPv6 only.
IPv6 support is planned for v0.2.0.

### Watch mode not starting

`porthog watch` requires a TTY. It won't work in piped or CI environments.
Use `porthog list --json` for scripted monitoring.

## Architecture

porthog uses a hexagonal (ports & adapters) architecture:

```
cmd/porthog/          CLI entry point (Cobra)
internal/core/
  domain/             Entities (PortBinding, ProcessIdentity)
  ports/              Interfaces (Enumerator, Terminator, etc.)
  services/           Business logic (ListPorts, KillByPort, etc.)
internal/adapters/
  os/{linux,darwin,windows}/  Platform-specific implementations
  process/            Process metadata enrichment (gopsutil)
  output/             Renderers (table, JSON, plain)
internal/tui/watch/   Bubbletea TUI for watch mode
```

## License

MIT
