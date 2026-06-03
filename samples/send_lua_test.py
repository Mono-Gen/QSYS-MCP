import socket
import json
import base64

def send_vertical_text_to_qsys():
    # Connection details
    ip = "127.0.0.1"  # Q-SYS Core IP Address
    port = 1710  # Standard port for QRC (Q-SYS Remote Control)
    component_name = "Text_Controller"
    control_name = "Control_1"
    text = "マイク"
    
    # Generate SVG for vertical text
    fontSize = 24
    runeCount = len(text)
    width = float(fontSize) * 1.6
    height = float(fontSize) * float(runeCount) * 1.2 + 20.0
    xPos = width / 2.0
    yPos = height / 2.0
    
    # Raw SVG XML string
    svg = f"""<?xml version="1.0" encoding="utf-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}">
  <text x="{xPos}" y="{yPos}" font-family="sans-serif" font-size="{fontSize}" fill="#FFFFFF" 
        writing-mode="vertical-rl" text-orientation="upright" glyph-orientation-vertical="0"
        text-anchor="middle" alignment-baseline="middle">{text}</text>
</svg>"""

    # Base64-encode the SVG content
    svg_base64 = base64.b64encode(svg.encode('utf-8')).decode('utf-8')

    # Construct Style JSON
    style = {
        "DrawChrome": True,
        "IconData": svg_base64,
        "Legend": ""
    }
    style_json = json.dumps(style)

    # Build the QRC Component.Set command
    payload = {
        "jsonrpc": "2.0",
        "method": "Component.Set",
        "params": {
            "Name": component_name,
            "Controls": [
                {
                    "Name": control_name,
                    "String": style_json
                }
            ]
        },
        "id": 1
    }

    print(f"Connecting to Q-SYS Core at {ip}:{port}...")
    try:
        # Establish TCP socket connection
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(5.0)
            s.connect((ip, port))
            print("Connected!")
            
            # Send JSON data (with null character \x00 at the end)
            message = json.dumps(payload) + "\x00"
            s.sendall(message.encode('utf-8'))
            print(f"Vertical text SVG applied to '{component_name}' -> '{control_name}'.")
            
            # Receive response
            response = s.recv(4096)
            response_text = response.decode('utf-8', errors='ignore').strip('\x00')
            print("Raw response from Q-SYS Core:")
            print(repr(response_text))
            
            try:
                # Multiple messages might be received; attempt to parse each
                for part in response_text.split('\x00'):
                    if part.strip():
                        parsed = json.loads(part)
                        print("Parsed Response:")
                        print(json.dumps(parsed, indent=2))
            except Exception as pe:
                print(f"(Could not parse response as single JSON: {pe})")
            
    except Exception as e:
        print(f"Error occurred: {e}")

if __name__ == "__main__":
    send_vertical_text_to_qsys()
