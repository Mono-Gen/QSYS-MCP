# Q-SYS MCP Server Setup & Reference Manual (v0.9.0)

This document is a comprehensive manual for installing, running, and using all features (MCP Tools) provided by the Q-SYS MCP Server (`qsys-mcp`) v0.9.0.

---

## Installation & Setup

### 1. File Structure
The following two files are required and should be placed in the same folder:
*   `qsys-mcp` (or `qsys-mcp.exe` for Windows)
*   `config.json` (Configuration file)

### 2. Editing `config.json`
Configure the server settings according to your Q-SYS Core network environment and access control requirements.

*   **When the connected Core has password protection (Access Control enabled)**:
    Append the username and password (PIN) to the `cores` entry.
    *   **Format**: `"cores": "alias=IP:QRC_Port:ECP_Port:username:password(PIN)"`
    *   **Example**:
        ```json
        {
          "cores": "default=127.0.0.1:1710:1702:admin:1234",
          "styles_dir": "C:\\Users\\YOUR_USERNAME\\Documents\\QSC\\Q-SYS Designer\\Styles",
          "lua_reference_path": "reference/reference_lua.md",
          "css_reference_path": "reference/reference_css.md",
          "polling_interval": 350,
          "debug": true
        }
        ```
*   **Without Password Protection**:
    Omit ports and credentials.
    *   **Example**: `"cores": "default=127.0.0.1"`

### Configuration Parameters Detail (`config.json` options)

The meaning and usage of each configuration option in `config.json` is as follows:

*   **`cores`** (string, required)
    *   **Description**: Network connection information for target Q-SYS Cores.
    *   **Format**: `"alias=IP:QRC_Port:ECP_Port:username:password(PIN)"`
    *   **Field Details**:
        *   `alias`: An identifier for the Core (e.g. `default`, `sub-core`). When a tool call omits the core alias, the `default` alias is used.
        *   `IP`: The IP address or hostname of the Q-SYS Core.
        *   `QRC_Port`: Q-SYS Remote Control protocol port (defaults to `1710`).
        *   `ECP_Port`: External Control Protocol port (defaults to `1702`).
        *   `username` / `password(PIN)`: Administrator credentials if Access Control is enabled on the Core. If no access control is configured, ports and credentials can be omitted entirely (e.g., `"cores": "default=127.0.0.1"`).
*   **`styles_dir`** (string, required)
    *   **Description**: Local directory path to read/write CSS style sheets for Q-SYS Designer.
    *   **Details**: On Windows, this is typically `C:\Users\YOUR_USERNAME\Documents\QSC\Q-SYS Designer\Styles`. The AI can view/modify CSS styles in this directory using the `qsys_write_local_css` and `qsys_read_local_css` tools.
*   **`lua_reference_path`** (string, required)
    *   **Description**: Local file path to the Q-SYS Lua API specification sheet (Markdown).
    *   **Details**: Defaults to `reference/reference_lua.md`. This is read and returned when the AI invokes `qsys_get_lua_reference`.
*   **`css_reference_path`** (string, required)
    *   **Description**: Local file path to the Q-SYS UCI CSS styling reference sheet (Markdown).
    *   **Details**: Defaults to `reference/reference_css.md`. This is read and returned when the AI invokes `qsys_get_css_reference`.
*   **`polling_interval`** (number, optional)
    *   **Description**: Polling rate in milliseconds for Q-SYS Core connections. Defaults to `350`.
    *   **Details**: Determines the interval for background status checks and change group tracking.
*   **`debug`** (boolean, optional)
    *   **Description**: Enable or disable debug logging. Defaults to `false`.
    *   **Details**: If `true`, all incoming and outgoing JSON-RPC communications are printed to the standard error console, helping with troubleshooting.
*   **`protected_controls`** (string, optional)
    *   **Description**: Regular expression pattern to protect specific controls from unauthorized modification.
    *   **Details**: For example, `"system\\..*,.*\\.password"` blocks any external write attempt to controls matching these naming patterns.

### 3. Running on Different OS

#### ■ Windows
1. Place `qsys-mcp.exe` and `config.json` in the same directory.
2. Run via Command Prompt or PowerShell:
   ```powershell
   .\qsys-mcp.exe
   ```

#### ■ macOS (Apple Silicon/Intel)
1. Prepare the correct binary for your CPU structure (`arm64` for Apple Silicon, `amd64` for Intel) and rename it to `qsys-mcp`.
2. Place it in the same folder as `config.json`.
3. **Grant Execution Rights (First time only)**:
   ```bash
   chmod +x ./qsys-mcp
   ```
4. **Bypass Gatekeeper Warning (First time only)**:
   Right-click (or Ctrl + Click) on the binary in Finder and select **Open**. Click "Open" in the popup window.
5. Run the binary from terminal.

### 4. Integrating with AI Clients

This MCP server adheres to standard `stdio` communication and can be integrated with various AI assistants and editor plugins.

#### Configuration File Location
Edit the configuration file corresponding to your tool and OS:
*   **Claude Desktop**:
    *   **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
    *   **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
*   **Claude Code (CUI)**:
    *   **Common**: `~/.claude.json`
*   **Antigravity 2.0 / MCP Agents**:
    *   **Common**: `~/.gemini/antigravity/mcp_settings.json` (or settings panel)

#### Configuration Examples (Windows via cmd.exe)
On Windows, using `cmd.exe /c` wraps the execution safely:
```json
{
  "mcpServers": {
    "qsys-mcp": {
      "command": "cmd.exe",
      "args": [
        "/c",
        "<absolute_path_to_qsys-mcp.exe_directory>\\qsys-mcp.exe"
      ],
      "cwd": "<absolute_path_to_config.json_directory>"
    }
  }
}
```
*   **Note**: Make sure to replace `<absolute_path_to_qsys-mcp.exe_directory>` and `<absolute_path_to_config.json_directory>` with the actual folder paths on your machine. The `cwd` must be the directory containing `config.json`.
*   **Note**: In Windows JSON configurations, all backslashes (`\`) in paths must be doubled (i.e. `\\`).

---

## 1. Connection Management (Connection)

### `qsys_list_cores`
Retrieves a list of all configured Q-SYS Core aliases and IP addresses from `config.json`.

*   **Parameters**: None
*   **Return Value**: List of configured Cores.

### `qsys_core_status`
Gets the current status (Active/Standby/Emulator) and the running design name of a specified Core.

*   **Parameters**:
    *   `core` (string, optional): Core alias. Omit if only one Core is configured.
*   **Return Value**: Core running status, platform, and design information.

---

## 2. Control Management (Controls)

Reads or sets the state of Q-SYS **Named Controls**.

### `qsys_get_control`
Gets the current state (Value, String, Position) of a Named Control.

*   **Parameters**:
    *   `control` (string, required): Named Control name.
    *   `core` (string, optional): Core alias.
*   **Return Value**: Object containing `Value`, `String`, and `Position` (0.0 to 1.0).

### `qsys_set_control`
Sets the state of a Named Control. Supports numerical Values, display Strings, and float Positions.

*   **Parameters**:
    *   `control` (string, required): Named Control name.
    *   `value` (number, optional): Value to set.
    *   `string` (string, optional): Display text to set.
    *   `position` (number, optional): Position float (0.0 to 1.0) to set.
    *   `core` (string, optional): Core alias.
*   **Return Value**: Success/failure response message.

### `qsys_get_controls`
Retrieves the states of multiple Named Controls in a single request.

*   **Parameters**:
    *   `controls` (array of string, required): List of Named Control names.
    *   `core` (string, optional): Core alias.
*   **Return Value**: State list of requested controls.

---

## 3. Component Control (Components)

Interacts with controls within named schematic block **Components**.

### `qsys_list_components`
Lists all Named Components (Code Name and Type) configured in the active design.

*   **Parameters**:
    *   `core` (string, optional): Core alias.
*   **Return Value**: List of schematic component info.

### `qsys_get_component_controls`
Gets the state of all controls inside a specified Named Component.

*   **Parameters**:
    *   `component` (string, required): Component name.
    *   `core` (string, optional): Core alias.
*   **Return Value**: List of control states inside the component.

### `qsys_set_component_controls`
Sets the states of multiple controls in a specified component at once.

*   **Parameters**:
    *   `component` (string, required): Component name.
    *   `controls` (array of object, required): List of control definitions to set:
        *   `Name` (string, required)
        *   `Value` (number, optional)
        *   `String` (string, optional)
        *   `Position` (number, optional)
    *   `core` (string, optional): Core alias.
*   **Return Value**: Result message.

---

## 4. Snapshot Control (Snapshots)

Controls Q-SYS **Snapshot Banks**.

### `qsys_list_snapshots`
Lists all Snapshot Banks and saved snapshot numbers configured in the design.

*   **Parameters**:
    *   `core` (string, optional): Core alias.
*   **Return Value**: List of bank names and snapshots.

### `qsys_load_snapshot`
Loads a specific snapshot index from a bank, with an optional fade/ramp time.

*   **Parameters**:
    *   `bank` (string, required): Snapshot Bank name.
    *   `number` (number, required): Snapshot index (1+).
    *   `ramp_time` (number, optional): Transition time in seconds.
    *   `core` (string, optional): Core alias.
*   **Return Value**: Action result.

### `qsys_save_snapshot`
Saves the current control states to a specific snapshot index.

*   **Parameters**:
    *   `bank` (string, required): Snapshot Bank name.
    *   `number` (number, required): Snapshot index (1+).
    *   `core` (string, optional): Core alias.
*   **Return Value**: Action result.

---

## 5. Change Groups (Change Groups)

Efficiently groups controls to monitor values only when they change.

### `qsys_create_change_group`
Creates a change group and registers controls to monitor.

*   **Parameters**:
    *   `id` (number, required): Change group ID (1 to 4).
    *   `controls` (array of string, required): List of Named Controls to monitor.
    *   `core` (string, optional): Core alias.
*   **Return Value**: Success response.

### `qsys_poll_change_group`
Polls the change group and returns only the controls that changed since the last poll.

*   **Parameters**:
    *   `id` (number, required): Change group ID (1 to 4).
    *   `core` (string, optional): Core alias.
*   **Return Value**: List of controls that changed and their new states.

### `qsys_destroy_change_group`
Destroys a change group.

*   **Parameters**:
    *   `id` (number, required): Change group ID (1 to 4).
    *   `core` (string, optional): Core alias.
*   **Return Value**: Success response.

---

## 6. UCI Control (UCI)

Manipulates Q-SYS **User Control Interfaces (UCI)**.

### `qsys_set_uci_page`
Triggers a page transition on a specified UCI.

*   **Parameters**:
    *   `uci_name` (string, required): UCI name.
    *   `page_name` (string, required): Target page name.
    *   `core` (string, optional): Core alias.
*   **Return Value**: Result message.

### `qsys_get_uci_status`
Retrieves connection status and active page name of a specified UCI.

*   **Parameters**:
    *   `uci_name` (string, required): UCI name.
    *   `core` (string, optional): Core alias.
*   **Return Value**: UCI status information.

---

## 7. Local CSS Actions (Local CSS)

Utilities to read and write CSS files inside the local styles directory.

### `qsys_write_local_css`
Creates or overwrites a CSS style sheet in the configured styles directory.

*   **Parameters**:
    *   `filename` (string, required): CSS filename (e.g. 'custom.css', path dividers are prohibited).
    *   `content` (string, required): CSS content string.
*   **Return Value**: Success message.

### `qsys_read_local_css`
Reads the content of a CSS file from the styles directory.

*   **Parameters**:
    *   `filename` (string, required): CSS filename.
*   **Return Value**: CSS file text content.

---

## 8. Local Reference Retrieval (Local Reference)

Retrieves Q-SYS Lua and CSS guides configured in `config.json` to assist AIs during development.

### `qsys_get_css_reference`
Retrieves selectors, pseudo-states, and icon property sheets for Q-SYS UCI CSS.

*   **Parameters**: None
*   **Return Value**: CSS guide text (Markdown format).

### `qsys_get_lua_reference`
Retrieves API specifications (`TcpSocket`, `HttpClient`, `Timer`, etc.) and garbage collection templates for Q-SYS Lua.

*   **Parameters**: None
*   **Return Value**: Lua guide text (Markdown format).

---

## AI-Driven Development Workflows

By integrating this MCP server with AI editors/agents (like Antigravity, Cursor, Claude Code, Cline), you can significantly accelerate Q-SYS Lua scripting and UCI CSS styling. Here are the recommended step-by-step procedures for beginners.

### 1. Workflow for Writing Lua Scripts

When asking the AI assistant to write scripts for Q-SYS scriptable components (e.g. Text Controller, Block Controller):

1.  **Prompt the AI to load specifications**
    At the start of your chat, instruct the AI to load the Q-SYS API guidelines.
    *   **Example Prompt**:
        > "I want to write a Q-SYS Lua script. Please first call the `qsys_get_lua_reference` tool to check Q-SYS specific APIs and garbage collection warning templates."
2.  **AI Code Generation**
    Based on the reference guidelines, the AI will generate clean code adhering to critical Q-SYS principles:
    *   It will avoid assigning asynchronous objects like `TcpSocket.New()` or `Timer.New()` to local variables (`local conn = TcpSocket.New()`) inside local functions, preventing them from being garbage-collected prematurely.
    *   It will properly register non-blocking callback functions to sockets and timers (e.g. `conn.Connected` and `conn.Data`).
3.  **Deploy and Verify**
    Write the code to a local script file and use utility scripts (like `samples/send_lua.py`) to send the Lua code directly into the active Q-SYS design component for testing.

---

### 2. Workflow for Writing UCI CSS Style Sheets

When decorating Q-SYS User Control Interfaces (UCIs) using custom CSS classes and stylesheets:

1.  **Prompt the AI to load styling rules**
    Tell the AI to analyze Q-SYS CSS naming conventions and selector limitations first.
    *   **Example Prompt**:
        > "I want to design a custom theme for a Q-SYS UCI. Please first call the `qsys_get_css_reference` tool to study supported pseudo-classes, font rules, and layout constraints."
2.  **Prompt the layout requirements**
    Specify your design requirements to the AI.
    *   **Example Prompt**:
        > "Following the reference guidelines, create a CSS stylesheet that styles a mute button class `.button.mute`. It should flash red `#FF0000` when active, and show as a semi-transparent grey when inactive."
3.  **Write CSS directly to the style directory**
    Let the AI call `qsys_write_local_css` to save the stylesheet directly into the Q-SYS Designer styles folder configured in `config.json`.
    *   Once written, you can immediately select and apply the style to components inside Q-SYS Designer and see updates in real-time.
