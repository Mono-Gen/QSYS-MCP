package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
)

func RegisterLocalCssTools(s *mcp.Server) {
	// qsys_write_local_css
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_write_local_css",
		Description: "Create or overwrite a CSS style sheet in the Q-Sys Designer Styles directory.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"filename": {
					Type:        "string",
					Description: "The CSS filename to write (e.g. 'custom.css'). Must end in .css.",
				},
				"content": {
					Type:        "string",
					Description: "The CSS style contents to write into the file.",
				},
			},
			Required: []string{"filename", "content"},
		},
	}, func(params map[string]interface{}) (string, error) {
		filename, _ := params["filename"].(string)
		content, _ := params["content"].(string)

		if filename == "" {
			return "", errors.New("'filename' is required")
		}
		if content == "" {
			return "", errors.New("'content' is required")
		}

		// Validation: Prevent directory traversal (ensure no path separators are present)
		if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
			return "", errors.New("invalid filename: directory traversal or subdirectories are not allowed")
		}

		// Extension check: only .css is allowed
		if !strings.HasSuffix(strings.ToLower(filename), ".css") {
			return "", errors.New("invalid filename: must have .css extension")
		}

		stylesDir := config.CurrentConfig.StylesDir
		if stylesDir == "" {
			return "", errors.New("Styles directory is not configured. Please set 'styles_dir' in config.json or QSYS_STYLES_DIR environment variable")
		}

		// Verify or create the directory if it does not exist
		if err := os.MkdirAll(stylesDir, 0755); err != nil {
			return "", fmt.Errorf("failed to verify or create styles directory: %w", err)
		}

		targetPath := filepath.Join(stylesDir, filename)
		err := os.WriteFile(targetPath, []byte(content), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write CSS file: %w", err)
		}

		return fmt.Sprintf("Successfully wrote CSS style to %s", filename), nil
	})

	// qsys_read_local_css
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_read_local_css",
		Description: "Read the contents of a CSS style sheet from the Q-Sys Designer Styles directory.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"filename": {
					Type:        "string",
					Description: "The CSS filename to read (e.g. 'custom.css'). Must end in .css.",
				},
			},
			Required: []string{"filename"},
		},
	}, func(params map[string]interface{}) (string, error) {
		filename, _ := params["filename"].(string)

		if filename == "" {
			return "", errors.New("'filename' is required")
		}

		// Validation: Prevent directory traversal (ensure no path separators are present)
		if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
			return "", errors.New("invalid filename: directory traversal or subdirectories are not allowed")
		}

		// Extension check: only .css is allowed
		if !strings.HasSuffix(strings.ToLower(filename), ".css") {
			return "", errors.New("invalid filename: must have .css extension")
		}

		stylesDir := config.CurrentConfig.StylesDir
		if stylesDir == "" {
			return "", errors.New("Styles directory is not configured. Please set 'styles_dir' in config.json or QSYS_STYLES_DIR environment variable")
		}

		targetPath := filepath.Join(stylesDir, filename)
		data, err := os.ReadFile(targetPath)
		if err != nil {
			return "", fmt.Errorf("failed to read CSS file: %w", err)
		}

		return string(data), nil
	})

	// qsys_get_css_reference
	s.RegisterTool(mcp.Tool{
		Name:        "qsys_get_css_reference",
		Description: "Get the CSS styling reference for Q-Sys UCI Styles (selectors, class naming, specific properties).",
		InputSchema: mcp.InputSchema{
			Type: "object",
		},
	}, func(params map[string]interface{}) (string, error) {
		refPath := config.CurrentConfig.CssReferencePath
		if refPath == "" {
			return "", errors.New("CSS reference path is not configured. Please set 'css_reference_path' in config.json")
		}

		data, err := os.ReadFile(refPath)
		if err != nil {
			return "", fmt.Errorf("failed to read CSS reference file from %s: %w", refPath, err)
		}

		return string(data), nil
	})
}
