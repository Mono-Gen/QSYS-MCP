import socket
import json
import os

def send_lua_file_to_qsys():
    # Connection details
    ip = "127.0.0.1"  # Q-SYS Core IP Address
    port = 1710  # QRC port
    component_name = "Text_Controller"
    
    # Path to the Lua file to send (using relative path from CWD)
    lua_filename = "us41_control.lua"
    
    if not os.path.exists(lua_filename):
        print(f"Error: {lua_filename} not found in the current directory.")
        return
        
    print(f"Reading {lua_filename}...")
    with open(lua_filename, "r", encoding="utf-8") as f:
        lua_code = f.read()

    # Construct QRC Component.Set command
    payload = {
        "jsonrpc": "2.0",
        "method": "Component.Set",
        "params": {
            "Name": component_name,
            "Controls": [
                {
                    "Name": "code",
                    "Value": lua_code
                }
            ]
        },
        "id": 1
    }

    print(f"Connecting to Q-SYS Core at {ip}:{port}...")
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(5.0)
            s.connect((ip, port))
            print("Connected!")
            
            # Send JSON data (with null character \x00 at the end)
            message = json.dumps(payload) + "\x00"
            s.sendall(message.encode('utf-8'))
            print(f"Successfully loaded and sent {lua_filename} to Q-SYS component '{component_name}'.")
            
            # Receive response
            response = s.recv(4096)
            response_text = response.decode('utf-8', errors='ignore').strip('\x00')
            print("\nRaw response from Q-SYS Core:")
            print(repr(response_text))
            
            # Parse responses individually
            for part in response_text.split('\x00'):
                if part.strip():
                    try:
                        parsed = json.loads(part)
                        print("\nParsed JSON Response:")
                        print(json.dumps(parsed, indent=2))
                    except json.JSONDecodeError:
                        pass
            
    except Exception as e:
        print(f"Error occurred: {e}")

if __name__ == "__main__":
    send_lua_file_to_qsys()
