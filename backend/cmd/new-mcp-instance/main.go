package main

import (
	"flag"
	"fmt"
	"os"
)

type App struct {
	InstanceName string
}

func (app *App) parseFlags() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
		  ./mcpClient --mcp-server-image <json> [--env <json>]

		Options:
		  --mcp-server-image		JSON-encoded MCPServerImage object (required)
		  --server-env              JSON-encoded env variables (optional)
		  --help, -h                Show this help message
		`+"\n")
	}

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			flag.Usage()
			os.Exit(0)
		}
	}

	serverImageJSON := flag.String("mcp-server", "", "JSON-encoded types/MCPServerImage object")
	serverEnvJSON := flag.String("env", "", "JSON-encoded env variables")

	flag.Parse()
}

func (app *App) initApp() {

}

func main() {

	fmt.Printf("LAUNCHING MCP INSTANCE - %s\n")
}
