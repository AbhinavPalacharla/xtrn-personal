package oauth_provider

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const GMAIL_PROVIDER_NAME = "gmail"
const GMAIL_CALLBACK_URL = "http://localhost:8080/auth/" + GMAIL_PROVIDER_NAME + "/callback"

var GMAIL_SCOPES = []string{"email", "profile", "https://www.googleapis.com/auth/gmail.readonly", "https://www.googleapis.com/auth/gmail.send"}

var GmailOauthProvider = newGmailOauthProvider()

func newGmailOauthProvider() *types.OauthProvider {
	clientID, err := shared.GetEnv("GOOGLE_CLIENT_ID")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_ID` env variable")
	}
	clientSecret, err := shared.GetEnv("GOOGLE_CLIENT_SECRET")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_SECRET` env variable")
	}

	p := types.NewOauthProvider(
		GMAIL_PROVIDER_NAME,
		clientID,
		clientSecret,
		GMAIL_CALLBACK_URL,
		GMAIL_SCOPES,
		func() goth.Provider {
			gmailProvider := google.New(
				clientID,
				clientSecret,
				GMAIL_CALLBACK_URL,
				GMAIL_SCOPES...,
			)

			gmailProvider.SetAccessType("offline")
			gmailProvider.SetPrompt("consent")
			gmailProvider.SetName(GMAIL_PROVIDER_NAME)

			return gmailProvider
		},
	)

	return p
}