package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorOptions struct {
	Err  string
	Code int
}

// Default returns HTTP Code 500 (Internal Server Error)
func HTTPReturnError(w http.ResponseWriter, opts ErrorOptions) error {
	w.Header().Set("Content-Type", "application/json")

	if opts.Code != 0 {
		w.WriteHeader(opts.Code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"error": opts.Err}); err != nil {
		return err
	}

	return nil
}

type JSONResponseOptions struct {
	StatusCode int
	Headers    map[string]string
}

func HTTPSendJSON[T any](w http.ResponseWriter, payload T, opts *JSONResponseOptions) error {
	status := http.StatusOK
	if opts != nil && opts.StatusCode != 0 {
		status = opts.StatusCode
	}

	w.Header().Set("Content-Type", "application/json")
	if opts != nil {
		for k, v := range opts.Headers {
			w.Header().Set(k, v)
		}
	}

	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func DecodeJSONBody[T any](r *http.Request, w http.ResponseWriter) (*T, error) {
	var payload T

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&payload); err != nil {
		HTTPReturnError(w, (ErrorOptions{
			Err:  "Malformed request body",
			Code: http.StatusBadRequest,
		}))

		return nil, err
	}

	return &payload, nil
}

func ViewObjectAsJSON(prefix string, obj any, printFn func(fmtString string, a ...any)) {
	b, err := json.MarshalIndent(obj, "", "   ")

	var fmtStr string

	if err != nil {
		fmtStr = fmt.Sprintf("\n\n%s: ERROR - Failed to marshal: %v\n\n", prefix, err)
	} else {
		fmtStr = fmt.Sprintf("\n\n%s: %s\n\n", prefix, string(b))
	}

	if printFn != nil {
		printFn(fmtStr)
	} else {
		fmt.Print(fmtStr)
	}
}
