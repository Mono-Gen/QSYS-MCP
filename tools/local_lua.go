package tools

import (
	"errors"
	"fmt"
	"os"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
)

func RegisterLocalLuaTools(s *mcp.Server) {
	// qsys_get_lua_reference
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_lua_reference",
		Description: "Get the Q-Sys Lua scripting API reference (TcpSocket, HttpClient, Timer, NamedControl, Component) and code templates.",
		InputSchema: mcp.InputSchema{
			Type: "object",
		},
	}, func(params map[string]interface{}) (string, error) {
		refPath := config.CurrentConfig.LuaReferencePath
		if refPath == "" {
			return "", errors.New("Lua reference path is not configured. Please set 'lua_reference_path' in config.json")
		}

		data, err := os.ReadFile(refPath)
		if err != nil {
			return "", fmt.Errorf("failed to read Lua reference file from %s: %w", refPath, err)
		}

		return string(data), nil
	})
}
