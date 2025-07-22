package oauth_provider

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const GOOGLE_SIGNIN_PROVIDER_NAME = "google-signin"
const GOOGLE_SIGNIN_CALLBACK_URL = "http://localhost:8080/auth/" + GOOGLE_SIGNIN_PROVIDER_NAME + "/callback"

var GOOGLE_SIGNIN_SCOPES = []string{"email", "profile", "https://www.googleapis.com/auth/calendar"}

var GoogleSigninOauthProvider = newGoogleSigninOauthProvider()

func newGoogleSigninOauthProvider() *types.OauthProvider {
	clientID, err := shared.GetEnv("GOOGLE_CLIENT_ID")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_ID` env variable")
	}
	clientSecret, err := shared.GetEnv("GOOGLE_CLIENT_SECRET")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_SECRET` env variable")
	}

	p := types.NewOauthProvider(
		GOOGLE_SIGNIN_PROVIDER_NAME,
		clientID,
		clientSecret,
		GOOGLE_SIGNIN_CALLBACK_URL,
		GOOGLE_SIGNIN_SCOPES,
		func() goth.Provider {
			googleSignin := google.New(
				clientID,
				clientSecret,
				GOOGLE_SIGNIN_CALLBACK_URL,
				GOOGLE_SIGNIN_SCOPES...,
			)

			googleSignin.SetAccessType("offline")
			googleSignin.SetPrompt("consent")
			googleSignin.SetName(GOOGLE_SIGNIN_PROVIDER_NAME)

			return googleSignin
		},
	)

	return p
}
