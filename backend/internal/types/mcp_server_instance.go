package types

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type MCPServerInstance struct {
	MCPServerImage
	InstanceID  string
	InstanceEnv models.EnvSchema
	Address     string
}

const GOOGLE_REFRESH_TOKEN = ""

func saveMCPServerInstanceToDB(inst *MCPServerInstance) error {
	return Q.InsertMCPServerInstance(context.Background(), db.InsertMCPServerInstanceParams{
		ID:      inst.InstanceID,
		Slug:    inst.Slug,
		Version: int64(inst.Version),
		Address: inst.Address,
		Env:     inst.InstanceEnv,
	})
}

func addCommandFlag(cmd *string, flag string, value string) {
	*cmd += " " + flag + "=" + value
}

func startMCPServerInstance(inst *MCPServerInstance) error {
	socketPath := fmt.Sprintf("/tmp/xtrn/%s.sock", inst.InstanceID)
	os.Remove(socketPath)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	//Start Server
	binDir := os.Getenv("BIN_DIR")
	if binDir == "" {
		panic("`BIN_DIR` Environment vairable not set run: eval $(make setup-env)")
	}

	//Format command
	// command := fmt.Sprintf("%s/start-mcp-instance", binDir)
	// addCommandFlag(&command, "--instance-id", inst.InstanceID)
	// addCommandFlag(&command, "--docker-image", inst.DockerImage)
	// addCommandFlag(&command, "--callback-address", socketPath) //When instance HTTP server is ready it will write it to the socket

	// envJSON, _ := json.Marshal(inst.InstanceEnv)
	// addCommandFlag(&command, "--instance-env", string(envJSON)) //Pass env as json string

	// commandArgs := strings.Split(command, " ")

	commandArgs := []string{
		fmt.Sprintf("%s/start-mcp-instance", binDir),
		"--instance-id=" + inst.InstanceID,
		"--docker-image=" + inst.DockerImage,
		"--callback-address=" + socketPath,
	}
	envJSON, _ := json.Marshal(inst.InstanceEnv)
	commandArgs = append(commandArgs, "--instance-env="+string(envJSON))

	fmt.Print(commandArgs)

	//Run command
	cmd := exec.Command(commandArgs[0], commandArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	//Read address from socket into inst.addr
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)

	//Save address
	inst.Address = "http://" + string(buf[:n])

	return nil
}

func NewMCPServerInstace(imageID string, userEnv map[string]string) (*MCPServerInstance, error) {

	img, err := Q.GetMCPServerImage(context.Background(), imageID)
	fmt.Printf("IMAGE: %v", img)
	if err != nil {
		return nil, err
	}

	// Instance ID
	id, _ := gonanoid.New()
	instID := img.ID + "-inst-" + id

	instanceEnv := models.EnvSchema{}

	for k, v := range img.EnvSchema {
		//Handle template values
		if v == "$provider.oauth_client_id" {
			instanceEnv[k] = img.ClientID.String

		} else if v == "$provider.oauth_client_secret" {
			instanceEnv[k] = img.ClientSecret.String
		} else if v == "$user.oauth_refresh_token" {
			// instanceEnv[k] = "somerefreshtoken123"
			// instanceEnv[k] = GOOGLE_REFRESH_TOKEN //TODO: Change to goth eventually

			//Get refresh token
			token, err := Q.GetOauthTokenByProvider(context.Background(), img.OauthProvider.String)
			if err != nil {
				return nil, err
			}

			instanceEnv[k] = token.RefreshToken
		} else {
			//If not template then see if it is in userEnv if not then invalid user schema

			if val, ok := userEnv[k]; ok {
				instanceEnv[k] = val
			} else {
				// return nil, errors.New("Invalid User Schema - Missing key %v\n\nExpected Schema: %v\n")
				return nil, fmt.Errorf("Invalid User Schema - Missing key `%v`\n\nExpected Schema: %v\n", k, userEnv)
			}
		}
	}

	inst := MCPServerInstance{
		MCPServerImage: MCPServerImage{
			Slug:        img.Slug,
			Version:     int(img.Version),
			DockerImage: img.DockerImage,
		},
		InstanceID:  instID,
		InstanceEnv: instanceEnv,
	}

	// fmt.Printf("%#v\n", inst)

	if err := startMCPServerInstance(&inst); err != nil {
		return nil, err
	}

	if err := saveMCPServerInstanceToDB(&inst); err != nil {
		return nil, err
	}

	return &inst, nil
}
