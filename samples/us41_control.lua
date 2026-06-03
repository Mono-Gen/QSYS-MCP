-- IMAGENICS US-41 HDMI Selector Control Script for Q-SYS Text Controller
-- 2026-05-30

-- Configuration
local IP_ADDRESS = "192.168.2.254" -- IP address of the HDMI selector
local PORT = 1300                 -- Default TCP port number
local POLL_INTERVAL = 1.0          -- Interval for querying status (seconds)

-- UI Controls
-- Button control definitions (Trigger or StateTrigger)
local btn_inputs = {
  [1] = Controls.btn_input_1,
  [2] = Controls.btn_input_2,
  [3] = Controls.btn_input_3,
  [4] = Controls.btn_input_4,
  [0] = Controls.btn_off        -- For OFF state
}

-- Colors
local COLOR_ACTIVE = "#00FF00"   -- Button color when active (green)
local COLOR_INACTIVE = "#808080" -- Button color when inactive (grey)

-- TCP Socket
local conn = TcpSocket.New()
conn.ReadTimeout = 5.0
conn.WriteTimeout = 5.0

-- Status variables
local current_input = -1 -- Current input state (-1: unknown, 0: OFF, 1-4: input 1-4)

-- Functions
-- Update button color and state
local function updateUI(active_input)
  current_input = active_input
  for idx, ctrl in pairs(btn_inputs) do
    if ctrl then
      if idx == active_input then
        ctrl.Color = COLOR_ACTIVE
        -- Set Boolean property if it is a StateTrigger
        if ctrl.Boolean ~= nil then
          ctrl.Boolean = true
        end
      else
        ctrl.Color = COLOR_INACTIVE
        if ctrl.Boolean ~= nil then
          ctrl.Boolean = false
        end
      end
    end
  end
end

-- Send status query command
local function queryStatus()
  if conn.IsConnected then
    conn:Write("z1\r")
  end
end

-- Timer setup
local pollTimer = Timer.New()
pollTimer.EventHandler = function()
  queryStatus()
end

-- Socket Events
conn.Connected = function(sock)
  print("US-41: Connected to " .. IP_ADDRESS .. ":" .. PORT)
  pollTimer:Start(POLL_INTERVAL)
  queryStatus() -- Query status immediately after connection
end

conn.Closed = function(sock, err)
  print("US-41: Disconnected. Error: " .. (err or "None"))
  pollTimer:Stop()
  -- Start reconnection timer
  Timer.CallAfter(function()
    if not conn.IsConnected then
      print("US-41: Retrying connection...")
      conn:Connect(IP_ADDRESS, PORT)
    end
  end, 5.0)
end

conn.Data = function(sock)
  -- Read incoming data
  local data = conn:Read(conn.BufferLength)
  if not data then return end
  
  -- Parse received data (e.g. "002\r")
  -- Matrix switcher z1 query response format is "000\r" to "004\r"
  local status = string.match(data, "(%d%d%d)")
  if status then
    local active_input = tonumber(status)
    if active_input then
      updateUI(active_input)
    end
  end
end

-- UI Event Handlers
-- Send command on button press
local function setupUI()
  for idx, ctrl in pairs(btn_inputs) do
    if ctrl then
      ctrl.EventHandler = function()
        if conn.IsConnected then
          local cmd = ""
          if idx == 0 then
            cmd = "q,1\r" -- OFF
          else
            cmd = tostring(idx) .. ",1\r" -- Switch input
          end
          conn:Write(cmd)
          -- After sending, wait briefly for the reply or poll immediately
          Timer.CallAfter(queryStatus, 0.1)
        else
          print("US-41: Cannot send command. Socket not connected.")
        end
      end
    end
  end
end

-- Initialize
setupUI()
print("US-41: Initializing connection to " .. IP_ADDRESS .. "...")
conn:Connect(IP_ADDRESS, PORT)
