package oauth_provider

import (
	"github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const GOOGLE_CALENDAR_PROVIDER_NAME = "google-calendar"
const GOOGLE_CALENDAR_CALLBACK_URL = "http://localhost:8080/auth/" + GOOGLE_CALENDAR_PROVIDER_NAME + "/callback"

var SCOPES = []string{"email", "profile", "https://www.googleapis.com/auth/calendar"}

var GoogleCalendarOauthProvider = newGoogleCalendarOauthProvider()

func newGoogleCalendarOauthProvider() *types.OauthProvider {
	clientID, err := shared.GetEnv("GOOGLE_CLIENT_ID")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_ID` env variable")
	}
	clientSecret, err := shared.GetEnv("GOOGLE_CLIENT_SECRET")
	if err != nil {
		shared.StdErrLogger.Panic("Could not read `GOOGLE_CLIENT_SECRET` env variable")
	}

	p := types.NewOauthProvider(
		GOOGLE_CALENDAR_PROVIDER_NAME,
		clientID,
		clientSecret,
		GOOGLE_CALENDAR_CALLBACK_URL,
		SCOPES,
		func() goth.Provider {
			googleCalendar := google.New(
				clientID,
				clientSecret,
				GOOGLE_CALENDAR_CALLBACK_URL,
				SCOPES...,
			)

			googleCalendar.SetAccessType("offline")
			googleCalendar.SetPrompt("consent")
			googleCalendar.SetName(GOOGLE_CALENDAR_PROVIDER_NAME)

			return googleCalendar
		},
	)

	return p
}
