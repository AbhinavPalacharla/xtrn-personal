package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type App struct {
	Mux       *http.ServeMux
	Listener  net.Listener
	Logger    *log.Logger
	ErrLogger *log.Logger
}

func NewApp() (*App, error) {
	a := App{}

	//Configure HTTP Router
	a.Mux = http.NewServeMux()
	//TODO: NEED TO FIGURE OUT BETTER PATHS FOR ROUTES
	a.Mux.HandleFunc("/chats", a.handleChat)                        // SEND MESSAGE
	a.Mux.HandleFunc("/chats/{chatID}/messages", a.handleChat)      // SEND MESSAGE
	a.Mux.HandleFunc("/messages/{chatID}", a.handleGetChatMessages) // GET MESSAGE HISTORY
	a.Mux.HandleFunc("/hello", a.helloHandler)

	if listener, err := net.Listen("tcp", ":8080"); err != nil {
		return nil, err
	} else {
		a.Listener = listener
	}
	loggers := NewAPILoggers()

	a.Logger = loggers.Logger
	a.ErrLogger = loggers.ErrLogger

	return &a, nil
}

func getMessageHistory(chatID string) ([]llms.MessageContent, error) {
	msgHist := []llms.MessageContent{}

	messages, err := Q.GetViewChatMessges(context.Background(), chatID)
	if err != nil {
		return nil, fmt.Errorf("Could not get messages from DB - %w", err)
	}

	for _, m := range messages {

		if m.Role == string(llms.ChatMessageTypeHuman) {
			// Human Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{
						Text: m.Content.String,
					},
				},
			})
		} else if m.Role == string(llms.ChatMessageTypeAI) {
			// AI Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: func() []llms.ContentPart {
					contentParts := []llms.ContentPart{}

					for _, p := range m.AiMessage {
						if p.Type == "text" {
							contentParts = append(contentParts, llms.TextContent{
								Text: func() string {
									if p.Text != nil {
										return *p.Text
									} else {
										return ""
									}
								}(),
							})
						} else if p.Type == "function" {
							contentParts = append(contentParts, llms.ToolCall{
								ID:   p.ToolCallID,
								Type: p.Type,
								FunctionCall: &llms.FunctionCall{
									Name:      p.Name,
									Arguments: string(p.Arguments),
								},
							})
						}
					}

					return contentParts
				}(),
			})
		} else if m.Role == string(llms.ChatMessageTypeTool) {
			// Tool Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{
					ToolCallID: m.ToolResult.ToolCallID,
					Name:       m.ToolResult.Name,
					Content:    string(m.ToolResult.Content),
				}},
			})
		}
	}

	return msgHist, nil
}

func (app *App) handleGetChatMessages(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chatID")

	messages, err := Q.GetViewChatMessges(context.Background(), chatID)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err:  fmt.Errorf("Could not get messages - %w", err).Error(),
			Code: http.StatusBadRequest,
		})
		app.ErrLogger.Print(err)
	}

	ViewObjectAsJSON("RAW MESSAGES", messages, nil)
	fmt.Print("\n\n\n")

	msgHist := []llms.MessageContent{}

	for _, m := range messages {

		if m.Role == string(llms.ChatMessageTypeHuman) {
			// Human Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{
						Text: m.Content.String,
					},
				},
			})
		} else if m.Role == string(llms.ChatMessageTypeAI) {
			// AI Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: func() []llms.ContentPart {
					contentParts := []llms.ContentPart{}

					for _, p := range m.AiMessage {
						if p.Type == "text" {
							contentParts = append(contentParts, llms.TextContent{
								Text: func() string {
									if p.Text != nil {
										return *p.Text
									} else {
										return ""
									}
								}(),
							})
						} else if p.Type == "function" {
							contentParts = append(contentParts, llms.ToolCall{
								ID:   p.ToolCallID,
								Type: p.Type,
								FunctionCall: &llms.FunctionCall{
									Name:      p.Name,
									Arguments: string(p.Arguments),
								},
							})
						}
					}

					return contentParts
				}(),
			})
		} else if m.Role == string(llms.ChatMessageTypeTool) {
			// Tool Message

			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{llms.ToolCallResponse{
					ToolCallID: m.ToolResult.ToolCallID,
					Name:       m.ToolResult.Name,
					Content:    string(m.ToolResult.Content),
				}},
			})
		}
	}

	ViewObjectAsJSON("MESSAGE HISTORY", msgHist, nil)

	// HTTPSendJSON(w, msgHist, nil)
}

type Message struct {
	Content string `json:"content"`
}

type MCPInstanceTool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ChatMCPInstance struct {
	ID      string
	Address string
	ImgID   string
	Tools   []MCPInstanceTool
}

func getMCPTools() ([]llms.Tool, map[string]string, error) {
	// Fetch MCP instance tools
	instanceRows, err := Q.GetMCPServerInstances(context.Background())

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get MPC instances - %w", err)
	}

	instances := map[string]*ChatMCPInstance{}

	for _, i := range instanceRows {

		inst, ok := instances[i.InstanceID]
		if !ok {
			inst = &ChatMCPInstance{
				ID:      i.InstanceID,
				Address: i.Address,
				ImgID:   i.ImageID.String,
				Tools:   []MCPInstanceTool{},
			}
			ok = true
			instances[i.InstanceID] = inst
		}

		if ok {
			schema := map[string]any{}
			json.Unmarshal([]byte(i.ToolSchema.String), &schema)

			inst.Tools = append(inst.Tools, MCPInstanceTool{
				Name:        i.ToolName.String,
				Description: i.ToolDesc.String,
				InputSchema: schema,
			})
		}
	}

	ViewObjectAsJSON("MCP INSTANCES", instances, nil)

	tools := []llms.Tool{}
	toolToAddr := map[string]string{}

	for _, inst := range instances {
		for _, tool := range inst.Tools {
			toolName := inst.ID + "___" + tool.Name

			toolToAddr[toolName] = inst.Address

			tools = append(tools, llms.Tool{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        toolName,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
	}

	return tools, toolToAddr, nil
}

func (app *App) handleChat(w http.ResponseWriter, r *http.Request) {
	//CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("x-vercel-ai-ui-message-stream", "v1")

	//XTRN frontend headers
	w.Header().Set("Access-Control-Expose-Headers", "x-xtrn-chat-id")

	newChat := false
	chatID := r.PathValue("chatID")
	if chatID == "" {
		// If not chatID then this is the first message - need to init the chat and possibly configure tools specially
		id, _ := gonanoid.New()
		chatID = id
		newChat = true
	}

	w.Header().Set("x-xtrn-chat-id", chatID)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to read request body - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}
	defer r.Body.Close()

	ViewObjectAsJSON("RAW MESSAGE", bodyBytes, nil)

	msg := Message{}

	if err := json.Unmarshal(bodyBytes, &msg); err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err:  fmt.Errorf("Malformed request body - %w", err).Error(),
			Code: http.StatusBadRequest,
		})
		app.ErrLogger.Print(err)
	}

	ViewObjectAsJSON("MESSAGE RECIEVED", msg, nil)

	/***************** INITIALIZATION *****************/
	tx, _ := DB.BeginTx(context.Background(), nil)
	defer tx.Rollback()

	qtx := Q.WithTx(tx)

	if newChat {
		// Create chat in DB
		qtx.InsertChat(context.Background(), chatID)
	}

	msgID, _ := gonanoid.New()

	// Add message to chat
	err = qtx.InsertMessage(context.Background(), db.InsertMessageParams{
		ID:   msgID,
		Role: string(llms.ChatMessageTypeHuman),
		Content: sql.NullString{
			String: msg.Content,
			Valid:  msg.Content != "",
		},
		StopReason: sql.NullString{
			String: "",
			Valid:  false,
		},
		ChatID: chatID,
	})

	if err != nil {
		fmt.Printf("ERROR INSERTING MESSAGE - %v\n", err)
	}

	if err = tx.Commit(); err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to insert message - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
		return
	}

	tools, toolToAddr, err := getMCPTools()
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: err.Error(),
		})
		app.ErrLogger.Print(err)
	}
	_ = toolToAddr

	openAIKey, err := GetEnv("OPENAI_KEY")
	if err != nil {
		panic(err)
	}

	llm, err := openai.New(
		openai.WithToken(openAIKey),
		openai.WithModel("gpt-4.1-mini"),
	)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to create LLM instance - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	/***************** CHAT LOOP *****************/

	msgHist, err := getMessageHistory(chatID)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err:  fmt.Errorf("Failed to get messages for chatID=%s - %w", chatID, err).Error(),
			Code: http.StatusInternalServerError,
		})
	}

	// Get initial response from LLM
	resp, err := llm.GenerateContent(context.Background(), msgHist, llms.WithTools(tools))
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to get response from LLM - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	fmt.Println("LLM RESPONSE SUCCEEDED")
	ViewObjectAsJSON("RAW LLM RESPONSE", resp, nil)

	// Add LLM response to msg history
	msgHist, err = updateMessageHistory(w, context.Background(), chatID, msgHist, resp)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to update message history - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
		return
	}

	ViewObjectAsJSON("PRE-TOOLS MSG HIST LLM", msgHist, nil)

	// Execute tool calls
	for {
		// Check if there are any tool calls to execute and execute them
		if len(resp.Choices[0].ToolCalls) == 0 {
			break // No tools needed so end conversation loop and wait for user to send next message
		}

		msgHist, err = execToolCalls(chatID, msgHist, resp, toolToAddr)
		if err != nil {
			HTTPReturnError(w, ErrorOptions{
				Err: fmt.Errorf("Failed to save execute tool call - %w", err).Error(),
			})
			app.ErrLogger.Print(err)
			return
		}

		ViewObjectAsJSON("MSG HIST TOOL", msgHist, nil)

		// Call LLM again with tool call responses
		resp, err = llm.GenerateContent(context.Background(), msgHist, llms.WithTools(tools))
		if err != nil {
			HTTPReturnError(w, ErrorOptions{
				Err: fmt.Errorf("Failed to get response from LLM - %w", err).Error(),
			})
			app.ErrLogger.Print(err)
			return
		}

		// Add LLM response to msg history
		msgHist, err = updateMessageHistory(w, context.Background(), chatID, msgHist, resp)
		if err != nil {
			HTTPReturnError(w, ErrorOptions{
				Err: fmt.Errorf("Failed to update message history - %w", err).Error(),
			})
			app.ErrLogger.Print(err)
			return
		}

		ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)
	}

	// HTTPSendJSON(w, msgHist[len(msgHist)-1].Parts, &JSONResponseOptions{})
}

func updateMessageHistory(w http.ResponseWriter, ctx context.Context, chatID string, messageHistory []llms.MessageContent, resp *llms.ContentResponse) ([]llms.MessageContent, error) {
	respchoice := resp.Choices[0]

	fmtResp := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{
				Text: respchoice.Content,
			},
		},
	}

	for _, tc := range respchoice.ToolCalls {
		fmtResp.Parts = append(fmtResp.Parts,
			llms.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      tc.FunctionCall.Name,
					Arguments: tc.FunctionCall.Arguments,
				},
			},
		)
	}

	/***************** STREAM LLM RESPONSE *****************/
	sendMessageStart(w)

	// Stream Text Content
	sendText(respchoice.Content, w)

	// Stream Tool call requests
	for _, tc := range respchoice.ToolCalls {
		sendToolCallRequest(tc.ID, tc.FunctionCall.Name, tc.FunctionCall.Arguments, w)
	}

	/***************** SAVE LLM MESSAGE TO DB *****************/
	tx, _ := DB.BeginTx(ctx, nil)
	defer tx.Rollback()
	qtx := Q.WithTx(tx)

	// Insert base message
	msgID, _ := gonanoid.New()
	if err := qtx.InsertMessage(ctx, db.InsertMessageParams{
		ID:         msgID,
		Role:       string(fmtResp.Role),
		Content:    sql.NullString{Valid: false},
		StopReason: sql.NullString{String: respchoice.StopReason, Valid: respchoice.StopReason != ""},
		ChatID:     chatID,
	}); err != nil {
		return nil, fmt.Errorf("Failed to insert message - %w", err)
	}

	// Insert text part
	AITextMsgPartID, err := qtx.InsertAIMessagePart(ctx, db.InsertAIMessagePartParams{
		Type:      "text",
		PartIndex: 0,
		MessageID: msgID,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to insert AI message part - %w", err)
	}
	if err := qtx.InsertTextPart(ctx, db.InsertTextPartParams{
		Text: sql.NullString{
			String: fmtResp.Parts[0].(llms.TextContent).Text,
			Valid:  fmtResp.Parts[0].(llms.TextContent).Text != "",
		},
		MessagePartID: AITextMsgPartID,
	}); err != nil {
		return nil, fmt.Errorf("Failed to insert text part - %w", err)
	}

	// Insert tool calls
	for i, tc := range respchoice.ToolCalls {
		AIToolMsgPartID, err := qtx.InsertAIMessagePart(ctx, db.InsertAIMessagePartParams{
			Type:      tc.Type, // Should be function in most cases
			PartIndex: int64(i + 1),
			MessageID: msgID,
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to insert AI message part - %w", err)
		}

		if err := qtx.InsertToolCallPart(ctx, db.InsertToolCallPartParams{
			ToolCallID:    tc.ID,
			Name:          tc.FunctionCall.Name,
			Arguments:     tc.FunctionCall.Arguments,
			MessagePartID: AIToolMsgPartID,
		}); err != nil {
			return nil, fmt.Errorf("Failed to insert tool call part - %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Failed to insert AI message into DB - %w", err)
	}

	// Return updated msg hist
	return append(messageHistory, fmtResp), nil
}

type ToolCallRequest struct {
	ToolUseID string         `json:"tool_use_id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCallResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"is_error"`
}

func execToolCalls(chatID string, msgHist []llms.MessageContent, resp *llms.ContentResponse, toolToAddr map[string]string) ([]llms.MessageContent, error) {
	fmt.Println("Executing", len(resp.Choices[0].ToolCalls), "tool calls")

	for _, tc := range resp.Choices[0].ToolCalls {
		addr := toolToAddr[tc.FunctionCall.Name]

		// PREPARE TC REQUEST
		var args map[string]any
		json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args)

		payload := ToolCallRequest{
			ToolUseID: tc.ID,
			Name: func() string {
				funcName := strings.Split(tc.FunctionCall.Name, "___")
				return funcName[1]
			}(),
			Arguments: args,
		}

		payloadJSONb, _ := json.Marshal(payload)

		ViewObjectAsJSON("REQUEST PAYLOAD", payload, nil)

		// SEND TC REQUEST
		res, err := http.Post(addr+"/callTool", "application/json", bytes.NewBuffer(payloadJSONb))
		if err != nil {
			return nil, fmt.Errorf("Failed to make request to /callTool - %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusUnauthorized {
			// Tool Response
			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    "Function could not be executed because user is unauthorized. User must re-authenticate to continue.",
					},
				},
			})

			ts := strings.Split(tc.FunctionCall.Name, "___")
			mcp := ts[0]

			// System message
			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart(
						fmt.Sprintf("Tools for the MCP instance `%s` have been disabled. Do not call these functions as they will not work. When the user has re-authenticated you will be notified and can use the tools.", mcp),
					),
				},
			})

			ctx := context.Background()
			tx, _ := DB.BeginTx(ctx, nil)
			defer tx.Rollback()
			qtx := Q.WithTx(tx)

			// Tool Response to DB
			tcMsgID, _ := gonanoid.New()
			qtx.InsertMessage(ctx, db.InsertMessageParams{
				ID:         tcMsgID,
				Role:       string(llms.ChatMessageTypeTool),
				Content:    sql.NullString{Valid: false},
				StopReason: sql.NullString{Valid: false},
				ChatID:     chatID,
			})
			qtx.InsertToolCallResult(ctx, db.InsertToolCallResultParams{
				MessageID:  tcMsgID,
				ToolCallID: tc.ID,
				Name:       tc.FunctionCall.Name,
				Content:    "Function could not be executed because user is unauthorized. User must re-authenticate to continue.",
				IsError:    true,
			})

			// System message to DB
			sysMsgID, _ := gonanoid.New()
			qtx.InsertMessage(ctx, db.InsertMessageParams{
				ID:   sysMsgID,
				Role: string(llms.ChatMessageTypeSystem),
				Content: sql.NullString{
					String: fmt.Sprintf("Tools for the MCP instance `%s` have been disabled. Do not call these functions as they will not work. When the user has re-authenticated you will be notified and can use the tools.", mcp),
					Valid:  true,
				},
				StopReason: sql.NullString{
					Valid: false,
				},
				ChatID: chatID,
			})

			if err := tx.Commit(); err != nil {
				return nil, err
			}

		} else {

			// IF RESPONSE HTTP TYPE is 401 then that means the user is needs to re-authenticate.
			/*
				Delete user Oauth tokens in DB
				Delete user MCP server instance
				Send request to user to re-authenticate (stream JSON object)
			*/

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("Failed to read request body - %w", err)
			}

			var tcRes ToolCallResult
			json.Unmarshal(body, &tcRes)

			ViewObjectAsJSON("TOOL CALL RESULT", tcRes, nil)

			// Add TC Res to local msg history
			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: func() []llms.ContentPart {
					parts := []llms.ContentPart{}

					for _, c := range tcRes.Content {
						parts = append(parts, llms.ToolCallResponse{
							ToolCallID: tcRes.ToolUseID,
							Name:       tc.FunctionCall.Name,
							Content:    c.Text,
						})
					}

					return parts
				}(),
			})

			// Add TC Res to DB
			ctx := context.Background()

			tx, err := DB.BeginTx(ctx, nil)
			defer tx.Rollback()
			qtx := Q.WithTx(tx)

			msgID, _ := gonanoid.New()

			if err := qtx.InsertMessage(ctx, db.InsertMessageParams{
				ID:         msgID,
				Role:       string(llms.ChatMessageTypeTool),
				Content:    sql.NullString{Valid: false},
				StopReason: sql.NullString{Valid: false},
				ChatID:     chatID,
			}); err != nil {
				return nil, fmt.Errorf("Failed to create message in DB - %w", err)
			}

			qtx.InsertToolCallResult(ctx, db.InsertToolCallResultParams{
				MessageID:  msgID,
				ToolCallID: tc.ID,
				Name:       tc.FunctionCall.Name,
				Content: func() string {
					contentJSONb, _ := json.Marshal(tcRes.Content)
					return string(contentJSONb)
				}(),
				IsError: tcRes.IsError,
			})

			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("Failed to insert tool call response into DB - %w", err)
			}
		}
	}

	return msgHist, nil
}

func (app *App) StartServer() error {
	app.Logger.Printf("🚀 Starting server on %s\n", app.Listener.Addr().String())

	return http.Serve(app.Listener, app.Mux)
}

func (app *App) PANIC(reason string) {
	e := errors.New(reason)
	app.ErrLogger.Print(e)
	panic(e)
}

func main() {
	a, err := NewApp()
	if err != nil {
		fmt.Print(err)
		a.PANIC(fmt.Errorf("Failed to create new app - %w", err).Error())
	}

	a.StartServer()
}

func sendMessageStart(w http.ResponseWriter) error {
	msgStartID, _ := gonanoid.New()
	msgStart := struct {
		Type      string `json:"type"`
		MessageID string `json:"messageId"`
	}{
		Type:      "start",
		MessageID: msgStartID,
	}

	msgStartJSONb, _ := json.Marshal(msgStart)

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(msgStartJSONb)); err != nil {
		return fmt.Errorf("Failed to write message start - %w", err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	return nil
}

func sendMessageEnd(w http.ResponseWriter) error {
	msgStart := struct {
		Type string `json:"type"`
	}{
		Type: "finish",
	}

	msgStartJSONb, _ := json.Marshal(msgStart)

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(msgStartJSONb)); err != nil {
		return fmt.Errorf("Failed to write message end - %w", err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	return nil
}

func sendStreamTerminate(w http.ResponseWriter) error {
	if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
		return fmt.Errorf("Failed to write message end - %w", err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	return nil
}

func sendText(content string, w http.ResponseWriter) error {

	textID, _ := gonanoid.New()

	/********** TEXT START **********/
	textStart := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "text-start",
		ID:   textID,
	}

	textStartJSONb, _ := json.Marshal(textStart)

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(textStartJSONb)); err != nil {
		return fmt.Errorf("Failed to write text start - %w", err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	/********** TEXT DELTA **********/

	textDelta := struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Delta string `json:"delta"`
	}{
		Type:  "text-delta",
		ID:    textID,
		Delta: content,
	}

	textDeltaJSONb, _ := json.Marshal(textDelta)

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(textDeltaJSONb)); err != nil {
		return fmt.Errorf("Failed to write text delta - %w", err)
	}

	flusher.Flush()

	/********** TEXT END **********/

	textEnd := struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}{
		Type: "text-end",
		ID:   textID,
	}

	textEndJSONb, _ := json.Marshal(textEnd)

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(textEndJSONb)); err != nil {
		return fmt.Errorf("Failed to write text end - %w", err)
	}

	flusher.Flush()

	return nil
}

func sendToolCallRequest(toolCallID string, toolName string, toolInput string, w http.ResponseWriter) error {
	/********** TOOL INPUT START **********/

	toolInputStart := struct {
		Type       string `json:"type"`
		ToolCallID string `json:"toolCallId"`
		ToolName   string `json:"toolName"`
	}{
		Type:       "tool-input-start",
		ToolCallID: toolCallID,
		ToolName:   toolName,
	}

	toolInputStartJSONb, _ := json.Marshal(toolInputStart)

	if _, err := fmt.Fprintf(w, "data: %s", string(toolInputStartJSONb)); err != nil {
		return fmt.Errorf("Failed to write tool input start %s - %w", toolCallID, err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	/********** TOOL INPUT AVAILABLE **********/

	toolInputAvailable := struct {
		Type       string `json:"type"`
		ToolCallID string `json:"toolCallId"`
		ToolName   string `json:"toolName"`
		Input      string `json:"input"`
	}{
		Type:       "tool-input-available",
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Input:      toolInput,
	}

	toolInputAvailableJSONb, _ := json.Marshal(toolInputAvailable)

	if _, err := fmt.Fprintf(w, "data: %s", string(toolInputAvailableJSONb)); err != nil {
		return fmt.Errorf("Failed to write tool input available %s - %w", toolCallID, err)
	}

	flusher.Flush()

	return nil
}

func sendToolCallResponse(toolCallID string, output map[string]any, w http.ResponseWriter) error {

	toolCallOutput := struct {
		Type       string         `json:"string"`
		ToolCallID string         `json:"toolCallId"`
		Output     map[string]any `json:"output"`
	}{
		Type:       "tool-output-available",
		ToolCallID: toolCallID,
		Output:     output,
	}

	toolCallOutputJSONb, _ := json.Marshal(toolCallOutput)

	if _, err := fmt.Fprintf(w, "data: %s", string(toolCallOutputJSONb)); err != nil {
		return fmt.Errorf("Failed to write tool output %s - %w", toolCallID, err)
	}

	flusher, _ := w.(http.Flusher)
	flusher.Flush()

	return nil
}

func (app *App) helloHandler(w http.ResponseWriter, r *http.Request) {
	//CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("x-vercel-ai-ui-message-stream", "v1")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to read request body - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}
	defer r.Body.Close()

	ViewObjectAsJSON("RAW MESSAGE", bodyBytes, nil)

	if err := sendMessageStart(w); err != nil {
		fmt.Printf("Failed to send message start - %v\n", err)
	}

	if err := sendText("ping pong", w); err != nil {
		fmt.Printf("Failed to send text - %v\n", err)
	}

	if err := sendMessageEnd(w); err != nil {
		fmt.Printf("Failed to send message end - %v\n", err)
	}

	if err := sendStreamTerminate(w); err != nil {
		fmt.Printf("Failed to send stream terminate - %v\n", err)
	}
}
