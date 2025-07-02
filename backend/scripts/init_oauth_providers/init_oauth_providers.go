package main

import (
	// oauth_provider "github.com/AbhinavPalacharla/xtrn-personal/internal/oauth-provider"

	"fmt"

	oauth_provider "github.com/AbhinavPalacharla/xtrn-personal/internal/oauth-provider"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

func main() {
	googleOauthProvider, err := oauth_provider.NewGoogleOauthProvider()

	if err != nil {
		ErrorLogger.Fatal(err)
	} else {
		fmt.Print("✅ Created Google OAuth Provider\n")
	}

	_ = googleOauthProvider

	// err := DB.InsertOauthProvider(context.Background(), db.InsertOauthProviderParams{
	// 	Name:         "google",
	// 	ClientID:     "someclientid",
	// 	ClientSecret: "someclientsecret",
	// })

	// if err != nil {
	// 	fmt.Print(err)
	// }
}
