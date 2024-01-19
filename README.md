# Golang REST API Specification 
[![license](https://img.shields.io/badge/license-MIT-blue)](https://github.com/jotorren/go-quickstart/blob/feature/openapi-spec/LICENSE)
[![GoDoc](https://godoc.org/github.com/go-swagger/go-swagger?status.svg)](http://godoc.org/github.com/go-swagger/go-swagger)

This project shows how to generate an **OpenAPI 2.0 specification** (fka *Swagger 2.0 specification*) from annotated go code by using [`go-swagger`](https://goswagger.io/) 
      
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
>
> - **OpenId Connect** token security

## Changes to baseline code

### API metadata

New API specification

> `src/infrastructure/transport/spec/swagger.go`
> 
> ```go
> // Package classification Quickstart
> //
> // Documentation of Quickstart public API.
> //
> //	 Schemes: http
> //	 BasePath: /api/v1
> //	 Version: 1.0.0
> //	 Host: localhost:8080
> //
> //	 Consumes:
> //	 - application/json
> //
> //	 Produces:
> //	 - application/json
> //
> //	 Security:
> //	 - OIDC
> //
> //	SecurityDefinitions:
> //	  OIDC:
> //	    type: oauth2
> //	    flow: password
> //	    tokenUrl: "http://127.0.0.1:8090/auth/realms/mycorp/protocol/openid-connect/token"
> //
> // swagger:meta
> package spec
> 
> import "tsib/quickstart/infrastructure/transport"
> 
> // invalid security token
> // swagger:response securityErrorResponse
> type SecurityErrorResponse struct {
> }
> 
> // service internal error
> // swagger:response internalErrorResponse
> type InternalErrorResponse struct {
> }
> 
> // swagger:response stringResponse
> // in: body
> type _ string
> 
> // swagger:route GET /ping Ping ping
> //
> // Check whether or not the service is running.
> //
> //   security:
> //     OIDC:
> //
> //   responses:
> //     200: pingOkResponse
> //     401: securityErrorResponse
> //     500: internalErrorResponse
> 
> // the service is up and running
> // swagger:response pingOkResponse
> type PingOkResponse struct {
> 	// in: body
> 	Body transport.Result
> }
> ```

### API model

Annotations added

> `src/infrastructure/transport/restcontroller.go`
> ```diff
> +// swagger:model
> type Result struct {
> +	// the service result object
> +	//
> +	// Extensions:
> +	// ---
> +	// x-go-type: interface{}
> +	// ---
> 	Value interface{} `json:"result"`
> }
> 
> +// swagger:model
> type ApiError struct {
> +	// the error details
> 	Code int `json:"code"`
> +	// http error code
> 	Message string `json:"error"`
> }
> ```

## Generate Swagger doc

First, install `go-swagger` via go

```shell
$ go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.4
go: downloading github.com/go-swagger/go-swagger v0.30.4
go: downloading github.com/go-openapi/runtime v0.25.0
go: downloading github.com/go-openapi/spec v0.20.8
go: downloading github.com/go-openapi/strfmt v0.21.3
go: downloading github.com/go-openapi/swag v0.22.3
go: downloading github.com/go-openapi/validate v0.22.0
go: downloading github.com/spf13/viper v1.14.0
go: downloading github.com/go-openapi/errors v0.20.3
go: downloading golang.org/x/tools v0.5.0
go: downloading golang.org/x/sys v0.4.0
go: downloading github.com/go-openapi/jsonpointer v0.19.5
go: downloading github.com/google/uuid v1.1.2
go: downloading github.com/spf13/cast v1.5.0
go: downloading github.com/rogpeppe/go-internal v1.9.0
go: downloading github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
go: downloading go.mongodb.org/mongo-driver v1.11.1
go: downloading github.com/go-openapi/jsonreference v0.20.0
go: downloading github.com/spf13/afero v1.9.2
go: downloading github.com/subosito/gotenv v1.4.1
go: downloading github.com/pelletier/go-toml/v2 v2.0.5
go: downloading golang.org/x/text v0.6.0
```

Then, run the following commands from the project's root directory:

```shell
quickstart$ cd src 
quickstart/src$ swagger generate spec -o ./resources/swagger.json --scan-models
quickstart/src$ ls -l ./resources/swagger.json 
-rwxrwxrwx 1 jotorren jotorren 2298 Jan 18 16:56 ./resources/swagger.json
```

## Publish Swagger doc

> `src/Makefile`
> ```diff
> all: test build
> 
> build:
> 	@echo '**********' Building binary...
> 	go build -o myapp cmd/local/main.go
> 	@echo
> 
> test:
> 	@echo '**********' Running tests...
> 	go test -v ./...
> 	@echo
> 
>+ gen-swagger:
>+ 	swagger generate spec -o ./resources/swagger.json --scan-models
>+ 
>+ swagger: gen-swagger
>+ 	swagger serve --port=8081 -F=swagger resources/swagger.json
> ```

> `src/resources.go`
> ```diff
> package src
> 
> import _ "embed"
> 
> //go:embed resources/application.yaml
> var ApplicationYaml []byte
> 
>+ //go:embed resources/swagger.json
>+ var SwaggerJson []byte
> ```

> `src/infrastructure/transport/httpadapter.go`
> ```diff
> package transport
> 
> import (
> 	"context"
> 	"net"
> 	"net/http"
> 	"time"
> 
>+ 	src "tsib/quickstart"
> 	"tsib/quickstart/infrastructure/config"
> 	"tsib/quickstart/infrastructure/security"
> 
> 	"github.com/google/uuid"
> 	"github.com/gorilla/mux"
> 	"github.com/rs/cors"
> 	"github.com/rs/zerolog"
> 	"go.uber.org/fx"
> )
> 
> ...
> 
> func NewMuxRouter(p MuxRouterParams) *mux.Router {
> 	ml, ok := p.Cfg.Log.PackagesLevel[HTTP_PACKAGE_NAME]
> 	if ok {
> 		p.Logger = p.Logger.Level(zerolog.Level(ml))
> 	}
> 
> 	router := mux.NewRouter()
> 	router.Use(loggerMiddleware(p.Logger))
>+ 	router.Handle("/swagger.json", byteHandler(src.SwaggerJson, "application/json"))
> 
> 	api := router.PathPrefix("/api/v1").Subrouter()
> 	if nil != p.Verifier {
> 		api.Use(security.NewOAuth2Middleware(p.Verifier))
> 	} else {
> 		api.Use(security.NewOAuth2AnonymousMiddleware())
> 	}
> 
> 	api.HandleFunc("/ping", p.Controller.Ping).Methods("GET")
> 
> 	return router
> }
> 
> ...
> 
>+ func byteHandler(b []byte, contentType string) http.HandlerFunc {
>+ 	return func(w http.ResponseWriter, _ *http.Request) {
>+ 		w.Header().Set("Content-Type", contentType)
>+ 		w.Write(b)
>+ 	}
>+ }
> ```


## Swagger UI

To visualize and interact with the API’s resources:

```shell
quickstart/src$ swagger serve --port=8081 -F=swagger resources/swagger.json
2024/01/18 16:58:47 serving docs at http://localhost:8081/docs
```

From this point on, you can access the swagger UI at http://localhost:8081/docs

## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
