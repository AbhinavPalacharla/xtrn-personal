package mcp_server_images

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
)

var AirBNBImageEnvSchema = map[string]string{}

func NewAirBNBImage() (*types.MCPServerImage, error) {
	img, err := types.NewMCPServerImage(
		"AirBNB",
		"airbnb",
		1,
		"airbnb-mcp",
		models.MCPServerTypePublic,
		"",
		AirBNBImageEnvSchema,
	)

	if err != nil {
		return nil, err
	}

	return img, nil
}
