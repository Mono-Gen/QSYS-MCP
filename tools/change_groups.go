package tools

import (
	"errors"
	"fmt"
	"strings"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

func RegisterChangeGroupTools(s *mcp.Server) {
	// qsys_create_change_group
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_create_change_group",
		Description: "Create a Q-Sys change group and register named controls to monitor.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"id": {
					Type:        "string",
					Description: "Group ID string — choose any unique name (e.g. 'mic-monitor', 'faders').",
				},
				"controls": {
					Type:        "array",
					Description: "Named controls to add to the group.",
				},
			},
			Required: []string{"id", "controls"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		id, _ := params["id"].(string)
		if id == "" {
			return "", errors.New("'id' is required")
		}

		controlsInterface, _ := params["controls"].([]interface{})
		if len(controlsInterface) == 0 {
			return "", errors.New("'controls' must be a non-empty array")
		}

		var controls []map[string]interface{}
		var names []string
		for _, c := range controlsInterface {
			if str, ok := c.(string); ok {
				controls = append(controls, map[string]interface{}{"Name": str})
				names = append(names, str)
			}
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		_, err = qrc.Call("ChangeGroup.AddControl", map[string]interface{}{
			"Id":       id,
			"Controls": controls,
		})
		if err != nil {
			return "", err
		}

		var bulletPoints []string
		for _, name := range names {
			bulletPoints = append(bulletPoints, fmt.Sprintf("  - %s", name))
		}

		return fmt.Sprintf("Created change group '%s' with %d control(s):\n%s", id, len(controls), strings.Join(bulletPoints, "\n")), nil
	})

	// qsys_poll_change_group
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_poll_change_group",
		Description: fmt.Sprintf("Poll a change group for controls that have changed since the last poll. Recommended poll interval: %dms", config.CurrentConfig.PollingInterval),
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"id": {
					Type:        "string",
					Description: "Group ID returned by qsys_create_change_group.",
				},
			},
			Required: []string{"id"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		id, _ := params["id"].(string)
		if id == "" {
			return "", errors.New("'id' is required")
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("ChangeGroup.Poll", map[string]interface{}{"Id": id})
		if err != nil {
			return "", err
		}

		var changes []interface{}
		if obj, ok := res.(map[string]interface{}); ok {
			if list, found := obj["Changes"].([]interface{}); found {
				changes = list
			}
		}

		if len(changes) == 0 {
			return fmt.Sprintf("No changes in group '%s' since last poll.", id), nil
		}

		var lines []string
		for _, item := range changes {
			if obj, ok := item.(map[string]interface{}); ok {
				name, _ := obj["Name"].(string)
				
				valStr := ""
				if val, found := obj["Value"]; found {
					valStr = fmt.Sprintf("value=%v", val)
				}
				
				posStr := ""
				if pos, found := obj["Position"].(float64); found {
					posStr = fmt.Sprintf("pos=%.3f", pos)
				}
				
				strVal := ""
				if str, found := obj["String"].(string); found {
					strVal = fmt.Sprintf("%q", str)
				}

				var parts []string
				if valStr != "" {
					parts = append(parts, valStr)
				}
				if posStr != "" {
					parts = append(parts, posStr)
				}
				if strVal != "" {
					parts = append(parts, strVal)
				}

				lines = append(lines, fmt.Sprintf("  %s: %s", name, strings.Join(parts, "  ")))
			}
		}

		return fmt.Sprintf("Changed controls in group '%s' (%d):\n%s", id, len(changes), strings.Join(lines, "\n")), nil
	})

	// qsys_destroy_change_group
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_destroy_change_group",
		Description: "Destroy a change group and free its resources on the Core.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"id": {
					Type:        "string",
					Description: "Group ID to destroy.",
				},
			},
			Required: []string{"id"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		id, _ := params["id"].(string)
		if id == "" {
			return "", errors.New("'id' is required")
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		_, err = qrc.Call("ChangeGroup.Destroy", map[string]interface{}{"Id": id})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Destroyed change group '%s'.", id), nil
	})
}
