package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

func RegisterConnectionTools(s *mcp.Server) {
	// qsys_list_cores
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_list_cores",
		Description: "List all configured Cores and their connection states.",
		InputSchema: mcp.InputSchema{
			Type:       "object",
			Properties: map[string]mcp.Property{},
			Required:   []string{},
		},
	}, func(params map[string]interface{}) (string, error) {
		cores := qsys.CurrentConnectionManager.ListCores()
		if len(cores) == 0 {
			return "No Cores configured. Set QSYS_CORES environment variable.", nil
		}

		var lines []string
		for _, c := range cores {
			isDefaultStr := ""
			if c["isDefault"].(bool) {
				isDefaultStr = " (default)"
			}
			lines = append(lines, fmt.Sprintf("  Alias: %s%s, Host: %s, State: %s", c["alias"], isDefaultStr, c["host"], c["state"]))
		}

		return fmt.Sprintf("Configured Cores (%d):\n%s", len(cores), strings.Join(lines, "\n")), nil
	})

	// qsys_core_status
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_core_status",
		Description: "Get detailed engine status from Q-Sys Core (Platform, Design Name, Status).",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
			},
			Required: []string{},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("Engine.Status", nil)
		if err != nil {
			return "", err
		}

		data, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return "", err
		}

		return string(data), nil
	})
}
