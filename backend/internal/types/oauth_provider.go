package types

import (
	"context"
	"database/sql"
	"encoding/json"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
	"github.com/markbates/goth"
)

type OauthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	Scopes       []string
	GothProvider goth.Provider
}

func (p *OauthProvider) StoreOauthProvider() {

	DB.InsertOauthProvider(context.Background(), db.InsertOauthProviderParams{
		Name:         p.Name,
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		CallbackUrl:  p.CallbackURL,
		Scopes: sql.NullString{
			String: func() string {
				jsonBytes, _ := json.Marshal(p.Scopes)

				return string(jsonBytes)
			}(),
			Valid: len(p.Scopes) > 0,
		},
	})
}

func NewOauthProvider(
	name string,
	clientID string,
	clientSecret string,
	callbackURL string,
	scopes []string,
	newGothProvider func() goth.Provider,
) *OauthProvider {
	p := OauthProvider{
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
		Scopes:       scopes,
		GothProvider: newGothProvider(),
	}

	return &p
}
