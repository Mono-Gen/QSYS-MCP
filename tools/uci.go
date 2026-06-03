package tools

import (
	"encoding/json"
	"errors"
	"fmt"

	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
)

func RegisterUciTools(s *mcp.Server) {
	// qsys_set_uci_page
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_set_uci_page",
		Description: "Switch the active page of a User Control Interface (UCI) on the Q-Sys Core.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "UCI name as defined in Q-Sys Designer (e.g. 'Tablet_UCI').",
				},
				"page": {
					Type:        "string",
					Description: "Name of the page to switch to (e.g. 'Main', 'Audio Settings').",
				},
			},
			Required: []string{"name", "page"},
		},
	}, func(params map[string]interface{}) (string, error) {
		coreAlias, _ := params["core"].(string)
		name, _ := params["name"].(string)
		page, _ := params["page"].(string)

		if name == "" {
			return "", errors.New("'name' is required")
		}
		if page == "" {
			return "", errors.New("'page' is required")
		}

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		_, err = qrc.Call("Uci.Set", map[string]interface{}{
			"Name": name,
			"Page": page,
		})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Successfully set UCI '%s' page to '%s'.", name, page), nil
	})

	// qsys_get_uci_status
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_uci_status",
		Description: "Get the current status of a User Control Interface (UCI), including the currently active page.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"core": {
					Type:        "string",
					Description: "Core alias (omit if only one Core is configured).",
				},
				"name": {
					Type:        "string",
					Description: "UCI name as defined in Q-Sys Designer.",
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

		qrc, _, err := qsys.CurrentConnectionManager.GetClients(qsys.CoreAlias(coreAlias))
		if err != nil {
			return "", err
		}

		res, err := qrc.Call("Uci.Get", map[string]interface{}{
			"Name": name,
		})
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
