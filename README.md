# Q-SYS Model Context Protocol (MCP) Server

`qsys-mcp` is an MCP server that allows AI agents (such as Claude Desktop or Gemini Antigravity) to monitor and control a Q-SYS Core processor over the network. It bridges Q-SYS Remote Control (QRC) and External Control Protocol (ECP) to Model Context Protocol, enabling intelligent AV automation, configuration, debugging, and UI styling.

## Features

- **Named Control Access**: Get and set value, position, or string representation of named controls.
- **Component Control**: Query and batch-set controls within named schematic components (mixers, EQs, etc.).
- **Snapshot Management**: Load and save snapshots by name or bank/number.
- **UCI Routing**: Switch active pages and query UCI status on touchscreen controllers.
- **Change Group Monitoring**: Efficiently poll controls for state changes.
- **Lua Script Execution**: Run Lua commands directly on the Core for debugging.
- **Local CSS Integration**: Edit and apply UCI CSS stylesheets locally and sync to Q-SYS Designer.

## Requirements

- Go 1.21+ (if compiling from source)
- A network-reachable Q-SYS Core processor (or Q-SYS Designer running in Emulation mode)

## Configuration (`config.json`)

Configure the server by editing `config.json` in the same directory as the executable:

```json
{
  "cores": "default=127.0.0.1:1710:1702:admin:1234",
  "styles_dir": "C:\\Users\\Username\\Documents\\QSC\\Q-SYS Designer\\Styles",
  "lua_reference_path": "reference/reference_lua.md",
  "css_reference_path": "reference/reference_css.md",
  "polling_interval": 350,
  "protected_controls": "system\\..*,.*\\.password",
  "debug": false
}
```

### Config Parameters

- `cores`: Target Q-SYS Cores. Format: `alias=IP:qrcPort:ecpPort:username:password`. Ports and auth are optional. Multiple cores can be comma-separated.
- `styles_dir`: Absolute path to Q-SYS Designer Styles directory for local CSS syncing.
- `lua_reference_path` / `css_reference_path`: Paths to markdown references for Lua API and CSS rules.
- `polling_interval`: Recommended change group polling interval in milliseconds.
- `protected_controls`: Comma-separated regex list of controls that should block write attempts.

## Download

Pre-built binaries are available on the [Releases page](https://github.com/Mono-Gen/QSYS-MCP/releases).

| Platform | File |
|---|---|
| Windows (x64) | `qsys-mcp.exe` (inside the zip) |
| macOS (Apple Silicon) | `qsys-mcp-mac-arm64` (inside the zip) |
| macOS (Intel) | `qsys-mcp-mac-amd64` (inside the zip) |

## Installation & Setup

### 1. Run/Compile the Server
To run from source:
```bash
go run .
```
To compile a binary:
```bash
go build -ldflags="-s -w -extldflags '-static'" -o qsys-mcp
```

### 2. Claude Desktop Integration

Edit the Claude configuration file for your OS:

- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

#### Windows

```json
{
  "mcpServers": {
    "qsys-mcp": {
      "command": "cmd.exe",
      "args": [
        "/c",
        "C:\\path\\to\\qsys-mcp.exe"
      ],
      "cwd": "C:\\path\\to\\config_directory"
    }
  }
}
```

#### macOS

```json
{
  "mcpServers": {
    "qsys-mcp": {
      "command": "/path/to/qsys-mcp",
      "args": [],
      "cwd": "/path/to/config_directory"
    }
  }
}
```

## Available Tools

The MCP server provides the following tools to the AI agent:
- `qsys_list_cores`: List configured cores and states.
- `qsys_core_status`: Get system engine status.
- `qsys_get_control` / `qsys_set_control`: Read and write named controls.
- `qsys_get_controls`: Batch read named controls.
- `qsys_list_components` / `qsys_get_component_controls` / `qsys_set_component_controls`: Component-level control.
- `qsys_list_snapshots` / `qsys_load_snapshot` / `qsys_save_snapshot`: Snapshot control.
- `qsys_create_change_group` / `qsys_poll_change_group` / `qsys_destroy_change_group`: Monitor control changes.
- `qsys_set_uci_page` / `qsys_get_uci_status`: Switch and monitor touchscreen UCI screens.
- `qsys_write_local_css` / `qsys_read_local_css`: Read/Write style sheets.
- `qsys_run_lua`: Run arbitrary Lua commands on the Core.
