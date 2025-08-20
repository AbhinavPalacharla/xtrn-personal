package mcp_server_images

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
)

var GmailEnvSchema = map[string]string{
	"CLIENT_ID":     "$provider.oauth_client_id",
	"CLIENT_SECRET": "$provider.oauth_client_secret",
	"REFRESH_TOKEN": "$user.oauth_refresh_token",
	"USER_EMAIL":    "",
}

// User should supply everything except the pre-filled fields
func NewGmailEnv(userEmail string) map[string]string {
	return map[string]string{
		"USER_EMAIL": userEmail,
	}
}

func NewGmailImage() (*types.MCPServerImage, error) {
	img, err := types.NewMCPServerImage(
		"Gmail",
		"gmail",
		1,
		"gmail-v1",
		models.MCPServerTypeAuthenticatedOauth,
		"gmail",
		GmailEnvSchema,
	)

	if err != nil {
		return nil, err
	}

	return img, nil
}