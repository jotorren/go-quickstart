package security

import (
	"context"
	"net/http"
	"time"

	"tsib/quickstart/infrastructure/config"

	"github.com/coreos/go-oidc"
)

type key string

const (
	OIDC_PACKAGE_NAME = "oidc"

	CONTEXT_PRINCIPAL_KEY key = "oidc_subject"
	CONTEXT_ROLES_KEY     key = "oidc_roles"

	MSG_OIDC_TOKEN_NOT_PRESENT = "token not present"
	MSG_OIDC_TOKEN_PROVIDER    = "unable to configure the token provider"
	MSG_OIDC_TOKEN_PARSE       = "token not valid"

	ACCESS_TOKEN_TYPE  string = "Bearer"
	REFRESH_TOKEN_TYPE string = "Refresh"

	OIDC_NOSECURITY_USER = "anonymous"
)

type ParseableToken struct {
	//jwt.StandardClaims: redefine aud as an array

	Audience                  []string `json:"aud,omitempty"`
	ExpiresAt                 int64    `json:"exp,omitempty"`
	Id                        string   `json:"jti,omitempty"`
	IssuedAt                  int64    `json:"iat,omitempty"`
	Issuer                    string   `json:"iss,omitempty"`
	NotBefore                 int64    `json:"nbf,omitempty"`
	Subject                   string   `json:"sub,omitempty"`
	Type                      string   `json:"typ,omitempty"`
	AuthorizedParty           string   `json:"azp,omitempty"`
	AuthContextClassReference string   `json:"acr,omitempty"`

	RealmAccess struct {
		Roles []string `json:"roles,omitempty"`
	} `json:"realm_access,omitempty"`

	ResourceAccess struct {
		GolangCli struct {
			Roles []string `json:"roles,omitempty"`
		} `json:"golang-cli,omitempty"`
		Account struct {
			Roles []string `json:"roles,omitempty"`
		} `json:"account,omitempty"`
	} `json:"resource_access,omitempty"`

	Scope             string `json:"scope,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	Email             string `json:"email,omitempty"`
}

type TokenVerifierTransport struct {
	http.Transport
}

type TokenVerifier struct {
	ctx      context.Context
	instance *oidc.IDTokenVerifier
}

func NewTokenVerifier(cfg *config.Configuration, tr *TokenVerifierTransport) (*TokenVerifier, error) {
	verifier := TokenVerifier{}

	verifier.ctx = oidc.ClientContext(context.Background(), &http.Client{
		Timeout:   time.Duration(5000) * time.Second,
		Transport: tr,
	})

	provider, err := oidc.NewProvider(verifier.ctx, cfg.Security.Oidc.Configurl)
	if err != nil {
		return nil, err
	}

	verifier.instance = provider.Verifier(&oidc.Config{
		ClientID: cfg.Security.Oidc.Clientid,
	})

	return &verifier, nil
}

func (v *TokenVerifier) Parse(rawToken string) (*ParseableToken, error) {
	idToken, err := v.instance.Verify(v.ctx, rawToken)
	if err != nil {
		return nil, err
	}

	var parsedToken ParseableToken
	if err := idToken.Claims(&parsedToken); err != nil {
		return nil, err
	}

	return &parsedToken, nil
}
