package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

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
	a.Mux.HandleFunc("/chats", a.handleChat)
	a.Mux.HandleFunc("/chats/{chatID}/messages", a.handleChat)

	// a.Mux.HandleFunc("/chat", a.handleMessage) //Eventually needs to handle /chat/[chatID]

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

	// Get chat history
	var msgHist []llms.MessageContent

	msgHist = append(msgHist, llms.TextParts(llms.ChatMessageTypeHuman, msg.Content))

	// if !newChat {
	// msgHist, err = getMessageHistory(chatID)
	// if err != nil {
	// 	HTTPReturnError(w, ErrorOptions{
	// 		Err:  fmt.Errorf("Failed to get messages for chatID=%s - %w", chatID, err).Error(),
	// 		Code: http.StatusInternalServerError,
	// 	})
	// }
	// }

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
	msgHist, err = updateMessageHistory(context.Background(), chatID, msgHist, resp)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to update message history - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
		return
	}

	ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)

	// Execute tool calls
	// for {
	// 	// Check if there are any tool calls to execute and execute them
	// 	if len(resp.Choices[0].ToolCalls) == 0 {
	// 		break // No tools needed so end conversation loop and wait for user to send next message
	// 	}

	// 	msgHist, err = execToolCalls(chatID, msgHist, resp, toolToAddr)
	// 	if err != nil {
	// 		HTTPReturnError(w, ErrorOptions{
	// 			Err: fmt.Errorf("Failed to save execute tool call - %w", err).Error(),
	// 		})
	// 		app.ErrLogger.Print(err)
	// 		return
	// 	}

	// 	ViewObjectAsJSON("MSG HIST TOOL", msgHist, nil)

	// 	// Call LLM again with tool call responses
	// 	resp, err = llm.GenerateContent(context.Background(), msgHist, llms.WithTools(tools))
	// 	if err != nil {
	// 		HTTPReturnError(w, ErrorOptions{
	// 			Err: fmt.Errorf("Failed to get response from LLM - %w", err).Error(),
	// 		})
	// 		app.ErrLogger.Print(err)
	// 		return
	// 	}

	// 	// Add LLM response to msg history
	// 	msgHist, err = updateMessageHistory(context.Background(), chatID, msgHist, resp)
	// 	if err != nil {
	// 		HTTPReturnError(w, ErrorOptions{
	// 			Err: fmt.Errorf("Failed to update message history - %w", err).Error(),
	// 		})
	// 		app.ErrLogger.Print(err)
	// 		return
	// 	}

	// 	ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)
	// }

	HTTPSendJSON(w, msgHist[len(msgHist)-1].Parts, &JSONResponseOptions{})
}

func updateMessageHistory(ctx context.Context, chatID string, messageHistory []llms.MessageContent, resp *llms.ContentResponse) ([]llms.MessageContent, error) {
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

// func execToolCalls(chatID string, msgHist []llms.MessageContent, resp *llms.ContentResponse, toolToAddr map[string]string) ([]llms.MessageContent, error) {
// 	fmt.Println("Executing", len(resp.Choices[0].ToolCalls), "tool calls")

// 	for _, toolCall := range resp.Choices[0].ToolCalls {

// 		fmt.Printf("TOOL CALL: %s\nTOOL ARGS:%s\n", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

// 		addr := toolToAddr[toolCall.FunctionCall.Name]

// 		fmtArgs := map[string]any{}
// 		json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &fmtArgs)

// 		toolName := strings.Split(toolCall.FunctionCall.Name, "___")

// 		payload := struct {
// 			ToolUseID string         `json:"tool_use_id"`
// 			Name      string         `json:"name"`
// 			Arguments map[string]any `json:"arguments"`
// 		}{
// 			ToolUseID: toolCall.ID,
// 			Name:      toolName[1],
// 			Arguments: fmtArgs,
// 		}

// 		payloadJsonB, _ := json.Marshal(payload)

// 		fmt.Printf("REQUEST PAYLOAD: %s\n", string(payloadJsonB))

// 		toolReqTX, _ := DB.BeginTx(context.Background(), nil)
// 		toolReqQTX := Q.WithTx(toolReqTX)

// 		toolReqMsgID, _ := gonanoid.New()
// 		toolReqQTX.InsertMessage(context.Background(), db.InsertMessageParams{
// 			ID:   toolReqMsgID,
// 			Role: string(llms.ChatMessageTypeTool),
// 			Content: sql.NullString{
// 				String: "",
// 				Valid:  false,
// 			},
// 			StopReason: sql.NullString{
// 				String: "",
// 				Valid:  false,
// 			},
// 			ChatID: chatID,
// 		})
// 		toolReqQTX.InsertToolCallRequest(context.Background(), db.InsertToolCallRequestParams{
// 			MessageID:  toolReqMsgID,
// 			ToolCallID: toolCall.ID,
// 			Name:       toolName[1],
// 			Arguments:  toolCall.FunctionCall.Arguments,
// 		})

// 		if err := toolReqTX.Commit(); err != nil {
// 			toolReqTX.Rollback()
// 			return nil, fmt.Errorf("Failed to save tool call request to DB - %w", err)
// 		}

// 		res, err := http.Post(addr+"/callTool", "application/json", bytes.NewBuffer(payloadJsonB))
// 		if err != nil {
// 			log.Panic(fmt.Errorf("Failed to call tool - %w\n", err))
// 		}
// 		defer res.Body.Close()

// 		fmt.Printf("RESPONSE RECIEVED\n")

// 		body, err := io.ReadAll(res.Body)
// 		if err != nil {
// 			log.Panic(fmt.Errorf("Failed to read request body - %w\n", err))
// 		}

// 		// resBody := map[string]any{}
// 		resBody := struct {
// 			ToolUseID string `json:"tool_use_id"`
// 			Content   []struct {
// 				Type string `json:"type"`
// 				Text string `json:"text"`
// 			}
// 			IsError bool `json:"is_error"`
// 		}{}

// 		json.Unmarshal(body, &resBody)

// 		contentB, _ := json.Marshal(resBody)

// 		fmt.Printf("RAW RESPONSE BODY: %s\n", string(contentB))

// 		toolResTX, _ := DB.BeginTx(context.Background(), nil)
// 		toolResQTX := Q.WithTx(toolResTX)

// 		toolResMsgID, _ := gonanoid.New()
// 		ee := toolResQTX.InsertMessage(context.Background(), db.InsertMessageParams{
// 			ID:   toolResMsgID,
// 			Role: string(llms.ChatMessageTypeTool),
// 			Content: sql.NullString{
// 				String: "",
// 				Valid:  false,
// 			},
// 			StopReason: sql.NullString{
// 				String: "",
// 				Valid:  false,
// 			},
// 			ChatID: chatID,
// 		})

// 		if ee != nil {
// 			fmt.Printf("Insert message res failed - %v\n", ee)
// 		}

// 		e := toolResQTX.InsertToolCallResult(context.Background(), db.InsertToolCallResultParams{
// 			MessageID:  toolResMsgID,
// 			ToolCallID: toolCall.ID,
// 			Name:       toolCall.FunctionCall.Name,
// 			Content:    resBody.Content[0].Text,
// 			IsError:    resBody.IsError,
// 		})

// 		if e != nil {
// 			fmt.Printf("Insert res failed - %v\n", e)
// 		}

// 		if err := toolResTX.Commit(); err != nil {
// 			toolResTX.Rollback()
// 			return nil, fmt.Errorf("Failed to save tool call response to DB - %w", err)
// 		}

// 		fmtContent := string(contentB)

// 		if resBody.IsError {
// 			fmtContent = fmt.Sprintf("ERROR: %s", fmtContent)
// 		}

// 		msgHist = append(msgHist, llms.MessageContent{
// 			Role: llms.ChatMessageTypeTool,
// 			Parts: []llms.ContentPart{
// 				llms.ToolCallResponse{
// 					ToolCallID: toolCall.ID,
// 					Name:       toolCall.FunctionCall.Name,
// 					Content:    fmtContent,
// 				},
// 			},
// 		})
// 	}

// 	return msgHist, nil
// }

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
