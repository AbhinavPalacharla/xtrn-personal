package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	gonanoid "github.com/matoous/go-nanoid/v2"
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
	instances, err := Q.GetMCPServerInstances(context.Background())

	if err != nil {
		HTTPReturnError(w, ErrorOptions{
			Err: fmt.Errorf("Failed to get MCP instances - %w", err).Error(),
		})
		app.ErrLogger.Print(err)
	}

	fmt.Printf("MCP INSTANCES\n%+v\n", instances)

	// for _, inst := range instances {
	// for _, t := range inst.Tools {

	// }
	// }

	//Load tools from instances into LLM

	// for _, t := range instances
	// for _, inst := range instances {
	// 	inst.
	// }

	HTTPSendJSON(w, msg, &JSONResponseOptions{})
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
