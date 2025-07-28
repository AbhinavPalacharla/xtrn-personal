package types

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrInvalidClientID = errors.New("Invalid CLIENT_ID value must be $provider.oauth_client_id")
var ErrInvalidClientSecret = errors.New("Invalid CLIENT_SECRET value must be $provider.oauth_client_id")
var ErrInvalidRefreshToken = errors.New("Invalid REFRESH_TOKEN value must be $user.oauth_refresh_token")

func validateEnvSchema(schema map[string]string) (bool, error) {
	for k, v := range schema {
		if k == "CLIENT_ID" && v != "$provider.oauth_client_id" {
			return false, ErrInvalidClientID
		} else if k == "CLIENT_SECRET" && v != "$provider.oauth_client_secret" {
			return false, ErrInvalidClientSecret
		} else if k == "REFRESH_TOKEN" && v != "$user.oauth_refresh_token" {
			return false, ErrInvalidRefreshToken
		}
	}

	return true, nil
}

type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
}

type MCPServerImage struct {
	ImageID     string               `json:"id"`
	Slug        string               `json:"slug"`
	Version     int                  `json:"version"`
	Name        string               `json:"name"`
	DockerImage string               `json:"docker_image"`
	ServerType  models.MCPServerType `json:"type"`
	Provider    string               `json:"provider"`
	EnvSchema   models.EnvSchema     `json:"env_schema"`
	Tools       []MCPTool            `json:"tools"`
}

func saveMCPServerImageToDB(img *MCPServerImage) error {
	tx, err := DB.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer func() {
		if r := recover(); r != nil { // Handle panics
			tx.Rollback()
			panic(r) // re-throw panic
		} else if err != nil { // Handle explicit errors
			log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback()
		}
	}()

	qtx := Q.WithTx(tx)

	if err := qtx.InsertMCPServerImage(context.Background(), db.InsertMCPServerImageParams{
		ID:          img.ImageID,
		Slug:        img.Slug,
		Version:     int64(img.Version),
		Name:        img.Name,
		DockerImage: img.DockerImage,
		Type:        string(img.ServerType),
		OauthProvider: sql.NullString{
			String: img.Provider,
			Valid:  img.Provider != "",
		},
		EnvSchema: img.EnvSchema,
	}); err != nil {
		// return err
		return fmt.Errorf("%w", err)

	}

	for _, tool := range img.Tools {
		id, _ := gonanoid.New()

		if err := qtx.InsertMCPServerInstanceTool(context.Background(), db.InsertMCPServerInstanceToolParams{
			ID:   id,
			Name: tool.Name,
			Description: sql.NullString{
				String: tool.Description,
				Valid:  tool.Description != "",
			},
			Schema:  tool.InputSchema,
			ImageID: img.ImageID,
		}); err != nil {
			return fmt.Errorf("Tools: %w", err)

		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Failed to insert MCP server instance by tx - %w", err)
	}

	return nil
}

func (img *MCPServerImage) getTools() ([]MCPTool, error) {
	tools := []MCPTool{}

	cmd := []string{"docker", "run", "-i", "--rm"}

	for k, v := range img.EnvSchema {
		_ = v
		cmd = append(cmd, "-e", fmt.Sprintf("%s=%s", k, "abc")) // Just need placeholder values to get tools
	}

	cmd = append(cmd, img.DockerImage)

	fmt.Printf("STARTING SERVER WITH CMD: %v\n", cmd)

	transport := transport.NewStdio(cmd[0], nil, cmd[1:]...)

	transport.Start(context.Background())

	//Async stderr logging
	go func() {
		stderr := transport.Stderr()
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			fmt.Print(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Could not read stderr: %v\n", err)
		}
	}()

	client := client.NewClient(transport)

	//Give container 30 sconds to start up
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Start MCP Server + MCP Client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "",
		Version: "0",
	}
	_, err := client.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}

	toolsRes, err := client.ListTools(ctx, mcp.ListToolsRequest{})

	for _, tool := range toolsRes.Tools {
		schemaBytes, err := json.Marshal(tool.RawInputSchema)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal tool input schema - %w\n", err)
		}

		tools = append(tools, MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: string(schemaBytes),
		})
	}

	return tools, nil
}

func NewMCPServerImage(name string,
	slug string,
	version int,
	dockerImage string,
	serverType models.MCPServerType,
	provider string,
	envSchema map[string]string,
) (*MCPServerImage, error) {
	//Validation
	if !serverType.IsValid() {
		return nil, models.InvalidMCPTypeError
	}

	if ok, err := validateEnvSchema(envSchema); !ok {
		return nil, fmt.Errorf("\nInvalid MCP Image Env Schema\n\t%w\n", err)
	}

	// Start a temporary instance to get tools

	s := MCPServerImage{
		ImageID:     slug + "-v" + strconv.Itoa(version),
		Slug:        slug,
		Version:     version,
		Name:        name,
		DockerImage: dockerImage,
		ServerType:  serverType,
		Provider:    provider,
		EnvSchema:   envSchema,
	}
	tools, err := s.getTools()
	if err != nil {
		return nil, err
	}
	s.Tools = tools

	if err := saveMCPServerImageToDB(&s); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &s, nil
}
