package new_mcp_instance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/mark3labs/mcp-go/mcp"
)

const MAX_TOOL_USE_TIME = time.Second * 20

func (s *HTTPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools, err := s.app.InstanceClient.ListTools(context.Background(), mcp.ListToolsRequest{})

	if err != nil {
		shared.HTTPReturnError(w, ErrorOptions{
			Err: fmt.Sprintf("Failed to list tools - %v", err),
		})

		return
	}

	if err := json.NewEncoder(w).Encode(tools); err != nil {
		shared.HTTPReturnError(w, ErrorOptions{
			Err: fmt.Sprintf("Failed to encode tools - %v\n", err),
		})

		return
	}
}

type ToolCallReq struct {
	ToolUseID string         `json:"tool_use_id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCallRes struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"` //JSON string of MCP response
	IsError   bool   `json:"is_error"`
}

func (s *HTTPServer) handleCallTool(w http.ResponseWriter, r *http.Request) {
	req, err := DecodeJSONBody[ToolCallReq](r, w)
	if err != nil {
		return
	}

	toolCallRequest := mcp.CallToolRequest{}
	toolCallRequest.Params.Name = req.Name
	toolCallRequest.Params.Arguments = req.Arguments

	ctx, cancel := context.WithTimeout(context.Background(), MAX_TOOL_USE_TIME)
	defer cancel()

	res, err := s.app.InstanceClient.CallTool(ctx, toolCallRequest)

	//Tool call took too long
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "Tool call request timeout", http.StatusRequestTimeout)
		s.app.ErrLogger.Printf("Tool call request timeout - %v\n", err)
	} else if err != nil {
		// Tool call failed
		toolCallResult := ToolCallRes{
			ToolUseID: req.ToolUseID,
			Content:   "",
			IsError:   true,
		}
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(toolCallResult)
		if err != nil {
			http.Error(w, "JSON encoding error", http.StatusInternalServerError)
			s.app.ErrLogger.Printf("Failed to encode json - %v\n", err)
			return
		}
		s.app.ErrLogger.Printf("Tool Call failed - %v\n", err)
	} else {
		contentJSONBytes, _ := json.Marshal(res.Content)
		toolCallResult := ToolCallRes{
			ToolUseID: req.ToolUseID,
			Content:   string(contentJSONBytes),
			IsError:   false,
		}

		err := HTTPSendJSON(w, toolCallResult, nil)

		if err != nil {
			s.app.ErrLogger.Printf("Failed to send Tool Call Result - %v\n", err)
		}
	}
}

func (s *HTTPServer) handleKill(w http.ResponseWriter, r *http.Request) {}
