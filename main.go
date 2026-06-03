package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"qsys-mcp/config"
	"qsys-mcp/mcp"
	"qsys-mcp/qsys"
	"qsys-mcp/tools"
)

func main() {
	// Log settings: stdio is used by the MCP protocol, so redirect all standard log output to Stderr
	log.SetOutput(os.Stderr)

	// Initialize configuration
	config.InitConfig()

	// Initialize connection manager
	if err := qsys.InitConnectionManager(); err != nil {
		log.Fatalf("Fatal: failed to initialize connection manager: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer()

	// Register various tools
	tools.RegisterConnectionTools(server)
	tools.RegisterControlTools(server)
	tools.RegisterComponentTools(server)
	tools.RegisterSnapshotTools(server)
	tools.RegisterChangeGroupTools(server)
	tools.RegisterUciTools(server)
	tools.RegisterLocalCssTools(server)
	tools.RegisterLocalLuaTools(server)

	fmt.Fprintln(os.Stderr, "[q-sys-mcp] Server running on stdio")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Fprintln(os.Stderr, "[q-sys-mcp] Shutting down...")
		if qsys.CurrentConnectionManager != nil {
			qsys.CurrentConnectionManager.DisconnectAll()
		}
		os.Exit(0)
	}()

	// Start server (bind stdin and stdout)
	server.Start(os.Stdin, os.Stdout)
}
