package mcp_server_instances

import "github.com/AbhinavPalacharla/xtrn-personal/internal/types"

const GOOGLE_CALENDAR_IMAGE_ID = "google-calendar-v1"

// const GOOGLE_CALENDAR_IMAGE_ID = "xmcp-google-calendar"

func NewGoogleCalendarInstance(userEnv map[string]string) (*types.MCPServerInstance, error) {
	img, err := types.NewMCPServerInstace(GOOGLE_CALENDAR_IMAGE_ID, userEnv)

	if err != nil {
		return nil, err
	}

	return img, nil
}
