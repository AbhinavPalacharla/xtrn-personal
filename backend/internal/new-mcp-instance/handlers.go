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
	ToolUseID string        `json:"tool_use_id"`
	Content   []mcp.Content `json:"content"`
	IsError   bool          `json:"is_error"`
}

type XtrnMessageType string

const (
	XtrnMessageTypeResponse XtrnMessageType = "RESPONSE"
	XtrnMessageTypeError    XtrnMessageType = "ERROR"
	XtrnMessageTypeLLMError XtrnMessageType = "LLM_ERROR_RESPONSE"
)

type XtrnHeader struct {
	XtrnMessageType XtrnMessageType `json:"xtrn_message_type"`
	ErrorType       *string         `json:"error_type,omitempty"`
	Message         *string         `json:"message,omitempty"`
}

func (s *HTTPServer) handleCallTool(w http.ResponseWriter, r *http.Request) {
	req, err := DecodeJSONBody[ToolCallReq](r, w)
	if err != nil {
		return
	}

	ViewObjectAsJSON("RAW REQUEST", req, s.app.Logger.Printf)

	toolCallRequest := mcp.CallToolRequest{}
	toolCallRequest.Params.Name = req.Name
	toolCallRequest.Params.Arguments = req.Arguments

	ctx, cancel := context.WithTimeout(context.Background(), MAX_TOOL_USE_TIME)
	defer cancel()

	res, err := s.app.InstanceClient.CallTool(ctx, toolCallRequest)

	if err != nil {
		s.app.ErrLogger.Printf("TOOL CALL ERROR: %s\n", err.Error())
	}

	ViewObjectAsJSON("RAW TOOL RESPONSE", res, s.app.Logger.Printf)

	if errors.Is(err, context.DeadlineExceeded) {

		//Tool call took too long
		HTTPReturnError(w, ErrorOptions{
			Err:  "Tool call request timeout",
			Code: http.StatusRequestTimeout,
		})

		s.app.ErrLogger.Printf("Tool call request timeout - %v\n", err)

	} else if err != nil {

		// Tool call failed for unknown reason
		HTTPReturnError(w, ErrorOptions{
			Err:  err.Error(),
			Code: http.StatusInternalServerError,
		})
		s.app.ErrLogger.Printf("Tool call failed for unknown reason - %v", err)

	} else {

		// Check xtrn header for type of request either REQUEST or ERROR
		xtrnHeaderRaw := res.Content[0].(mcp.TextContent).Text

		var xtrnHeader XtrnHeader

		if err := json.Unmarshal([]byte(xtrnHeaderRaw), &xtrnHeader); err != nil {
			s.app.ErrLogger.Printf("Failed to unmarshall header - %v\n", err)
		}

		ViewObjectAsJSON("XTRN HEADER", xtrnHeader, s.app.Logger.Printf)

		if xtrnHeader.XtrnMessageType == XtrnMessageTypeResponse {
			// REGULAR RESPONSE

			toolCallRes := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content:   res.Content[1:],
				IsError:   false,
			}

			ViewObjectAsJSON("TOOL CALL RES", toolCallRes, s.app.ErrLogger.Printf)

			if err := HTTPSendJSON(w, toolCallRes, nil); err != nil {
				s.app.ErrLogger.Printf("Failed to send JSON - %v\n", err)
			}

		} else if xtrnHeader.XtrnMessageType == XtrnMessageTypeLLMError {
			// LLM ERROR

			toolCallRes := ToolCallRes{
				ToolUseID: req.ToolUseID,
				Content:   res.Content[1:],
				IsError:   true,
			}

			ViewObjectAsJSON("TOOL CALL RES (LLM ERR)", toolCallRes, s.app.ErrLogger.Printf)

			HTTPSendJSON(w, toolCallRes, &JSONResponseOptions{
				StatusCode: http.StatusBadRequest,
			})
		} else if xtrnHeader.XtrnMessageType == XtrnMessageTypeError {
			// XTRN ERROR

			s.app.ErrLogger.Printf("AUTH INVALID - %v", xtrnHeader)

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

	Q.DeleteMCPServerInstance(context.Background(), s.app.InstanceID)

	HTTPSendJSON[any](w, nil, &JSONResponseOptions{})

	// s.app.Listener.Close()
}
