package types

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
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

type MCPServerImage struct {
	ID          string               `json:"id"`
	Slug        string               `json:"slug"`
	Version     int                  `json:"version"`
	Name        string               `json:"name"`
	DockerImage string               `json:"docker_image"`
	ServerType  models.MCPServerType `json:"type"`
	Provider    string               `json:"provider"`
	EnvSchema   models.EnvSchema     `json:"env_schema"`
}

func saveMCPServerImageToDB(img *MCPServerImage) error {
	err := DB.InsertMCPServerImage(context.Background(), db.InsertMCPServerImageParams{
		ID:          img.ID,
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
	})

	if err != nil {
		return err
	}

	return nil
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

	s := MCPServerImage{
		ID:          slug + "-v" + strconv.Itoa(version),
		Slug:        slug,
		Version:     version,
		Name:        name,
		DockerImage: dockerImage,
		ServerType:  serverType,
		Provider:    provider,
		EnvSchema:   envSchema,
	}

	if err := saveMCPServerImageToDB(&s); err != nil {
		return nil, err
	}

	return &s, nil
}
