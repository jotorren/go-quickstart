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

### HTTP middleware

## Identity and access management

keycloak

## Troubleshooting

### Keycloak port forwarding

> `quickstart/deploy$ docker logs 1a653e537687`
> ```log
> Thu, 11 Jan 2024 00:43:44 UTC INF app/src/cmd/docker/main.go:46 > application.yaml read
> Thu, 11 Jan 2024 00:43:44 UTC INF app/src/infrastructure/config/configuration.go:50 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]} Security:{Oidc:{Configurl:http://127.0.0.1:8090/auth/realms/evote Clientid:golang-cli}}}
> Thu, 11 Jan 2024 00:43:44 UTC ERR go/pkg/mod/github.com/ipfans/fxlogger@v0.2.0/zerolog.go:72 > invoke failed error="could not build arguments for function \"main\".main.func3 (/app/src/cmd/docker/main.go:51):
> failed to build *http.Server: could not build arguments for function \"tsib/quickstart/infrastructure/transport\".NewHTTPServer (/app/src/infrastructure/transport/httpadapter.go:24):
> failed to build *mux.Router: could not build arguments for function \"tsib/quickstart/infrastructure/transport\".NewMuxRouter (/app/src/infrastructure/transport/httpadapter.go:67):
> failed to build *security.TokenVerifier: received non-nil error from function \"tsib/quickstart/infrastructure/security\".NewTokenVerifier (/app/src/infrastructure/security/oidc.go:75):
> Get \"http://127.0.0.1:8090/auth/realms/evote/.well-known/openid-configuration\": dial tcp 127.0.0.1:8090: connect: connection refused" function=main.main.func3() stack="main.main\n\t/app/src/cmd/docker/main.go:44\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:267\n"
> ```



> [!CAUTION]
> `Get "http://127.0.0.1:8090/auth/realms/evote/.well-known/openid-configuration": dial tcp 127.0.0.1:8090: connect: connection refused"`
>


## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
