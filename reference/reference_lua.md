# Q-SYS Lua API Scripting Reference

This document is a cheat sheet for AI assistants and developers to refer to when writing Lua scripts for the Q-SYS Control Scripting environment (Text Controller, Scriptable Controls, etc.).

---

## 1. Crucial Rule: Garbage Collection Prevention

> [!WARNING]
> **"Never scope asynchronous objects like `TcpSocket` or `Timer` locally."**
> Declaring these as local variables will cause Lua's garbage collector to destroy them automatically during runtime. This leads to silent bugs where timers stop ticking or sockets disconnect unexpectedly. Always declare them globally.

*   **Incorrect (NG)**: `local myTimer = Timer.New()`
*   **Correct (OK)**: `myTimer = Timer.New()`

---

## 2. TcpSocket API

Used to establish TCP/IP client connections to remote devices on the network.

### Methods
*   **`TcpSocket.New()`**: Creates a new TcpSocket instance.
*   **`TcpSocket.NewTls()`**: Creates a new secure TCP socket instance using TLS.
*   **`:Connect(ipAddress, port)`**: Initiates connection to the remote host.
*   **`:Disconnect()`**: Disconnects the socket.
*   **`:Write(data)`**: Writes a string of data to the socket (always do this after the `Connected` event).
*   **`:Read(length)`**: Reads up to `length` bytes of data and removes them from the buffer.
*   **`:ReadLine(EOL, <custom>)`**: Reads a line up to the specified EOL sequence.
*   **`:Search(str, [start_pos])`**: Searches the buffer and returns the index of the string, or nil.

### Constants
*   **`TcpSocket.Events`**: `Connected`, `Reconnect`, `Data`, `Closed`, `Error`, `Timeout`
*   **`TcpSocket.EOL`**: `Any` (CR or LF), `CrLf` (recommended), `CrLfStrict` (`\r\n`), `Lf` (`\n`), `Null` (`\0`), `Custom`

### Properties
*   **`.EventHandler = function(socket, event, err) ... end`**: Event handler function to catch all socket events.
*   **`.ReadTimeout` / `.WriteTimeout`**: Timeout in seconds (0 to disable).
*   **`.ReconnectTimeout`**: Time in seconds to wait before auto-reconnect (default 5s, 0 to disable).
*   **`.IsConnected`**: Returns `true` if connected (read-only).
*   **`.BufferLength`**: Bytes currently in the buffer (read-only).

### Sample Code
```lua
target_ip = "192.168.1.100"
target_port = 1234
my_socket = TcpSocket.New() -- Global declaration (never local)

my_socket.ReadTimeout = 0
my_socket.WriteTimeout = 0
my_socket.ReconnectTimeout = 5 -- Attempt reconnect after 5 seconds

my_socket.EventHandler = function(sock, evt, err)
  if evt == TcpSocket.Events.Connected then
    print("Socket connected!")
    sock:Write("GET_STATUS\r\n")
  elseif evt == TcpSocket.Events.Reconnect then
    print("Socket reconnecting...")
  elseif evt == TcpSocket.Events.Data then
    -- Process buffer line by line
    local message = sock:ReadLine(TcpSocket.EOL.CrLf)
    while message ~= nil do
      print("Received message: " .. message)
      message = sock:ReadLine(TcpSocket.EOL.CrLf)
    end
  elseif evt == TcpSocket.Events.Closed then
    print("Socket closed by remote.")
  elseif evt == TcpSocket.Events.Error then
    print("Socket error: " .. tostring(err))
  end
end

my_socket:Connect(target_ip, target_port)
```

---

## 3. HttpClient API

Handles asynchronous HTTP/HTTPS client requests (GET, POST, PUT, etc.).

### Static Methods
*   **`HttpClient.Get( { Url=..., Timeout=..., EventHandler=... } )`**: Sends a GET request.
*   **`HttpClient.Post( { Url=..., Data=..., Headers=..., EventHandler=... } )`**: Sends a POST request.
*   **`HttpClient.Put( { Url=..., Data=..., Headers=..., EventHandler=... } )`**: Sends a PUT request.
*   **`HttpClient.Upload( { Url=..., Method="POST", Data=..., Headers=... } )`**: Uploads data (methods: POST, PUT, PATCH).
*   **`HttpClient.Download( { Url=..., Headers=... } )`**: Downloads data from a URL.

### Utility Methods
*   **`HttpClient.CreateUrl(table)`**: Builds a complete URL with encoded query parameters.
*   **`HttpClient.EncodeParams(table)`**: Encodes a table of parameters for a query string.
*   **`HttpClient.EncodeString(str)`**: URL-encodes a string.
*   **`HttpClient.DecodeString(str)`**: Decodes a URL-encoded string.

### Request Arguments (Table)
*   **`Url`**: Target URL. Prefix with `https://` for TLS.
*   **`Headers`**: Key-value table of headers (e.g., `{ ["Content-Type"] = "application/json" }`).
*   **`User` / `Password` / `Auth`**: Credentials and method (`"any"`, `"basic"`, `"digest"`).
*   **`Timeout`**: Timeout in seconds.
*   **`EventHandler`**: Callback function signature: `function(tbl, code, data, error, headers)`. `code` is the HTTP status (e.g., 200).

### Sample Code
```lua
-- Callback function for HTTP responses
function http_callback(tbl, code, data, err, headers)
  print(string.format("HTTP Response Code: %d", code))
  if code == 200 then
    print("Response Data: " .. data)
  else
    print("HTTP Error: " .. tostring(err))
  end
end

-- Create an encoded URL with query params
local query_url = HttpClient.CreateUrl({
  Host = "http://api.example.com",
  Path = "v1/status",
  Query = { device = "projector_1", command = "power_on" }
})

HttpClient.Get({
  Url = query_url,
  Timeout = 10,
  Headers = { ["Accept"] = "application/json" },
  EventHandler = http_callback
})
```

---

## 4. Timer API

Creates delays or repeating timed events (never use Lua's native sleep functions, as they freeze the entire Q-SYS engine).

### Methods
*   **`Timer.New()`**: Creates a new timer object (declare globally).
*   **`:Start(period_in_seconds)`**: Starts the timer with a repeat interval in seconds.
*   **`:Stop()`**: Stops the timer.
*   **`:IsRunning()`**: Returns true if the timer is running.
*   **`Timer.CallAfter(function, delay_in_seconds)`**: Triggers a function after a delay (single shot, no stop required).
*   **`Timer.Now()`**: Returns seconds since epoch (useful for calculating delta times).

### Properties
*   **`.EventHandler = function(timer) ... end`**: Function triggered when the timer elapses. Receives the timer object as an argument.

### Sample Code
```lua
pulse_timer = Timer.New() -- Global declaration
pulse_count = 0

pulse_timer.EventHandler = function(tmr)
  pulse_count = pulse_count + 1
  print("Pulse count: " .. pulse_count)
  
  if pulse_count >= 10 then
    print("Stopping repeating timer.")
    tmr:Stop()
    
    -- Delay trigger (single shot)
    Timer.CallAfter(function()
      print("Delayed trigger executed after 3 seconds.")
    end, 3.0)
  end
end

pulse_timer:Start(1.0)
```

---

## 5. NamedControl API

A static class used to read and write Q-SYS Named Controls directly from scripts.

### Static Methods
*   **`NamedControl.SetValue(name, value, [ramp_time])`**: Sets a control value (with optional ramp time in seconds).
*   **`NamedControl.GetValue(name)`**: Gets the current control value.
*   **`NamedControl.SetString(name, string)`**: Sets the control display string.
*   **`NamedControl.GetString(name)`**: Gets the control display string.
*   **`NamedControl.SetPosition(name, position, [ramp_time])`**: Sets control position (0.0 to 1.0).
*   **`NamedControl.GetPosition(name)`**: Gets control position (0.0 to 1.0).
*   **`NamedControl.Trigger(name)`**: Triggers a trigger-type control button.

### Sample Code
```lua
-- Ramp volume fader to -10 dB over 5 seconds
NamedControl.SetValue("FaderVolume", -10, 5)

-- Toggle a Mute button
local current_mute = NamedControl.GetValue("MuteButton")
if current_mute == 1 then
  NamedControl.SetValue("MuteButton", 0)
else
  NamedControl.SetValue("MuteButton", 1)
end

NamedControl.SetString("StatusDisplay", "System is Ready")
```

---

## 6. Component API

Used to interact with Q-SYS Named Components (blocks with Script Access enabled).

### Methods
*   **`Component.New("component_name")`**: Creates a reference to a Named Component.
*   **`Component.GetComponents([codename])`**: Returns a table of all Named Components (filterable by name).
*   **`Component.GetControls(component_obj)`**: Returns a list of all controls and properties within the component.

### Properties/Values access
*   **`Value`**: Numerical value.
*   **`String`**: Display string.
*   **`Position`**: Float position (0.0 to 1.0).
*   **`Boolean`**: Direct boolean assignment.
*   **`EventHandler`**: Event handler triggered when a control value changes.

### Escaping Decimal Control Names
If a control name contains a decimal point (e.g., `31.5Hz.gain`), you must escape the decimal with double backslashes in your script access:
*   **Example**: `my_component["31\\.5Hz.gain"].Value = 3.0`

### Sample Code
```lua
my_amplifier = Component.New("amp_rack_1")
my_mixer = Component.New("main_mixer")

-- Handle Mixer Mute changes
my_mixer.inputs_1_mute.EventHandler = function(ctl)
  if ctl.Boolean == true then
    print("Channel 1 is MUTED")
  else
    print("Channel 1 is UNMUTED")
  end
end

-- Write control value directly
my_amplifier.gain.Value = -6.0

-- Access control with decimal point
my_eq = Component.New("room_eq")
my_eq["62\\.5Hz.gain"].Value = -1.5
```
