# Q-SYS MCP Server (v0.9.0)

A Model Context Protocol (MCP) server that enables AI coding assistants (such as Cursor, Cline, Claude Desktop, Antigravity, etc.) to control Q-SYS Core systems and **highly assists in writing Q-SYS Lua scripts and CSS styling** by referencing official guidelines dynamically.

---

## Key Features

1.  **AI Scripting & Styling Assistance (New in v0.9.0)**:
    *   Exposes dynamic Q-SYS CSS and Lua reference sheets directly to AI models.
    *   AIs can inspect API specifications (`TcpSocket`, `HttpClient`, `Timer`, `NamedControl`, `Component`) and UCI CSS class definitions on-the-fly to write bug-free code.
2.  **Core Connectivity & Status**: Check core states and design versions.
3.  **Control Management**: Get and set Named Controls with values, strings, and positions.
4.  **Component Control**: Retrieve and modify properties inside Named Components (including decimal-escaped frequencies).
5.  **Change Groups & Snapshots**: Efficiently poll values via change groups and trigger snapshot bank saves/loads.
6.  **UCI Manipulation**: Remotely switch user control interface pages.

---

## Installation & Running

### Download Releases
Pre-compiled binaries and setup packages are available on the [GitHub Releases](https://github.com/Mono-Gen/QSYS-MCP/releases) page. Download the ZIP archive for the corresponding release version.

### 1. Prerequisites
Ensure you have the following files in the same folder:
*   `qsys-mcp` (or `qsys-mcp.exe` on Windows)
*   `config.json` (Configuration file)

### 2. Configure `config.json`
Edit your `config.json` file to define connected cores and reference file paths:

```json
{
  "cores": "default=127.0.0.1:1710:1702:admin:password",
  "styles_dir": "C:\\Users\\YOUR_USERNAME\\Documents\\QSC\\Q-SYS Designer\\Styles",
  "lua_reference_path": "reference/reference_lua.md",
  "css_reference_path": "reference/reference_css.md",
  "polling_interval": 350,
  "debug": true
}
```

*   **cores**: Connection string formatted as `alias=IP:QRC_Port:ECP_Port:username:pin`. Credentials and ports can be omitted if Access Control is disabled (e.g., `"cores": "default=192.168.1.100"`).
*   **lua_reference_path** / **css_reference_path**: Paths to the Markdown reference sheets generated from Q-SYS Help.

### 3. Execution

#### Windows
Run via Command Prompt or PowerShell:
```powershell
.\qsys-mcp.exe
```

#### macOS (Apple Silicon/Intel)
Rename the corresponding binary to `qsys-mcp`, grant execution rights, and run:
```bash
chmod +x ./qsys-mcp
./qsys-mcp
```

#### Running from Source (For Developers)
If you have the Go SDK installed and want to run the server directly from source code without compiling:
```powershell
go run .
```

---

## Integration with AI Clients

This server runs via standard I/O (`stdio`). You can register it in your client configuration.

### Claude Desktop Configuration

#### Option A: Using Pre-compiled Binary
Add the server configuration to your `%APPDATA%\Claude\claude_desktop_config.json` (Windows) or `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "qsys-mcp": {
      "command": "cmd.exe",
      "args": ["/c", "C:\\path\\to\\qsys-mcp.exe"],
      "cwd": "C:\\path\\to\\project_directory"
    }
  }
}
```

#### Option B: Running from Source (Recommended for Development)
If you want to run the server directly from the source code during development (allowing code changes to take effect immediately upon server restart):

```json
{
  "mcpServers": {
    "qsys-mcp-dev": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "C:\\path\\to\\project_directory"
    }
  }
}
```

---

## Available MCP Tools

*   **Connection**: `qsys_list_cores`, `qsys_core_status`
*   **Controls**: `qsys_get_control`, `qsys_set_control`, `qsys_get_controls`
*   **Components**: `qsys_list_components`, `qsys_get_component_controls`, `qsys_set_component_controls`
*   **Snapshots**: `qsys_list_snapshots`, `qsys_load_snapshot`, `qsys_save_snapshot`
*   **Change Groups**: `qsys_create_change_group`, `qsys_poll_change_group`, `qsys_destroy_change_group`
*   **UCI Page Control**: `qsys_set_uci_page`, `qsys_get_uci_status`
*   **Local CSS Styling**: `qsys_write_local_css`, `qsys_read_local_css`
*   **Developer References (New)**:
    *   `qsys_get_css_reference`: Retrieves Q-SYS UCI CSS selectors and class naming sheets.
    *   `qsys_get_lua_reference`: Retrieves Q-SYS Lua API specifications and GC prevention code templates.

---

## Documentation
Full detailed manuals are located in the `doc/` directory:
*   [English Manual](doc/manual_en.md)
*   [Japanese Manual](doc/manual_ja.md)
