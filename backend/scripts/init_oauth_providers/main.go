package main

import (
	"fmt"

	oauth_provider "github.com/AbhinavPalacharla/xtrn-personal/internal/oauth-provider"
)

func main() {
	oauth_provider.GoogleCalendarOauthProvider.StoreOauthProvider()

	oauth_provider.GmailOauthProvider.StoreOauthProvider()

	oauth_provider.GoogleSigninOauthProvider.StoreOauthProvider()

	fmt.Print("✅ Created Google OAuth Provider\n")
}
