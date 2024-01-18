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

### Swagger model

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

### API metadata

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

## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
