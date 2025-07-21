package main

import (
	"fmt"
	"log"

	new_mcp_instance "github.com/AbhinavPalacharla/xtrn-personal/internal/new-mcp-instance"
)

func main() {
	app, err := new_mcp_instance.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize MCP instance: %v\n", err)
	}

	//Start server app.startServer()
	s := new_mcp_instance.NewHTTPServer(app)
	if err := s.StartServer(); err != nil {
		panic(fmt.Sprintf("Failed to start HTTP Server - %v", err))
	}
}
