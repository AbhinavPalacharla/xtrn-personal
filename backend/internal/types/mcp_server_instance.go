package types

import gonanoid "github.com/matoous/go-nanoid/v2"

type MCPServerInstance struct {
	MCPServerImage
	ID      string
	UserEnv map[string]any
	Address string
}

func NewMCPServerInstace(image *MCPServerImage, userEnv map[string]any) (*MCPServerInstance, error) {
	id, _ := gonanoid.New()

	fmtID := image.ID + "-inst-" + id

	inst := MCPServerInstance{
		ID: fmtID,
	}

	return &inst, nil
}
