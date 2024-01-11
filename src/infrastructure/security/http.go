package security

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	OIDC_HTTP_MIDDLEWARE_NAME = "http.oidc"
)

type Res40XStruct struct {
	Status   string `json:"status" example:"FAILED"`
	HTTPCode int    `json:"httpCode" example:"401"`
	Message  string `json:"message" example:"authorization failed"`
}

func authorizationFailed(message string, httpCode int, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpCode)
	data := Res40XStruct{
		Status:   "FAILED",
		HTTPCode: httpCode,
		Message:  message,
	}
	res, _ := json.Marshal(data)
	w.Write(res)
}

func NewOAuth2Middleware(verifier *TokenVerifier) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := zerolog.Ctx(r.Context()).With().Str("func", OIDC_HTTP_MIDDLEWARE_NAME).Logger()

			auth := strings.TrimSpace(r.Header.Get(headers.Authorization))
			if auth == "" {
				logger.Error().Msg(MSG_OIDC_TOKEN_NOT_PRESENT)
				authorizationFailed(MSG_OIDC_TOKEN_NOT_PRESENT, http.StatusUnauthorized, w, r)
				return
			}
			if len(auth) < 7 {
				logger.Error().Msg(MSG_OIDC_TOKEN_NOT_PRESENT)
				authorizationFailed(MSG_OIDC_TOKEN_NOT_PRESENT, http.StatusUnauthorized, w, r)
				return
			}
			auth = auth[len(ACCESS_TOKEN_TYPE+" "):]
			logger.Debug().Str("raw", auth).Send()

			token, err := verifier.Parse(auth)
			if err != nil {
				logger.Error().Err(err).Msg(MSG_OIDC_TOKEN_PARSE)
				authorizationFailed(MSG_OIDC_TOKEN_PARSE, http.StatusUnauthorized, w, r)
				return
			}
			logger.Debug().Any("token", token).Send()

			ctx := context.WithValue(r.Context(), CONTEXT_PRINCIPAL_KEY, token.PreferredUsername)
			ctx = context.WithValue(ctx, CONTEXT_ROLES_KEY, token.ResourceAccess.GolangCli.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewOAuth2AnonymousMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := context.WithValue(r.Context(), CONTEXT_PRINCIPAL_KEY, OIDC_NOSECURITY_USER)
			ctx = context.WithValue(ctx, CONTEXT_ROLES_KEY, []string{})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
