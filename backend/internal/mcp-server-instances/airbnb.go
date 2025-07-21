package mcp_server_instances

import "github.com/AbhinavPalacharla/xtrn-personal/internal/types"

const AIRBNB_IMAGE_ID = "airbnb-v1"

func NewAirBNBInstance(userEnv map[string]string) (*types.MCPServerInstance, error) {
	img, err := types.NewMCPServerInstace(AIRBNB_IMAGE_ID, userEnv)

	if err != nil {
		return nil, err
	}

	return img, nil
}
