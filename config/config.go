package config

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	PollingInterval   int
	Debug             bool
	ProtectedPatterns []*regexp.Regexp
	StylesDir         string
	LuaReferencePath  string
	CssReferencePath  string
}

type jsonConfig struct {
	Cores             string `json:"cores"`
	StylesDir         string `json:"styles_dir"`
	PollingInterval   int    `json:"polling_interval"`
	Debug             bool   `json:"debug"`
	ProtectedControls string `json:"protected_controls"`
	LuaReferencePath  string `json:"lua_reference_path"`
	CssReferencePath  string `json:"css_reference_path"`
}

var CurrentConfig *Config

func InitConfig() {
	// If config.json exists, load it and set variables in the environment
	loadConfigJson()

	pollingInterval := 350
	if intervalStr, ok := os.LookupEnv("QSYS_POLLING_INTERVAL"); ok {
		if val, err := strconv.Atoi(intervalStr); err == nil {
			pollingInterval = val
		}
	}

	debug := os.Getenv("QSYS_MCP_DEBUG") == "true"

	var protectedPatterns []*regexp.Regexp
	if protectedStr, ok := os.LookupEnv("QSYS_PROTECTED_CONTROLS"); ok && protectedStr != "" {
		parts := strings.Split(protectedStr, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				if r, err := regexp.Compile("(?i)" + trimmed); err == nil {
					protectedPatterns = append(protectedPatterns, r)
				}
			}
		}
	}

	stylesDir := os.Getenv("QSYS_STYLES_DIR")
	luaReferencePath := os.Getenv("QSYS_LUA_REFERENCE_PATH")
	cssReferencePath := os.Getenv("QSYS_CSS_REFERENCE_PATH")

	CurrentConfig = &Config{
		PollingInterval:   pollingInterval,
		Debug:             debug,
		ProtectedPatterns: protectedPatterns,
		StylesDir:         stylesDir,
		LuaReferencePath:  luaReferencePath,
		CssReferencePath:  cssReferencePath,
	}

	if debug {
		log.Printf("[q-sys-mcp-debug] Config initialized: PollingInterval=%dms, ProtectedPatterns=%d, StylesDir=%q, LuaReferencePath=%q, CssReferencePath=%q\n", pollingInterval, len(protectedPatterns), stylesDir, luaReferencePath, cssReferencePath)
	}
}

func loadConfigJson() {
	file, err := os.Open("config.json")
	if err != nil {
		// Do nothing if config.json does not exist (fallback to environment variables)
		return
	}
	defer file.Close()

	var jc jsonConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jc); err != nil {
		log.Printf("[q-sys-mcp-warning] Failed to parse config.json: %v\n", err)
		return
	}

	if jc.Cores != "" {
		os.Setenv("QSYS_CORES", jc.Cores)
	}
	if jc.StylesDir != "" {
		os.Setenv("QSYS_STYLES_DIR", jc.StylesDir)
	}
	if jc.PollingInterval != 0 {
		os.Setenv("QSYS_POLLING_INTERVAL", strconv.Itoa(jc.PollingInterval))
	}
	if jc.Debug {
		os.Setenv("QSYS_MCP_DEBUG", "true")
	}
	if jc.ProtectedControls != "" {
		os.Setenv("QSYS_PROTECTED_CONTROLS", jc.ProtectedControls)
	}
	if jc.LuaReferencePath != "" {
		os.Setenv("QSYS_LUA_REFERENCE_PATH", jc.LuaReferencePath)
	}
	if jc.CssReferencePath != "" {
		os.Setenv("QSYS_CSS_REFERENCE_PATH", jc.CssReferencePath)
	}
}

func DebugLog(format string, v ...interface{}) {
	if CurrentConfig != nil && CurrentConfig.Debug {
		log.Printf("[q-sys-mcp] "+format, v...)
	}
}

func IsProtected(controlName string) bool {
	if CurrentConfig == nil {
		return false
	}
	for _, pattern := range CurrentConfig.ProtectedPatterns {
		if pattern.MatchString(controlName) {
			return true
		}
	}
	return false
}
