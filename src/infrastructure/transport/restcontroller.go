package transport

import (
	"encoding/json"
	"net/http"

	"tsib/quickstart/infrastructure/config"

	"github.com/rs/zerolog"
)

const (
	REST_SERVICE_PING = "http.RestController.Ping"
)

func SendJsonResponse(w http.ResponseWriter, statusCode int, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(result)
}

// swagger:model
type Result struct {
	// the service result object
	//
	// Extensions:
	// ---
	// x-go-type: interface{}
	// ---
	Value interface{} `json:"result"`
}

// swagger:model
type ApiError struct {
	// the error details
	Code int `json:"code"`
	// http error code
	Message string `json:"error"`
}

type RestController struct {
	globalConfiguration *config.Configuration
	logger              zerolog.Logger
}

func NewRestController(cfg *config.Configuration, logger zerolog.Logger) *RestController {
	return &RestController{
		globalConfiguration: cfg,
		logger:              logger,
	}
}

func (c *RestController) Ping(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context()).With().Str("func", REST_SERVICE_PING).Logger()

	logger.Info().Msg("request ends with no error")
	SendJsonResponse(w, http.StatusOK, Result{
		Value: "ping",
	})
}
