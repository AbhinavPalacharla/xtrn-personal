package mcp_server_images

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/db/models"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
)

var GoogleCalendarEnvSchema = map[string]string{
	"CLIENT_ID":     "$provider.oauth_client_id",
	"CLIENT_SECRET": "$provider.oauth_client_secret",
	"REFRESH_TOKEN": "$user.oauth_refresh_token",
	"WORK_CALENDAR": "",
}

// User should supply everything except the pre-filled fields
func NewGoogleCalendarEnv(workCalendar string) map[string]string {
	return map[string]string{
		"WORK_CALENDAR": workCalendar,
	}
}

func NewGoogleCalendarImage() (*types.MCPServerImage, error) {
	img, err := types.NewMCPServerImage(
		"Google Calendar",
		"google-calendar",
		1,
		"google-calendar-v1",
		models.MCPServerTypeAuthenticatedOauth,
		"google-calendar",
		GoogleCalendarEnvSchema,
	)

	if err != nil {
		return nil, err
	}

	return img, nil
}
