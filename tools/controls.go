package tools

import (
	"errors"
	"fmt"
	"strings"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

type ControlResult struct {
	Name     string
	Value    float64
	Position float64
	String   string
	Legend   string
	Via      string
}

func formatControl(c ControlResult) string {
	via := ""
	if c.Via != "" {
		via = " [via " + c.Via + "]"
	}
	legendStr := ""
	if c.Legend != "" {
		legendStr = fmt.Sprintf("\n  legend:   %s", c.Legend)
	}
	return fmt.Sprintf("%s:%s\n  value:    %g\n  position: %.4f\n  string:   %s%s",
		c.Name, via, c.Value, c.Position, c.String, legendStr)
}

func RegisterControlTools(s *mcp.Server) {
	// qsys_get_control
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_control",
		Description: "Get the current value, position (0-1), string label, and legend of a named Q-Sys control.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "Named control to read (e.g. 'MainPA.gain', 'LecternMic.mute').",
				},
			},
			Required: []string{"name"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		name, _ := params["name"].(string)
		if name == "" {
			return "", errors.New("'name' is required")
		}

		qrc, ecp, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		if qrc.IsConnected() {
			res, err := qrc.Call("Control.Get", map[string]interface{}{"Name": name})
			if err == nil {
				// QRC response is either an array or an object
				var val float64
				var pos float64
				var strVal string
				var legend string

				if arr, ok := res.([]interface{}); ok && len(arr) > 0 {
					if obj, ok := arr[0].(map[string]interface{}); ok {
						val, _ = obj["Value"].(float64)
						pos, _ = obj["Position"].(float64)
						strVal, _ = obj["String"].(string)
						legend, _ = obj["Legend"].(string)
					}
				} else if obj, ok := res.(map[string]interface{}); ok {
					val, _ = obj["Value"].(float64)
					pos, _ = obj["Position"].(float64)
					strVal, _ = obj["String"].(string)
					legend, _ = obj["Legend"].(string)
				}

				return formatControl(ControlResult{
					Name:     name,
					Value:    val,
					Position: pos,
					String:   strVal,
					Legend:   legend,
					Via:      "QRC",
				}), nil
			}
		}

		// ECP fallback
		res, err := ecp.GetControl(name)
		if err != nil {
			return "", err
		}

		return formatControl(ControlResult{
			Name:     name,
			Value:    res.Value,
			Position: res.Position,
			String:   res.String,
			Via:      "ECP",
		}), nil
	})

	// qsys_set_control
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_set_control",
		Description: "Set a named Q-Sys control by value or normalized position.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "Named control to set.",
				},
				"value": {
					Type:        "number",
					Description: "Raw control value (dB for gains, 0/1 for mutes, etc.).",
				},
				"position": {
					Type:        "number",
					Description: "Normalized position 0.0-1.0.",
				},
			},
			Required: []string{"name"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		name, _ := params["name"].(string)
		if name == "" {
			return "", errors.New("'name' is required")
		}

		if config.IsProtected(name) {
			return "", fmt.Errorf("write blocked — '%s' is a protected control", name)
		}

		val, hasValue := params["value"].(float64)
		pos, hasPosition := params["position"].(float64)

		if !hasValue && !hasPosition {
			return "", errors.New("provide either 'value' or 'position'")
		}

		if hasPosition && (pos < 0.0 || pos > 1.0) {
			return "", fmt.Errorf("'position' must be 0.0-1.0, got %f", pos)
		}

		qrc, ecp, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		if qrc.IsConnected() {
			payload := make(map[string]interface{})
			payload["Name"] = name
			if hasPosition {
				payload["Position"] = pos
			} else {
				payload["Value"] = val
			}

			_, err = qrc.Call("Control.Set", payload)
			if err == nil {
				what := ""
				if hasPosition {
					what = fmt.Sprintf("position=%g", pos)
				} else {
					what = fmt.Sprintf("value=%g", val)
				}
				return fmt.Sprintf("Set '%s' -> %s [via QRC]", name, what), nil
			}
		}

		// ECP fallback
		if hasPosition {
			err = ecp.SetControlPosition(name, pos)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Set '%s' -> position=%g [via ECP]", name, pos), nil
		}

		err = ecp.SetControlValue(name, val)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Set '%s' -> value=%g [via ECP]", name, val), nil
	})

	// qsys_get_controls
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_controls",
		Description: "Batch-get multiple named Q-Sys controls in a single call.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"names": {
					Type:        "array",
					Description: "List of named controls to read.",
				},
			},
			Required: []string{"names"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		namesInterface, _ := params["names"].([]interface{})
		if len(namesInterface) == 0 {
			return "", errors.New("'names' must be a non-empty array")
		}

		var names []string
		for _, n := range namesInterface {
			if str, ok := n.(string); ok {
				names = append(names, str)
			}
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		if !qrc.IsConnected() {
			return "", errors.New("QRC not connected — batch get requires QRC (ECP does not support batch)")
		}

		var reqList []map[string]interface{}
		for _, name := range names {
			reqList = append(reqList, map[string]interface{}{"Name": name})
		}

		res, err := qrc.Call("Control.Get", reqList)
		if err != nil {
			return "", err
		}

		var results []string
		if arr, ok := res.([]interface{}); ok {
			for i, item := range arr {
				if obj, ok := item.(map[string]interface{}); ok {
					name := names[i]
					val, _ := obj["Value"].(float64)
					pos, _ := obj["Position"].(float64)
					strVal, _ := obj["String"].(string)
					legend, _ := obj["Legend"].(string)

					results = append(results, formatControl(ControlResult{
						Name:     name,
						Value:    val,
						Position: pos,
						String:   strVal,
						Legend:   legend,
						Via:      "QRC",
					}))
				}
			}
		}

		return strings.Join(results, "\n\n"), nil
	})
}
