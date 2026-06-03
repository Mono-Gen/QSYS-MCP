package tools

import (
	"errors"
	"fmt"
	"strings"

	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

func resolveSnapshotTarget(params map[string]interface{}) (string, int, int, error) {
	name, _ := params["name"].(string)
	
	var bank int
	var number int
	hasBankNum := false

	if bankFloat, ok := params["bank"].(float64); ok {
		bank = int(bankFloat)
		if numFloat, ok := params["number"].(float64); ok {
			number = int(numFloat)
			hasBankNum = true
		}
	}

	if name == "" && !hasBankNum {
		return "", 0, 0, errors.New("provide either 'name' or both 'bank' and 'number'")
	}

	return name, bank, number, nil
}

func RegisterSnapshotTools(s *mcp.Server) {
	// qsys_list_snapshots
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_list_snapshots",
		Description: "List all snapshots available in the running Q-Sys design.",
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

		var result interface{}
		result, err = qrc.Call("SnapshotBank.GetAll", nil)
		if err != nil {
			// Fallback: execute Lua code on older firmware
			luaCode := strings.Join([]string{
				"local out = {}",
				"for bank = 1, 10 do",
				"  for num = 1, 10 do",
				"    local ok, name = pcall(Snapshot.GetName, bank, num)",
				"    if ok and name and name ~= '' then",
				"      table.insert(out, bank .. '/' .. num .. ': ' .. name)",
				"    end",
				"  end",
				"end",
				"return table.concat(out, '\\n')",
			}, "\n")

			luaRes, err := qrc.Call("Lua.Execute", map[string]interface{}{"code": luaCode})
			if err != nil {
				return "", err
			}

			output := ""
			if obj, ok := luaRes.(map[string]interface{}); ok {
				output, _ = obj["result"].(string)
			}
			
			if strings.TrimSpace(output) == "" {
				return "No snapshots found (Lua query)", nil
			}
			return fmt.Sprintf("Snapshots (via Lua):\n%s", output), nil
		}

		banks, ok := result.([]interface{})
		if !ok || len(banks) == 0 {
			return "No snapshot banks found in the running design.", nil
		}

		var lines []string
		for _, b := range banks {
			bankObj, ok := b.(map[string]interface{})
			if !ok {
				continue
			}

			bankNum, _ := bankObj["Bank"].(float64)
			snapshots, _ := bankObj["Snapshots"].([]interface{})

			for _, snap := range snapshots {
				snapObj, ok := snap.(map[string]interface{})
				if !ok {
					continue
				}

				num, _ := snapObj["Number"].(float64)
				name, _ := snapObj["Name"].(string)
				if name == "" {
					name = "(unnamed)"
				}

				lines = append(lines, fmt.Sprintf("  Bank %g, #%g: %s", bankNum, num, name))
			}
		}

		if len(lines) == 0 {
			return "No snapshots found.", nil
		}

		return fmt.Sprintf("Snapshots (%d):\n%s", len(lines), strings.Join(lines, "\n")), nil
	})

	// qsys_load_snapshot
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_load_snapshot",
		Description: "Load (trigger) a Q-Sys snapshot by name or by bank and number.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "Snapshot name (use this OR bank+number, not both).",
				},
				"bank": {
					Type:        "number",
					Description: "Snapshot bank number (use with 'number').",
				},
				"number": {
					Type:        "number",
					Description: "Snapshot number within the bank (use with 'bank').",
				},
			},
			Required: []string{},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		name, bank, number, err := resolveSnapshotTarget(params)
		if err != nil {
			return "", err
		}

		qrc, ecp, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		if qrc.IsConnected() {
			payload := make(map[string]interface{})
			if name != "" {
				payload["Name"] = name
			} else {
				payload["Bank"] = bank
				payload["Number"] = number
			}

			_, err = qrc.Call("Snapshot.Load", payload)
			if err == nil {
				if name != "" {
					return fmt.Sprintf("Loaded snapshot '%s' [via QRC]", name), nil
				}
				return fmt.Sprintf("Loaded snapshot bank=%d #%d [via QRC]", bank, number), nil
			}
		}

		// ECP fallback
		if name != "" {
			err = ecp.LoadSnapshotByName(name)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Loaded snapshot '%s' [via ECP]", name), nil
		}

		err = ecp.LoadSnapshotByPosition(bank, number)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Loaded snapshot bank=%d #%d [via ECP]", bank, number), nil
	})

	// qsys_save_snapshot
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_save_snapshot",
		Description: "Save the current state of controls to a Q-Sys snapshot by name or bank/number.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "Snapshot name (use this OR bank+number, not both).",
				},
				"bank": {
					Type:        "number",
					Description: "Snapshot bank number (use with 'number').",
				},
				"number": {
					Type:        "number",
					Description: "Snapshot number within the bank (use with 'bank').",
				},
			},
			Required: []string{},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		name, bank, number, err := resolveSnapshotTarget(params)
		if err != nil {
			return "", err
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		payload := make(map[string]interface{})
		if name != "" {
			payload["Name"] = name
		} else {
			payload["Bank"] = bank
			payload["Number"] = number
		}

		_, err = qrc.Call("Snapshot.Save", payload)
		if err != nil {
			return "", err
		}

		if name != "" {
			return fmt.Sprintf("Saved snapshot '%s'", name), nil
		}
		return fmt.Sprintf("Saved snapshot bank=%d #%d", bank, number), nil
	})

	// qsys_run_lua
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_run_lua",
		Description: "Execute a Lua script on the Q-Sys Core and return any output. WARNING: Executes arbitrary code.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"code": {
					Type:        "string",
					Description: "Lua code to execute on the Core.",
				},
			},
			Required: []string{"code"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		code, _ := params["code"].(string)
		if strings.TrimSpace(code) == "" {
			return "", errors.New("'code' is required")
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("Lua.Execute", map[string]interface{}{"code": code})
		if err != nil {
			return "", err
		}

		output := ""
		if obj, ok := res.(map[string]interface{}); ok {
			if r, found := obj["result"]; found {
				output = fmt.Sprintf("%v", r)
			} else if out, found := obj["output"]; found {
				output = fmt.Sprintf("%v", out)
			} else {
				// The result object itself
				output = fmt.Sprintf("%v", obj)
			}
		} else {
			output = fmt.Sprintf("%v", res)
		}

		if output == "" {
			return "(no output)", nil
		}
		return output, nil
	})
}
