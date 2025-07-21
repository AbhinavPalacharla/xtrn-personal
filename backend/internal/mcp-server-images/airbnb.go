package mcp_server_images

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
)

var AirBNBEnvSchema = map[string]string{}

func NewAirBNBEnv() map[string]string {
	return map[string]string{}
}

func NewAirBNBImage() (*types.MCPServerImage, error) {
	img, err := types.NewMCPServerImage(
		"AirBNB",
		"airbnb",
		1,
		"airbnb",
		models.MCPServerTypePublic,
		"",
		AirBNBEnvSchema,
	)

	if err != nil {
		return nil, err
	}

	return img, nil
}
