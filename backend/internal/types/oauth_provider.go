package types

import (
	"context"
	"fmt"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	. "github.com/AbhinavPalacharla/xtrn-personal/internal/shared"
)

type OauthProvider struct {
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func saveOauthProviderToDB(prov *OauthProvider) error {
	err := DB.InsertOauthProvider(context.TODO(), db.InsertOauthProviderParams{
		Name:         prov.Name,
		ClientID:     prov.ClientID,
		ClientSecret: prov.ClientSecret,
	})

	if err != nil {
		return err
	}

	return nil
}

func NewOauthProvider(name string, clientId string, clientSecret string) (*OauthProvider, error) {
	p := OauthProvider{
		Name:         name,
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}

	if err := saveOauthProviderToDB(&p); err != nil {
		return nil, fmt.Errorf("Failed to save provider to DB - %w\n", err)
	}

	return &p, nil
}
