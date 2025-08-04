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
	// s.app.ErrLogger.Printf("RAW REQUEST: %v\n", req)

	rb, _ := json.MarshalIndent(req, "", "   ")
	s.app.ErrLogger.Printf("RAW REQUEST: %s\n", rb)

	toolCallRequest := mcp.CallToolRequest{}
	toolCallRequest.Params.Name = req.Name
	toolCallRequest.Params.Arguments = req.Arguments

	ctx, cancel := context.WithTimeout(context.Background(), MAX_TOOL_USE_TIME)
	defer cancel()

	res, err := s.app.InstanceClient.CallTool(ctx, toolCallRequest)

	if err != nil {
		s.app.ErrLogger.Printf("TOOL CALL ERROR: %s\n", err.Error())
	}

	_ = res

	b, err := json.MarshalIndent(res, "", "  ")

	if err != nil {
		s.app.ErrLogger.Printf("TOOL CALL MARSHALL ERROR: %s\n", err.Error())
	}

	_ = b
	s.app.ErrLogger.Printf("RAW TOOL RESPONSE: %s\n", string(b))

	//Tool call took too long
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "Tool call request timeout", http.StatusRequestTimeout)
		s.app.ErrLogger.Printf("Tool call request timeout - %v\n", err)
	} else if err != nil {
		// Tool call failed for unknown reason
	} else {
		// Tool call response recieved

		// Check xtrn header for type of request either REQUEST or ERROR
		xtrnHeaderRaw := res.Content[0].(mcp.TextContent).Text

		b, _ := json.MarshalIndent(xtrnHeaderRaw, "", "    ")
		s.app.ErrLogger.Printf("XTRN HEADER: %s\n", string(b))

		xtrnHeader := struct {
			XtrnMessageType string  `json:"xtrn_message_type"` // "ERROR" or "RESPONSE"
			ErrorType       *string `json:"error_type,omitempty"`
			Message         *string `json:"message,omitempty"`
		}{}

		err := json.Unmarshal([]byte(xtrnHeaderRaw), &xtrnHeader)

		if err != nil {
			s.app.ErrLogger.Printf("Failed to unmarshall header - %v\n", err)
		}

		if xtrnHeader.XtrnMessageType == "RESPONSE" {
			// contentBytes := []byte(res.Content[1].(mcp.TextContent).Text)

			toolCallRes := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content:   res.Content[1:],
				IsError:   false,
			}

			ViewObjectAsJSON("TOOL CALL RES", toolCallRes, s.app.ErrLogger.Printf)

			// b, _ := json.MarshalIndent(toolCallRes, "", "    ")

			// s.app.ErrLogger.Printf("TOOL CALL RES: %s\n", string(b))

			if err := HTTPSendJSON(w, toolCallRes, nil); err != nil {
				s.app.ErrLogger.Printf("Failed to send JSON - %v\n", err)
			}
		} else if xtrnHeader.XtrnMessageType == "LLM_ERROR_RESPONSE" {

			toolCallRes := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content:   res.Content[1:],
				IsError:   true,
			}

			HTTPSendJSON(w, toolCallRes, &JSONResponseOptions{
				StatusCode: http.StatusBadRequest,
			})
		} else if xtrnHeader.XtrnMessageType == "ERROR" {
			if *xtrnHeader.ErrorType == "AUTH_INVALID_GRANT" {
				HTTPReturnError(w, ErrorOptions{
					Err:  fmt.Sprintf("ERROR: %s REASON: %s", *xtrnHeader.ErrorType, *xtrnHeader.Message),
					Code: http.StatusUnauthorized,
				})
			}
		} else {
			// Unsupported type
		}

	}
}

func (s *HTTPServer) handleKill(w http.ResponseWriter, r *http.Request) {
	s.app.Logger.Printf("KILL REQUEST RECIEVED\n")

	s.app.InstanceClient.Close()

	s.app.Listener.Close()

	Q.DeleteMCPServerInstance(context.Background(), s.app.InstanceID)

	HTTPSendJSON[any](w, nil, &JSONResponseOptions{})
}
