# Secure REST API in Golang with OpenID Connect

This project focuses on how to secure (authentication and authorization) REST APIs in Golang by using the OIDC/OAuth2 standard. 
      
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

YAML

### OIDC token verifier

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
> `Get \"http://127.0.0.1:8090/auth/realms/evote/.well-known/openid-configuration\": dial tcp 127.0.0.1:8090: connect: connection refused"`
>


## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
