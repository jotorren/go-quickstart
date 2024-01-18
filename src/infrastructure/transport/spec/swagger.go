// Package classification Quickstart
//
// Documentation of Quickstart public API.
//
//	 Schemes: http
//	 BasePath: /api/v1
//	 Version: 1.0.0
//	 Host: localhost:8080
//
//	 Consumes:
//	 - application/json
//
//	 Produces:
//	 - application/json
//
//	 Security:
//	 - OIDC
//
//	SecurityDefinitions:
//	  OIDC:
//	    type: oauth2
//	    flow: password
//	    tokenUrl: "http://127.0.0.1:8090/auth/realms/mycorp/protocol/openid-connect/token"
//
// swagger:meta
package spec

import "tsib/quickstart/infrastructure/transport"

// invalid security token
// swagger:response securityErrorResponse
type SecurityErrorResponse struct {
}

// service internal error
// swagger:response internalErrorResponse
type InternalErrorResponse struct {
}

// swagger:response stringResponse
// in: body
type _ string

// swagger:route GET /ping Ping ping
//
// Check whether or not the service is running.
//
//   security:
//     OIDC:
//
//   responses:
//     200: pingOkResponse
//     401: securityErrorResponse
//     500: internalErrorResponse

// the service is up and running
// swagger:response pingOkResponse
type PingOkResponse struct {
	// in: body
	Body transport.Result
}
