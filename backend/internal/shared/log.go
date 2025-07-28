package shared

import (
	"fmt"
	"log"
	"os"
	"path"
)

var StdErrLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

type MCPInstanceLoggers struct {
	Logger    *log.Logger
	ErrLogger *log.Logger
}

func NewMCPInstanceLogger(instanceID string) MCPInstanceLoggers {
	logDir := os.Getenv("LOG_DIR")

	if logDir == "" {
		panic("Environment variable `LOG_DIR` not set. Please run `eval $(make setup-env)`")
	}

	logFile, err := os.OpenFile(path.Join(logDir, instanceID+".log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		panic(err)
	}

	l := log.New(logFile, fmt.Sprintf("%s: ", instanceID), log.Llongfile)
	el := log.New(logFile, fmt.Sprintf("ERROR: %s: ", instanceID), log.Llongfile)

	return MCPInstanceLoggers{
		Logger:    l,
		ErrLogger: el,
	}
}

type APILoggers struct {
	Logger    *log.Logger
	ErrLogger *log.Logger
}

func NewAPILoggers() APILoggers {
	l := log.New(os.Stdout, fmt.Sprintf("API"), log.Llongfile)
	el := log.New(os.Stderr, fmt.Sprintf("ERROR: API: "), log.Llongfile)

	return APILoggers{
		Logger:    l,
		ErrLogger: el,
	}
}
