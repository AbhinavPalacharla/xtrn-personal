package main

import (
	"fmt"

	oauth_provider "github.com/AbhinavPalacharla/xtrn-personal/internal/oauth-provider"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	googleOauthProvider, err := oauth_provider.NewGoogleOauthProvider()

	if err != nil {
		StdErrLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created Google OAuth Provider\n")
	}

	_ = googleOauthProvider
}
