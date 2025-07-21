package new_mcp_instance

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func formatEnv(env map[string]string) []string {
	// fmtEnv := ""

	// for k, v := range env {
	// 	if fmtEnv != "" {
	// 		fmtEnv += " "
	// 	}

	// 	fmtEnv += fmt.Sprintf("-e %s=%s", k, v)
	// }

	// return fmtEnv

	fmtEnv := []string{}

	for k, v := range env {
		// fmtEnv = append(fmtEnv, fmt.Sprintf("-e \"%s=%s\"", k, v))
		fmtEnv = append(fmtEnv, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	return fmtEnv
}

func (app *App) formatMCPServerStartCommand() []string {
	// command := ""

	// if len(app.InstanceEnv) > 0 {
	// 	command = fmt.Sprintf("docker run --name %s -i --rm %s %s", app.InstanceID, formatEnv(app.InstanceEnv), app.DockerImage)
	// } else {
	// 	command = fmt.Sprintf("docker run --name %s -i --rm %s", app.InstanceID, app.DockerImage)
	// }

	// commandArgs := strings.Split(command, " ")

	command := []string{
		"docker",
		"run",
		"--name", app.InstanceID,
		"-i", "--rm",
	}

	if len(app.InstanceEnv) > 0 {
		command = append(command, formatEnv(app.InstanceEnv)...)
	}

	command = append(command, app.DockerImage)

	return command
}

func (app *App) createClient() (*client.Client, error) {
	mcpServerStartCommand := app.formatMCPServerStartCommand()

	fmt.Printf("\nRunning container with command: %v\n", mcpServerStartCommand)
	fmt.Printf("DOCKER CONTAINER: %s\n", app.DockerImage)
	app.Logger.Printf("\n\nRunning container with command: %v\n", mcpServerStartCommand)

	transport := transport.NewStdio(mcpServerStartCommand[0], nil, mcpServerStartCommand[1:]...)
	transport.Start(context.Background())

	instanceLoggers := shared.NewMCPInstanceLogger(app.InstanceID)
	errorLogger := instanceLoggers.ErrLogger

	//Async stderr logging
	go func() {
		stderr := transport.Stderr()
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			errorLogger.Print(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			errorLogger.Printf("Could not read stderr: %v\n", err)
		}
	}()

	client := client.NewClient(transport)

	//Give container 30 sconds to start up
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Start MCP Server + MCP Client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    fmt.Sprint(app.InstanceID),
		Version: "0",
	}
	initResult, err := client.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}

	//Client initialized with server
	app.Logger.Printf("CLIENT INITIALIZED\n↳ Server Info: %s %s\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	return client, nil
}
