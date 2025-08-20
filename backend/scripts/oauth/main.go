package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	oauth_provider "github.com/AbhinavPalacharla/xtrn-personal/internal/oauth-provider"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var googleSignin = oauth_provider.GoogleSigninOauthProvider
var googleCalendar = oauth_provider.GoogleCalendarOauthProvider
var gmail = oauth_provider.GmailOauthProvider

type Server struct {
	sessionStore *sessions.CookieStore
}

func ConfigureGoth(s *Server) error {

	goth.UseProviders(
		googleSignin.GothProvider,
		googleCalendar.GothProvider,
		gmail.GothProvider,
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

func storeOauthToken(providerName string, refreshToken string) error {
	_, err := Q.GetOauthTokenByProvider(context.Background(), providerName)
	if err == sql.ErrNoRows {
		//Token does not exist so insert
		id, _ := gonanoid.New()
		err := Q.InsertOauthToken(context.Background(), db.InsertOauthTokenParams{
			ID:            id,
			RefreshToken:  refreshToken,
			OauthProvider: providerName,
		})

		if err != nil {
			return err
		}
	}

	//Token does exist so update it
	err = Q.UpdateOauthTokenByProivder(context.Background(), db.UpdateOauthTokenByProivderParams{
		RefreshToken:  refreshToken,
		OauthProvider: providerName,
	})

	return err
}

func handleCallbackHandler(providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("provider", providerName)
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

		storeOauthToken(providerName, user.RefreshToken)
	}
}

func beginAuthHandler(providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Set("provider", providerName)
		r.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(w, r)
	}
}

func createOauthHandlers(providerName string) {
	route := "/auth/" + providerName
	callbackRoute := "/auth/" + providerName + "/callback"

	http.HandleFunc(route, beginAuthHandler(providerName))
	http.HandleFunc(callbackRoute, handleCallbackHandler(providerName))
}

func main() {
	s, err := NewServer()
	_ = s

	if err != nil {
		StdErrLogger.Panicf("%v", err)
	}

	fmt.Println(googleSignin.Name, googleCalendar.Name, gmail.Name)

	// Routes
	createOauthHandlers(googleSignin.Name)
	createOauthHandlers(googleCalendar.Name)
	createOauthHandlers(gmail.Name)
	// http.HandleFunc("/auth/google-signin", beginAuthHandler("google-signin"))
	// http.HandleFunc("/auth/google-calendar", beginAuthHandler("google-calendar"))
	// http.HandleFunc("/auth/google-signin/callback", handleCallbackHandler("google-signin"))
	// http.HandleFunc("/auth/google-calendar/callback", handleCallbackHandler("google-calendar"))

	fmt.Println("Open one of these in your browser:")
	fmt.Println("  http://localhost:8080/auth/google-signin")
	fmt.Println("  http://localhost:8080/auth/google-calendar")
	fmt.Println("  http://localhost:8080/auth/gmail")

	StdErrLogger.Fatal(http.ListenAndServe(":8080", nil))
}
