package new_mcp_instance

import (
	"fmt"
	"log"
	"net"

	"github.com/mark3labs/mcp-go/client"
)

type App struct {
	InstanceID      string            //cmd arg
	DockerImage     string            //cmd arg
	InstanceEnv     map[string]string //cmd arg
	CallbackAddress string            //cmd arg
	Address         string            //runtime
	Logger          *log.Logger       //runtime
	ErrLogger       *log.Logger       //runtime
	InstanceClient  *client.Client    //runtime
	Listener        net.Listener      //runtime
}

func NewApp() (*App, error) {
	app := App{}
	app.init()

	client, err := app.createClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to start MCP Client - %w", err)
	}

	app.InstanceClient = client

	return &app, nil
}
