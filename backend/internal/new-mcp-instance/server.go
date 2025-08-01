package new_mcp_instance

import (
	"net"
	"net/http"
)

type HTTPServer struct {
	app *App
}

func NewHTTPServer(app *App) *HTTPServer {
	s := HTTPServer{
		app: app,
	}

	return &s
}

func (s *HTTPServer) StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/listTools", s.handleListTools)
	mux.HandleFunc("/callTool", s.handleCallTool)
	mux.HandleFunc("/kill", s.handleKill)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s.app.Listener = listener
	s.app.Address = listener.Addr().String()

	//Write listen address back to creator
	s.app.Logger.Print("TRANSMITTING SERVER ADDRESS\n")
	conn, err := net.Dial("unix", s.app.CallbackAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Write([]byte(s.app.Address))

	//Run Server
	s.app.Logger.Printf("🚀 SERVER RUNNING ON %s\n", s.app.Address)

	return http.Serve(listener, mux)
}
