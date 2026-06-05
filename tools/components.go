package tools

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

func RegisterComponentTools(s *mcp.Server) {
	// qsys_list_components
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_list_components",
		Description: "List all named components in the running Q-Sys design.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"filter": {
					Type:        "string",
					Description: "Optional regex to filter results by component name or type (case-insensitive).",
				},
			},
			Required: []string{},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		filterParam, _ := params["filter"].(string)

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("Component.GetComponents", nil)
		if err != nil {
			return "", err
		}

		var components []map[string]interface{}
		if arr, ok := res.([]interface{}); ok {
			for _, item := range arr {
				if obj, ok := item.(map[string]interface{}); ok {
					components = append(components, obj)
				}
			}
		}

		if len(components) == 0 {
			return "No named components found in the running design.", nil
		}

		var filterRegex *regexp.Regexp
		if filterParam != "" {
			r, err := regexp.Compile("(?i)" + filterParam)
			if err != nil {
				return "", fmt.Errorf("invalid filter regex: %w", err)
			}
			filterRegex = r
		}

		var lines []string
		for _, c := range components {
			name, _ := c["Name"].(string)
			compType, _ := c["Type"].(string)

			if filterRegex != nil {
				if !filterRegex.MatchString(name) && !filterRegex.MatchString(compType) {
					continue
				}
			}

			lines = append(lines, fmt.Sprintf("- %s (%s)", name, compType))
		}

		if len(lines) == 0 {
			return fmt.Sprintf("No components matched filter %q.", filterParam), nil
		}

		suffix := ""
		if filterParam != "" {
			suffix = fmt.Sprintf(" matching %q", filterParam)
		}

		return fmt.Sprintf("Components in running design (%d%s):\n%s", len(lines), suffix, strings.Join(lines, "\n")), nil
	})

	// qsys_get_component_controls
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_component_controls",
		Description: "Get all controls for a named component — fader positions, mute states, EQ bands, etc.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"component": {
					Type:        "string",
					Description: "Component name exactly as it appears in the Q-Sys design.",
				},
			},
			Required: []string{"component"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		component, _ := params["component"].(string)
		if component == "" {
			return "", errors.New("'component' is required")
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("Component.GetControls", map[string]interface{}{"Name": component})
		if err != nil {
			return "", err
		}

		var controls []interface{}
		if arr, ok := res.([]interface{}); ok {
			controls = arr
		} else if obj, ok := res.(map[string]interface{}); ok {
			if ctrlList, found := obj["Controls"].([]interface{}); found {
				controls = ctrlList
			}
		}

		if len(controls) == 0 {
			return fmt.Sprintf("No controls found for component '%s'.", component), nil
		}

		var lines []string
		for _, item := range controls {
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

		return fmt.Sprintf("%s controls (%d):\n%s", component, len(lines), strings.Join(lines, "\n")), nil
	})

	// qsys_set_component_controls
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_set_component_controls",
		Description: "Set one or more controls on a named Q-Sys component in a single call.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"component": {
					Type:        "string",
					Description: "Component name exactly as it appears in the Q-Sys design.",
				},
				"controls": {
					Type:        "array",
					Description: "List of controls to set.",
				},
			},
			Required: []string{"component", "controls"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		component, _ := params["component"].(string)
		if component == "" {
			return "", errors.New("'component' is required")
		}

		controlsInterface, _ := params["controls"].([]interface{})
		if len(controlsInterface) == 0 {
			return "", errors.New("'controls' must be a non-empty array")
		}

		var blocked []string
		var payload []map[string]interface{}
		var summary []string

		for i, item := range controlsInterface {
			obj, ok := item.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("controls[%d] is not a valid object", i)
			}

			name, _ := obj["name"].(string)
			if name == "" {
				return "", fmt.Errorf("controls[%d]: 'name' is required", i)
			}

			fullName := fmt.Sprintf("%s.%s", component, name)
			if config.IsProtected(fullName) {
				blocked = append(blocked, fullName)
				continue
			}

			val, hasValue := obj["value"]
			pos, hasPosition := obj["position"].(float64)

			if !hasValue && obj["position"] == nil && !hasPosition {
				return "", fmt.Errorf("controls[%d] ('%s'): provide either 'value' or 'position'", i, name)
			}

			payloadItem := map[string]interface{}{
				"Name": name,
			}

			what := ""
			// If position is provided, validate it
			if obj["position"] != nil {
				if pos < 0.0 || pos > 1.0 {
					return "", fmt.Errorf("controls[%d] ('%s'): position must be 0.0-1.0, got %f", i, name, pos)
				}
				payloadItem["Position"] = pos
				what = fmt.Sprintf("position=%g", pos)
			} else {
				payloadItem["Value"] = val
				what = fmt.Sprintf("value=%v", val)
			}

			payload = append(payload, payloadItem)
			summary = append(summary, fmt.Sprintf("  %s -> %s", name, what))
		}

		if len(blocked) > 0 {
			return "", fmt.Errorf("write blocked — protected control(s): %s", strings.Join(blocked, ", "))
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		_, err = qrc.Call("Component.Set", map[string]interface{}{
			"Name":     component,
			"Controls": payload,
		})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Set %d control(s) on '%s':\n%s", len(payload), component, strings.Join(summary, "\n")), nil
	})
}
