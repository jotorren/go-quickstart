# Secure REST API in Golang with OpenID Connect

This project focuses on how to secure REST APIs in Golang using the OIDC/OAuth2 standard assuming an Identity Provider (IDP) is already available.
      
> [!NOTE]
> Inherited *key aspects* from baseline code:
>
> - **Hexagonal architecture**
> 
> - **uber-go/fx** dependency injection framework
> 
> - **uber-go/config** injection-friendly YAML configuration
> 
> - **net/http** with **CORS** security
> 
> - **gorilla/mux** router and a **subrouter** to use specific middleware for specific routes
> 
> - **rs/zerolog** logging library. **Per request contextual logging**  (all traces within the same request will share the same unique id)

## Changes

### Security configuration

Edit `src/resources/application.yaml` to add the following lines:

>```diff
> log:
>   defaultlevel: ${LOG_LEVEL:1}
>   packageslevel:
>     http: ${LOG_LEVEL_HTTP:1}
> 
> server:
>   port: ${SERVER_PORT:8080}
>   origins: ${SERVER_ALLOWED_ORIGINS}
> 
>+ security:
>+   oidc:
>+     configurl: ${OIDC_SERVER}
>+     clientid: 'golang-cli'
>```

Where:
- **security.oidc.configurl** is the token issuer realm URL and,
- **security.oidc.clientid** is the public identifier for the application
  
### OIDC token verifier
Token validation algorithm implementation

>`src/infrastructure/security/oidc.go`
>```go
>...
> func NewTokenVerifier(cfg *config.Configuration, tr *TokenVerifierTransport) (*TokenVerifier, error) {
> 	verifier := TokenVerifier{}
> 
> 	verifier.ctx = oidc.ClientContext(context.Background(), &http.Client{
> 		Timeout:   time.Duration(5000) * time.Second,
> 		Transport: tr,
> 	})
> 
> 	provider, err := oidc.NewProvider(verifier.ctx, cfg.Security.Oidc.Configurl)
> 	if err != nil {
> 		return nil, err
> 	}
> 
> 	verifier.instance = provider.Verifier(&oidc.Config{
> 		ClientID: cfg.Security.Oidc.Clientid,
> 	})
> 
> 	return &verifier, nil
> }
> 
> func (v *TokenVerifier) Parse(rawToken string) (*ParseableToken, error) {
> 	idToken, err := v.instance.Verify(v.ctx, rawToken)
> 	if err != nil {
> 		return nil, err
> 	}
> 
> 	var parsedToken ParseableToken
> 	if err := idToken.Claims(&parsedToken); err != nil {
> 		return nil, err
> 	}
> 
> 	return &parsedToken, nil
> }
>```

> [!IMPORTANT]
> **The token issuer claim (iss) must match the value passed as issuer (cfg.Security.Oidc.Configurl) when instantiating the verifier**

The current implementation only supports tokens that comply with:
> ```go
> type ParseableToken struct {
> 	Audience                  []string `json:"aud,omitempty"`
> 	ExpiresAt                 int64    `json:"exp,omitempty"`
> 	Id                        string   `json:"jti,omitempty"`
> 	IssuedAt                  int64    `json:"iat,omitempty"`
> 	Issuer                    string   `json:"iss,omitempty"`
> 	NotBefore                 int64    `json:"nbf,omitempty"`
> 	Subject                   string   `json:"sub,omitempty"`
> 	Type                      string   `json:"typ,omitempty"`
> 	AuthorizedParty           string   `json:"azp,omitempty"`
> 	AuthContextClassReference string   `json:"acr,omitempty"`
> 
> 	RealmAccess struct {
> 		Roles []string `json:"roles,omitempty"`
> 	} `json:"realm_access,omitempty"`
> 
> 	ResourceAccess struct {
> 		GolangCli struct {
> 			Roles []string `json:"roles,omitempty"`
> 		} `json:"golang-cli,omitempty"`
> 		Account struct {
> 			Roles []string `json:"roles,omitempty"`
> 		} `json:"account,omitempty"`
> 	} `json:"resource_access,omitempty"`
> 
> 	Scope             string `json:"scope,omitempty"`
> 	Name              string `json:"name,omitempty"`
> 	PreferredUsername string `json:"preferred_username,omitempty"`
> 	GivenName         string `json:"given_name,omitempty"`
> 	FamilyName        string `json:"family_name,omitempty"`
> 	Email             string `json:"email,omitempty"`
> }
> ```

Finally you must register the verifier constructor function in **fx**. To do so add these lines to the file `src/cmd/docker/main.go`:

> ```diff
> func main() {
> 	fx.New(
> 		fx.Provide(
> 			NewRootLogger,
> 			config.NewConfiguration,
> 			transport.NewHTTPServer,
> 			transport.NewMuxRouter,
> 			transport.NewRestController,
>+ 			security.NewTokenVerifier,
> 		),
> 		fx.WithLogger(func(rootLogger zerolog.Logger) fxevent.Logger {
> 			return fxlogger.WithZerolog(rootLogger.Level(zerolog.WarnLevel))()
> 		}),
>+ 		fx.Supply(
>+ 			&security.TokenVerifierTransport{
>+ 				Transport: http.Transport{
>+ 					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
>+ 				},
>+ 			},
>+ 		),
> 		fx.Invoke(
> 			func(cfg *config.Configuration, rootLogger zerolog.Logger) {
> 				rootLogger.Info().Msg("application.yaml read")
> 				cfg.Print(rootLogger)
> 			},
> 			func(*http.Server) {
> 				// start server
> 			},
> 		),
> 	).Run()
> }
> ```

> [!IMPORTANT]
> The second green block (fx.Supply...) is only required when accessing the IDP using the http protocol instead of https

### HTTP middleware

New convenience methods to easily validate http requests

> `src/infrastructure/security/http.go`
> ```go
> func NewOAuth2Middleware(verifier *TokenVerifier) mux.MiddlewareFunc {
> 	return func(next http.Handler) http.Handler {
> 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
> 			logger := zerolog.Ctx(r.Context()).With().Str("func", OIDC_HTTP_MIDDLEWARE_NAME).Logger()
> 
> 			auth := strings.TrimSpace(r.Header.Get(headers.Authorization))
> 			if auth == "" {
> 				logger.Error().Msg(MSG_OIDC_TOKEN_NOT_PRESENT)
> 				authorizationFailed(MSG_OIDC_TOKEN_NOT_PRESENT, http.StatusUnauthorized, w, r)
> 				return
> 			}
> 			if len(auth) < 7 {
> 				logger.Error().Msg(MSG_OIDC_TOKEN_NOT_PRESENT)
> 				authorizationFailed(MSG_OIDC_TOKEN_NOT_PRESENT, http.StatusUnauthorized, w, r)
> 				return
> 			}
> 			auth = auth[len(ACCESS_TOKEN_TYPE+" "):]
> 			logger.Debug().Str("raw", auth).Send()
> 
> 			token, err := verifier.Parse(auth)
> 			if err != nil {
> 				logger.Error().Err(err).Msg(MSG_OIDC_TOKEN_PARSE)
> 				authorizationFailed(MSG_OIDC_TOKEN_PARSE, http.StatusUnauthorized, w, r)
> 				return
> 			}
> 			logger.Debug().Any("token", token).Send()
> 
> 			ctx := context.WithValue(r.Context(), CONTEXT_PRINCIPAL_KEY, token.PreferredUsername)
> 			ctx = context.WithValue(ctx, CONTEXT_ROLES_KEY, token.ResourceAccess.GolangCli.Roles)
> 
> 			next.ServeHTTP(w, r.WithContext(ctx))
> 		})
> 	}
> }
> 
> func NewOAuth2AnonymousMiddleware() mux.MiddlewareFunc {
> 	return func(next http.Handler) http.Handler {
> 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
> 
> 			ctx := context.WithValue(r.Context(), CONTEXT_PRINCIPAL_KEY, OIDC_NOSECURITY_USER)
> 			ctx = context.WithValue(ctx, CONTEXT_ROLES_KEY, []string{})
> 
> 			next.ServeHTTP(w, r.WithContext(ctx))
> 		})
> 	}
> }
> ```

Add the verifier dependency to `src/infrastructure/transport/httpadapter.go`:
> ```diff
> type MuxRouterParams struct {
> 	fx.In
> 
> 	Cfg        *config.Configuration
>+ 	Verifier   *security.TokenVerifier `optional:"true"`
> 	Controller *RestController
> 	Logger     zerolog.Logger
> }
> 
> func NewMuxRouter(p MuxRouterParams) *mux.Router {
> 	ml, ok := p.Cfg.Log.PackagesLevel[HTTP_PACKAGE_NAME]
> 	if ok {
> 		p.Logger = p.Logger.Level(zerolog.Level(ml))
> 	}
> 
> 	router := mux.NewRouter()
> 	router.Use(loggerMiddleware(p.Logger))
> 
> 	api := router.PathPrefix("/api/v1").Subrouter()
>+ 	if nil != p.Verifier {
>+ 		api.Use(security.NewOAuth2Middleware(p.Verifier))
>+ 	} else {
>+ 		api.Use(security.NewOAuth2AnonymousMiddleware())
>+ 	}
> 
> 	api.HandleFunc("/ping", p.Controller.Ping).Methods("GET")
> 
> 	return router
> }
> ```

> [!NOTE]
> The verifier is instantiated directly without using **fx**, because it is an optional object and its required properties will not always be included in the application configuration.
> Note that security is only applied at the mux sub-router level, leaving the main router intact.

## Identity and access management

The `keycloak/docker-compose.yaml` file contains:
> ```yaml
> version: '3'
>     
> services:
>   postgres:
>       image: postgres
>       volumes:
>         - myapp_kcdb:/var/lib/postgresql/data
>       environment:
>         POSTGRES_DB: keycloak
>         POSTGRES_USER: keycloak
>         POSTGRES_PASSWORD: r0CaToLive
>   keycloak:
>       image: quay.io/keycloak/keycloak:legacy
>       environment:
>         DB_VENDOR: POSTGRES
>         DB_ADDR: postgres
>         DB_DATABASE: keycloak
>         DB_USER: keycloak
>         DB_SCHEMA: public
>         DB_PASSWORD: r0CaToLive
>         KEYCLOAK_USER: admin
>         KEYCLOAK_PASSWORD: Pa55w0rd
>       ports:
>         - 8090:8080
>       depends_on:
>         - postgres
> 
> volumes:
>   myapp_kcdb:
>     driver: local
> ```

To start a **keycloak** local instance with a **postgres** database run the following command:
```bash
go-quickstart$ cd keycloak/
go-quickstart/keycloak$ docker-compose up -d
Creating network "keycloak_default" with the default driver
Creating volume "keycloak_myapp_kcdb" with local driver
Creating keycloak_postgres_1 ... done
Creating keycloak_keycloak_1 ... done

go-quickstart/keycloak$ docker ps
CONTAINER ID   IMAGE                              COMMAND                  CREATED         STATUS         PORTS                              NAMES
c02b7ef9a48e   quay.io/keycloak/keycloak:legacy   "/opt/jboss/tools/do…"   3 seconds ago   Up 2 seconds   8443/tcp, 0.0.0.0:8090->8080/tcp   keycloak_keycloak_1
ca77b8562e35   postgres                           "docker-entrypoint.s…"   4 seconds ago   Up 2 seconds   5432/tcp                           keycloak_postgres_1
```

Keycloak console UI will be accessible at http://localhost:8090/auth/ 

Once logged-in:
1. Create realm with `name` = `mycorp` (and default values)
2. Create public client (`authorization field set to off`) with `name` = `golang-cli` (and default values)
3. Create client roles (tab `Roles` - `Create role`): admin, player
4. Create a user filling `username`, `email`, `email verified (true)`, `first name`, `last name` 
5. Set user credentials password (`temporary off`)
6. Assign to the user one or more roles from the list created in the 3rd step (tab `Role mapping` - `Assign role`)

If everything went well, you will be able to request a token:

```bash
curl -X POST --location 'http://127.0.0.1:8090/auth/realms/mycorp/protocol/openid-connect/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'client_id=golang-cli' \
--data-urlencode 'username=jotorren' \
--data-urlencode 'password=password' \
--data-urlencode 'grant_type=password'
```

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIyTVdOMDJKMms0bThIT2RTdGEtaVdqRlY1dWN2UFBDTnhFcmlncnlrT0V3In0.eyJleHAiOjE3MDUwNzI1MDMsImlhdCI6MTcwNTA3MjIwMywianRpIjoiYjRjOWZhYTgtYWEzYy00ODIxLWE0NTEtNWZlMDY2YmFmOWQzIiwiaXNzIjoiaHR0cDovLzEyNy4wLjAuMTo4MDkwL2F1dGgvcmVhbG1zL215Y29ycCIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiIxNmIxYjFmYS04MTRlLTQzMjktYTc4NS02ZWIxYTFlN2RjYzEiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJnb2xhbmctY2xpIiwic2Vzc2lvbl9zdGF0ZSI6ImZiYzQ5NDZiLTMzNzQtNDU2MS05N2Q1LWI5YzZkNzE0OWFkYiIsImFjciI6IjEiLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsib2ZmbGluZV9hY2Nlc3MiLCJ1bWFfYXV0aG9yaXphdGlvbiIsImRlZmF1bHQtcm9sZXMtbXlhcHAiXX0sInJlc291cmNlX2FjY2VzcyI6eyJnb2xhbmctY2xpIjp7InJvbGVzIjpbImFkbWluIiwicGxheWVyIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6InByb2ZpbGUgZW1haWwiLCJzaWQiOiJmYmM0OTQ2Yi0zMzc0LTQ1NjEtOTdkNS1iOWM2ZDcxNDlhZGIiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkpvcmRpIFRvcnJlbnRlIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiam90b3JyZW4iLCJnaXZlbl9uYW1lIjoiSm9yZGkiLCJmYW1pbHlfbmFtZSI6IlRvcnJlbnRlIiwiZW1haWwiOiJqb3RvcnJlbkBtYWlsLmNvbSJ9.Kd4P90GJEsoUpnfrgMPMeRsJqlA0OLzMfRopAbiLG_cCRLQH-KiOidKkGJ81RsH1ezDjQPYuD2IPdg2OEa_HLUQ28pBnUwVoz0LN_9xsybNviSGI5m_-BYstGQkmOe_Ko2I52YFsI6Q8nvtU7-XjZYmS5nSdG1a5xtE_fJa-i3HgXvX9jdBob-yJzlYTz7NtohnYl0hUNGN8zshuzS2cFJMsJPp0LOszjpWhqJ5PCYw7kYcBGKD3GuR0gDTRPDASk01_ZXTASwstoYuNQ86dJPbPp2qn4BqVu2vYDpRCbVVa3HJPSth64zdJSM2RLMtrZ1NMNg95o_dFTdrJ4Uvnbg",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI0ZTY3YmUzNS04ZWMxLTQ1N2QtOGY2Zi03OTc1ZTgwM2MxNWEifQ.eyJleHAiOjE3MDUwNzQwMDMsImlhdCI6MTcwNTA3MjIwMywianRpIjoiYzJmOWNhYzEtOThjOS00NmQwLTg5ZjAtMTljYTQwMGEyZWI5IiwiaXNzIjoiaHR0cDovLzEyNy4wLjAuMTo4MDkwL2F1dGgvcmVhbG1zL215Y29ycCIsImF1ZCI6Imh0dHA6Ly8xMjcuMC4wLjE6ODA5MC9hdXRoL3JlYWxtcy9teWNvcnAiLCJzdWIiOiIxNmIxYjFmYS04MTRlLTQzMjktYTc4NS02ZWIxYTFlN2RjYzEiLCJ0eXAiOiJSZWZyZXNoIiwiYXpwIjoiZ29sYW5nLWNsaSIsInNlc3Npb25fc3RhdGUiOiJmYmM0OTQ2Yi0zMzc0LTQ1NjEtOTdkNS1iOWM2ZDcxNDlhZGIiLCJzY29wZSI6InByb2ZpbGUgZW1haWwiLCJzaWQiOiJmYmM0OTQ2Yi0zMzc0LTQ1NjEtOTdkNS1iOWM2ZDcxNDlhZGIifQ.L75M3RCiedik7rsvHRBVDcKrJmuzzbuWpxThewvlKkk",
  "token_type": "Bearer",
  "not-before-policy": 0,
  "session_state": "fbc4946b-3374-4561-97d5-b9c6d7149adb",
  "scope": "profile email"
}
```

## Troubleshooting

### Keycloak port forwarding

If you compile the source code shown so far and run the executable binary from the command line, everything should work as expected. But, if you run the application inside a docker container, you will most likely receive the following error:

> `quickstart/deploy$ docker logs 1a653e537687`
> ```log
> Thu, 11 Jan 2024 00:43:44 UTC INF app/src/cmd/docker/main.go:46 > application.yaml read
> Thu, 11 Jan 2024 00:43:44 UTC INF app/src/infrastructure/config/configuration.go:50 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]} Security:{Oidc:{Configurl:http://127.0.0.1:8090/auth/realms/mycorp Clientid:golang-cli}}}
> Thu, 11 Jan 2024 00:43:44 UTC ERR go/pkg/mod/github.com/ipfans/fxlogger@v0.2.0/zerolog.go:72 > invoke failed error="could not build arguments for function \"main\".main.func3 (/app/src/cmd/docker/main.go:51):
> failed to build *http.Server: could not build arguments for function \"tsib/quickstart/infrastructure/transport\".NewHTTPServer (/app/src/infrastructure/transport/httpadapter.go:24):
> failed to build *mux.Router: could not build arguments for function \"tsib/quickstart/infrastructure/transport\".NewMuxRouter (/app/src/infrastructure/transport/httpadapter.go:67):
> failed to build *security.TokenVerifier: received non-nil error from function \"tsib/quickstart/infrastructure/security\".NewTokenVerifier (/app/src/infrastructure/security/oidc.go:75):
> Get \"http://127.0.0.1:8090/auth/realms/mycorp/.well-known/openid-configuration\": dial tcp 127.0.0.1:8090: connect: connection refused" function=main.main.func3() stack="main.main\n\t/app/src/cmd/docker/main.go:44\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:267\n"
> ```

More concretely:

> [!CAUTION]
> `Get "http://127.0.0.1:8090/auth/realms/mycorp/.well-known/openid-configuration": dial tcp 127.0.0.1:8090: connect: connection refused"`
>



## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
