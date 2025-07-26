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

	s.app.ErrLogger.Print("TOOL LIST REQUEST RECIEVED")

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

type ToolCallErrRes struct {
	Type      string `json:"type"`
	ErrorCode string `json:"error_code"`
	Error     string `json:"error"`
}

type ToolCallTextRes struct {
	Type string `json:"type"`
	Text string
}

type ToolCallRes struct {
	ToolUseID string `json:"tool_use_id"`
	// Content   string `json:"content"` //JSON string of MCP response
	Content []mcp.Content `json:"content"`
	IsError bool          `json:"is_error"`
}

func (s *HTTPServer) handleCallTool(w http.ResponseWriter, r *http.Request) {
	req, err := DecodeJSONBody[ToolCallReq](r, w)
	if err != nil {
		return
	}

	s.app.ErrLogger.Println("TOOL CALL REQUEST RECIEVED")
	s.app.ErrLogger.Println(req)

	toolCallRequest := mcp.CallToolRequest{}
	toolCallRequest.Params.Name = req.Name
	toolCallRequest.Params.Arguments = req.Arguments

	ctx, cancel := context.WithTimeout(context.Background(), MAX_TOOL_USE_TIME)
	defer cancel()

	res, err := s.app.InstanceClient.CallTool(ctx, toolCallRequest)

	s.app.Logger.Printf("RAW TOOL RESPONSE: %v | err; %v\n", res, err)

	//Tool call took too long
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "Tool call request timeout", http.StatusRequestTimeout)
		s.app.ErrLogger.Printf("Tool call request timeout - %v\n", err)
	} else if err != nil {
		// Tool call failed for external reason
		s.app.ErrLogger.Printf("Tool Call failed - %v\n", err)

		toolCallResult := ToolCallRes{
			ToolUseID: req.ToolUseID,
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("%v", err)),
			},
			IsError: true,
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(toolCallResult)
	} else {
		//Extract main response
		contentBytes := []byte(res.Content[0].(mcp.TextContent).Text)

		fmt.Printf("CONTENT BYTES: %v\n", string(contentBytes))

		type ContentItem struct {
			Type      string `json:"type"`
			Text      string `json:"text,omitempty"`
			ErrorCode string `json:"error_code,omitempty"`
			Error     string `json:"error,omitempty"`
		}

		mainContent := struct {
			IsError bool          `json:"is_error"`
			Content []ContentItem `json:"content"`
		}{}

		if err := json.Unmarshal([]byte(contentBytes), &mainContent); err != nil {
			s.app.ErrLogger.Printf("Failed to unmarshal main content - %v\n", err)

			toolCallResult := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content: []mcp.Content{
					mcp.NewTextContent(fmt.Sprintf("%v", err)),
				},
				IsError: true,
			}

			err := shared.HTTPSendJSON(w, toolCallResult, &JSONResponseOptions{StatusCode: http.StatusInternalServerError})

			if err != nil {
				s.app.ErrLogger.Printf("Failed to send Tool Call Result - %v\n", err)
			}
		}

		s.app.ErrLogger.Printf("MAIN CONTENT: %v\n", mainContent)

		// Check if XTRN error in content
		if c := mainContent.Content[0]; c.Type == "error" {
			errContent := ToolCallErrRes{
				Type:      c.Type,
				ErrorCode: c.ErrorCode,
				Error:     c.Error,
			}

			HTTPSendJSON(w, errContent, &JSONResponseOptions{
				StatusCode: http.StatusUnauthorized,
			})
			return

		} else if c := mainContent.Content[0]; c.Type == "text" {
			textContent := ToolCallTextRes{
				Text: c.Text,
				Type: c.Type,
			}

			toolCallResult := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content: []mcp.Content{
					mcp.NewTextContent(textContent.Text),
				},
				IsError: false,
			}

			err := HTTPSendJSON(w, toolCallResult, nil)

			if err != nil {
				s.app.ErrLogger.Printf("Failed to send Tool Call Result - %v\n", err)
			}
		}
	}
}

func (s *HTTPServer) handleKill(w http.ResponseWriter, r *http.Request) {
	s.app.Logger.Printf("KILL REQUEST RECIEVED\n")

	s.app.InstanceClient.Close()

	s.app.Listener.Close()

	DB.DeleteMCPServerInstance(context.Background(), s.app.InstanceID)
}
