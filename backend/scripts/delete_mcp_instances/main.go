package main

import (
	"context"

	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	if err := Q.DeleteAllMCPinstances(context.Background()); err != nil {
		panic("❌ Failed to delete MCP Instances")
	}
}
