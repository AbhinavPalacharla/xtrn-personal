package new_mcp_instance

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func (app *App) parseFlags() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  ./new-mcp-client [options]

Options:
  --instance-id     	string		MCP instance ID (Required)
  --docker-image    	string		Docker image name (Required)
  --instance-env    	string		JSON-encoded ENV object (default: "{}") (Required)
  --callback-address	string		Address to write listener address to
  --help, -h                 Show this help message`)
	}

	// Flags
	instanceID := flag.String("instance-id", "", "MCP instance ID")
	dockerImage := flag.String("docker-image", "", "Docker image name")
	instanceEnvRaw := flag.String("instance-env", "{}", `JSON-encoded ENV object (default: "{}")`)
	callbackAddress := flag.String("callback-address", "", "Address for instance to run on")

	// Handle help
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			flag.Usage()
			os.Exit(0)
		}
	}

	flag.Parse()

	// Convert instance-env to map[string]string
	envMap := make(map[string]string)

	fmt.Printf("\n\nRAW INSTANCE ENV: %s\n", *instanceEnvRaw)

	if err := json.Unmarshal([]byte(*instanceEnvRaw), &envMap); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid --instance-env value: %v\n", err)
		os.Exit(1)
	}

	app.InstanceID = *instanceID
	app.DockerImage = *dockerImage
	app.InstanceEnv = envMap
	app.CallbackAddress = *callbackAddress
}

func (app *App) PANIC(reason string) {
	e := errors.New(reason)
	app.ErrLogger.Print(e)
	panic(e)
}

func (app *App) init() {
	app.parseFlags()
	instanceLoggers := shared.NewMCPInstanceLogger(app.InstanceID)
	app.Logger = instanceLoggers.Logger
	app.ErrLogger = instanceLoggers.ErrLogger
}
