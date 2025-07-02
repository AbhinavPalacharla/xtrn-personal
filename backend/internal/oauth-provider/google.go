package oauth_provider

import (
	"errors"
	"os"

	"github.com/AbhinavPalacharla/xtrn-personal/internal/types"
)

func NewGoogleOauthProvider() (*types.OauthProvider, error) {
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientId == "" {
		return nil, errors.New("Could not read `GOOGLE_CLIENT_ID` env variable")
	}

	if clientSecret == "" {
		return nil, errors.New("Could not read `GOOGLE_CLIENT_SECRET` env variable")
	}

	p, err := types.NewOauthProvider("google", clientId, clientSecret)

	if err != nil {
		return nil, err
	}

	return p, nil
}
