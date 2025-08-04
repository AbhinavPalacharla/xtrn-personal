package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

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
	a.Mux.HandleFunc("/chat", a.handleChat)
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

func (app *App) handleChat(w http.ResponseWriter, r *http.Request) {
	//CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	//XTRN frontend headers
	w.Header().Set("Access-Control-Expose-Headers", "x-xtrn-chat-id")

	chatID := r.PathValue("chatID")
	if chatID == "" {
		// If not chatID then this is the first message - need to init the chat and possibly configure tools specially
		id, _ := gonanoid.New()
		chatID = id
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

	llm, err := openai.New(
		openai.WithToken(""),
		openai.WithModel("gpt-4.1-mini"),
	)
	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to create LLM instance - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	msgHist := []llms.MessageContent{
		// llms.TextParts(llms.ChatMessageTypeHuman, "What is on my calendar this week. For the week of 7/27/2025?"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Create an event on my calendar called dinner for this friday at 7pm. For the week of 7/27/2025. I am trying to test out tool calling so the first time you call the create calendar tool do not include a timzeone. On the second call, make the timezone PST. Never ask to proceed for tool calls, just make the tool call request.."),
		// llms.TextParts(llms.ChatMessageTypeHuman, "Create a calendar event for next week on Friday called dinner"),
	}

	fmt.Printf("Checking calendar...")

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
	msgHist = updateMessageHistory(msgHist, resp)

	ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)

	for {
		// Check if there are any tool calls to execute and execute them
		if len(resp.Choices[0].ToolCalls) == 0 {
			break // No tools needed so end conversation loop and wait for user to send next message
		}
		msgHist = execToolCalls(msgHist, resp, toolToAddr)

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
		msgHist = updateMessageHistory(msgHist, resp)

		ViewObjectAsJSON("MSG HIST LLM", msgHist, nil)
	}

	HTTPSendJSON(w, msg, &JSONResponseOptions{})
}

func updateMessageHistory(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	respchoice := resp.Choices[0]

	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	return append(messageHistory, assistantResponse)
}

func execToolCalls(msgHist []llms.MessageContent, resp *llms.ContentResponse, toolToAddr map[string]string) []llms.MessageContent {
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

		// args := bytes.NewBufferString(toolCall.FunctionCall.Arguments)
		// fmt.Printf("ARGS: %s\n", args.String())

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

		resBody := map[string]any{}
		json.Unmarshal(body, &resBody)

		contentB, _ := json.Marshal(resBody)

		fmt.Printf("RAW RESPONSE BODY: %s\n", string(contentB))

		msgHist = append(msgHist, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: toolCall.ID,
					Name:       toolCall.FunctionCall.Name,
					Content:    string(contentB),
				},
			},
		})
	}

	return msgHist
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
