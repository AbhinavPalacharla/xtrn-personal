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
	a.Mux.HandleFunc("/chats", a.handleChat)
	a.Mux.HandleFunc("/chats/{chatID}/messages", a.handleChat)
	a.Mux.HandleFunc("/messages", a.handleGetMessage)

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

func tmp(chatID string) {
	rows, err := Q.GetViewChatMessges(context.Background(), chatID)
	// rows, err := Q.GetChatMessages(context.Background(), chatID)
	if err != nil {
		panic(err)
	}

	for _, r := range rows {
		_ = r
	}
}

func getMessageHistory(chatID string) ([]llms.MessageContent, error) {
	messageRows, err := Q.GetMessages(context.Background(), chatID)
	if err != nil {
		return nil, err
	}

	msgHist := []llms.MessageContent{}

	// llmMsgIdx := 0

	var aiMsg *llms.MessageContent = nil

	for _, m := range messageRows {
		if m.TreqName.Valid {
			// Save tool call req

			if aiMsg != nil {
				aiMsg.Parts = append(aiMsg.Parts, llms.ToolCall{
					ID:   m.TreqToolCallID.String,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      m.TreqName.String,
						Arguments: m.TreqArgs.String,
					},
				})
			} else {
				panic("NO AI MESSAGE TO ADD TOOL CALL REQ TO")
			}

			// msgHist = append(msgHist, llms.MessageContent{
			// 	Role: llms.ChatMessageTypeTool,
			// 	Parts: []llms.ContentPart{
			// 		llms.ToolCall{
			// 			ID:   m.TreqToolCallID.String,
			// 			Type: "function",
			// 			FunctionCall: &llms.FunctionCall{
			// 				Name:      m.TreqName.String,
			// 				Arguments: m.TreqArgs.String,
			// 			},
			// 		},
			// 	},
			// })
			// msgHist[llmMsgIdx].Parts = append(msgHist[llmMsgIdx].Parts, llms.ToolCall{
			// 	ID:   m.TreqToolCallID.String,
			// 	Type: "function",
			// 	FunctionCall: &llms.FunctionCall{
			// 		Name:      m.TreqName.String,
			// 		Arguments: m.TreqArgs.String,
			// 	},
			// })
		} else if m.TresName.Valid {
			// First add ai msg to msgHist
			msgHist = append(msgHist, *aiMsg)

			// Save tool call res
			msgHist = append(msgHist, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: m.TresID.String,
						Name:       m.TresName.String,
						Content:    m.TresContent.String,
					},
				},
			})
		} else {
			if m.Role == "ai" {
				aiMsg = &llms.MessageContent{
					Role: llms.ChatMessageType(m.Role),
					Parts: []llms.ContentPart{
						llms.TextContent{
							Text: m.Content.String,
						},
					},
				}

				ViewObjectAsJSON("AI MESSAGE", *aiMsg, nil)
			} else {
				msgHist = append(msgHist, llms.TextParts(llms.ChatMessageType(m.Role), m.Content.String))
			}
		}
	}

	// If remaining ai msg add it to msg hist

	if aiMsg != nil {
		msgHist = append(msgHist, *aiMsg)
	}

	return msgHist, nil
}

func (app *App) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chatID")

	if chatID == "" {
		HTTPReturnError(w, ErrorOptions{
			Err:  "Missing chatID query param",
			Code: http.StatusBadRequest,
		})
	}

	msgHist, _ := getMessageHistory(chatID)

	// ViewObjectAsJSON("MESSAGE HISTORY", messageRows, nil)
	ViewObjectAsJSON("LANGCHAIN MESSSAGE HISTORY", msgHist, nil)

	HTTPSendJSON(w, msgHist, nil)
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
			Err: fmt.Errorf("Failed to parse request body - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	app.Logger.Printf("MESSAGE RECIEVED: %s\n", msg.Content)

	tx, _ := DB.BeginTx(context.Background(), nil)
	qtx := Q.WithTx(tx)

	if newChat {
		qtx.InsertChat(context.Background(), chatID)
	}

	msgID, _ := gonanoid.New()

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
		tx.Rollback()
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to insert message - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
		return
	}

	// Fetch MCP instance tools
	instanceRows, err := Q.GetMCPServerInstances(context.Background())

	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to get MCP instances - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
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

	var msgHist []llms.MessageContent

	if !newChat {
		msgHist, err = getMessageHistory(chatID)
		if err != nil {
			HTTPReturnError(w, ErrorOptions{
				Err:  fmt.Errorf("Failed to get messages for chatID=%s - %w", chatID, err).Error(),
				Code: http.StatusInternalServerError,
			})
		}
	}

	// msgHist = append(msgHist, llms.TextParts(llms.ChatMessageTypeHuman, msg.Content))

	// msgHist := []llms.MessageContent{
	// 	// llms.TextParts(llms.ChatMessageTypeHuman, "What is on my calendar this week. For the week of 7/27/2025?"),
	// 	// llms.TextParts(llms.ChatMessageTypeHuman, "Create an event on my calendar called dinner for this friday at 7pm. For the week of 7/27/2025. I am trying to test out tool calling so the first time you call the create calendar tool do not include a timzeone. On the second call, make the timezone PST. Never ask to proceed for tool calls, just make the tool call request.."),
	// 	// llms.TextParts(llms.ChatMessageTypeHuman, "Create a calendar event for next week on Friday called dinner"),
	// 	llms.TextParts(llms.ChatMessageTypeHuman, msg.Content),
	// }

	ViewObjectAsJSON("\nMESSAGE HISTORY\n\n", msgHist, nil)

	fmt.Printf("Generating LLM Response...")

	// Get initial response from LLM
	resp, err := llm.GenerateContent(context.Background(), msgHist, llms.WithTools(tools))
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to get response from LLM - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	fmt.Println("LLM RESPONSE SUCCEEDED")

	// Add LLM response to msg history
	msgHist = updateMessageHistory(chatID, msgHist, resp)

	ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)

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
		msgHist = updateMessageHistory(chatID, msgHist, resp)

		ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)
	}

	HTTPSendJSON(w, msgHist[len(msgHist)-1].Parts, &JSONResponseOptions{})
}

func updateMessageHistory(chatID string, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	respchoice := resp.Choices[0]

	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}

	// Save msg to DB
	msgID, _ := gonanoid.New()
	Q.InsertMessage(context.Background(), db.InsertMessageParams{
		ID:   msgID,
		Role: string(llms.ChatMessageTypeAI),
		Content: sql.NullString{
			String: respchoice.Content,
			Valid:  respchoice.Content != "",
		},
		StopReason: sql.NullString{
			String: respchoice.StopReason,
			Valid:  respchoice.StopReason != "",
		},
		ChatID: chatID,
	})

	return append(messageHistory, assistantResponse)
}

func execToolCalls(chatID string, msgHist []llms.MessageContent, resp *llms.ContentResponse, toolToAddr map[string]string) ([]llms.MessageContent, error) {
	fmt.Println("Executing", len(resp.Choices[0].ToolCalls), "tool calls")

	for _, toolCall := range resp.Choices[0].ToolCalls {

		fmt.Printf("TOOL CALL: %s\nTOOL ARGS:%s\n", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

		addr := toolToAddr[toolCall.FunctionCall.Name]

		fmtArgs := map[string]any{}
		json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &fmtArgs)

		toolName := strings.Split(toolCall.FunctionCall.Name, "___")

		payload := struct {
			ToolUseID string         `json:"tool_use_id"`
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}{
			ToolUseID: toolCall.ID,
			Name:      toolName[1],
			Arguments: fmtArgs,
		}

		payloadJsonB, _ := json.Marshal(payload)

		fmt.Printf("REQUEST PAYLOAD: %s\n", string(payloadJsonB))

		toolReqTX, _ := DB.BeginTx(context.Background(), nil)
		toolReqQTX := Q.WithTx(toolReqTX)

		toolReqMsgID, _ := gonanoid.New()
		toolReqQTX.InsertMessage(context.Background(), db.InsertMessageParams{
			ID:   toolReqMsgID,
			Role: string(llms.ChatMessageTypeTool),
			Content: sql.NullString{
				String: "",
				Valid:  false,
			},
			StopReason: sql.NullString{
				String: "",
				Valid:  false,
			},
			ChatID: chatID,
		})
		toolReqQTX.InsertToolCallRequest(context.Background(), db.InsertToolCallRequestParams{
			MessageID:  toolReqMsgID,
			ToolCallID: toolCall.ID,
			Name:       toolName[1],
			Arguments:  toolCall.FunctionCall.Arguments,
		})

		if err := toolReqTX.Commit(); err != nil {
			toolReqTX.Rollback()
			return nil, fmt.Errorf("Failed to save tool call request to DB - %w", err)
		}

		res, err := http.Post(addr+"/callTool", "application/json", bytes.NewBuffer(payloadJsonB))
		if err != nil {
			log.Panic(fmt.Errorf("Failed to call tool - %w\n", err))
		}
		defer res.Body.Close()

		fmt.Printf("RESPONSE RECIEVED\n")

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Panic(fmt.Errorf("Failed to read request body - %w\n", err))
		}

		// resBody := map[string]any{}
		resBody := struct {
			ToolUseID string `json:"tool_use_id"`
			Content   []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}
			IsError bool `json:"is_error"`
		}{}

		json.Unmarshal(body, &resBody)

		contentB, _ := json.Marshal(resBody)

		fmt.Printf("RAW RESPONSE BODY: %s\n", string(contentB))

		toolResTX, _ := DB.BeginTx(context.Background(), nil)
		toolResQTX := Q.WithTx(toolResTX)

		toolResMsgID, _ := gonanoid.New()
		ee := toolResQTX.InsertMessage(context.Background(), db.InsertMessageParams{
			ID:   toolResMsgID,
			Role: string(llms.ChatMessageTypeTool),
			Content: sql.NullString{
				String: "",
				Valid:  false,
			},
			StopReason: sql.NullString{
				String: "",
				Valid:  false,
			},
			ChatID: chatID,
		})

		if ee != nil {
			fmt.Printf("Insert message res failed - %v\n", ee)
		}

		e := toolResQTX.InsertToolCallResult(context.Background(), db.InsertToolCallResultParams{
			MessageID:  toolResMsgID,
			ToolCallID: toolCall.ID,
			Name:       toolCall.FunctionCall.Name,
			Content:    resBody.Content[0].Text,
			IsError:    resBody.IsError,
		})

		if e != nil {
			fmt.Printf("Insert res failed - %v\n", e)
		}

		if err := toolResTX.Commit(); err != nil {
			toolResTX.Rollback()
			return nil, fmt.Errorf("Failed to save tool call response to DB - %w", err)
		}

		fmtContent := string(contentB)

		if resBody.IsError {
			fmtContent = fmt.Sprintf("ERROR: %s", fmtContent)
		}

		msgHist = append(msgHist, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: toolCall.ID,
					Name:       toolCall.FunctionCall.Name,
					Content:    fmtContent,
				},
			},
		})
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
