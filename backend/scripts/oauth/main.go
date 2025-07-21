package main

import (
	"fmt"
	"net/http"
	"os"

	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Server struct {
	sessionStore *sessions.CookieStore
}

type Provider struct {
	name         string
	clientID     string
	clientSecret string
	callbackURL  string
	scopes       []string
	NewProvider  func(name string, clientID string, clientSecret string, callbackURL string, scopes ...string) *goth.Provider
}

func ConfigureGoth(s *Server) error {
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if googleClientID == "" {
		return fmt.Errorf("`GOOGLE_CLIENT_ID` env variable not defined")
	}

	if googleClientSecret == "" {
		return fmt.Errorf("`GOOGLE_CLIENT_SECRET` env variable not defined")
	}

	googleSignIn := google.New(
		googleClientID,
		googleClientSecret,
		"http://localhost:8080/auth/google-signin/callback",
		"email", "profile",
	)
	googleSignIn.SetAccessType("offline")
	googleSignIn.SetPrompt("consent")
	googleSignIn.SetName("google-signin")

	googleCalendar := google.New(
		googleClientID,
		googleClientSecret,
		"http://localhost:8080/auth/google-calendar/callback",
		"email", "profile", "https://www.googleapis.com/auth/calendar",
	)
	googleCalendar.SetAccessType("offline")
	googleCalendar.SetPrompt("consent")
	googleCalendar.SetName("google-calendar")

	goth.UseProviders(
		googleSignIn,
		googleCalendar,
	)

	gothic.Store = s.sessionStore

	return nil
}

func NewServer() (*Server, error) {
	s := Server{}

	sessionSecret, err := GetEnv("SESSION_SECRET")
	if err != nil {
		return nil, err
	}

	s.sessionStore = sessions.NewCookieStore([]byte(sessionSecret))
	s.sessionStore.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   false,
	}

	err = ConfigureGoth(&s)
	if err != nil {
		return nil, fmt.Errorf("Failed to configure GOTH - %w", err)
	}

	return &s, nil
}

func handleCallbackHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("provider", provider)
		r.URL.RawQuery = q.Encode()

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			HTTPReturnError(w, ErrorOptions{
				Err:  err.Error(),
				Code: http.StatusBadRequest,
			})
			StdErrLogger.Printf("OAuth Failed - %v", err)
			return
		}

		fmt.Fprintf(w, "Login successful. You can close this tab.\n")

		fmt.Print("\n==========================================================\n")
		fmt.Printf("Provider: %s\n", user.Provider)
		fmt.Printf("Email:    %s\n", user.Email)
		fmt.Printf("Access Token: %s\n", user.AccessToken)
		fmt.Printf("Refresh Token: %s\n", user.RefreshToken)
		fmt.Print("==========================================================\n")
	}
}

func beginAuthHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(w, r)
	}
}

func main() {
	s, err := NewServer()
	_ = s

	if err != nil {
		StdErrLogger.Panicf("%v", err)
	}

	// Routes
	http.HandleFunc("/auth/google-signin", beginAuthHandler("google-signin"))
	http.HandleFunc("/auth/google-calendar", beginAuthHandler("google-calendar"))
	http.HandleFunc("/auth/google-signin/callback", handleCallbackHandler("google-signin"))
	http.HandleFunc("/auth/google-calendar/callback", handleCallbackHandler("google-calendar"))

	fmt.Println("Open one of these in your browser:")
	fmt.Println("  http://localhost:8080/auth/google-signin")
	fmt.Println("  http://localhost:8080/auth/google-calendar")

	StdErrLogger.Fatal(http.ListenAndServe(":8080", nil))
}
