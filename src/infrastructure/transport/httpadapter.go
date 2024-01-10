package transport

import (
	"context"
	"net"
	"net/http"
	"time"

	"tsib/quickstart/infrastructure/config"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

const (
	HTTP_PACKAGE_NAME           = "http"
	HTTP_LOGGER_MIDDLEWARE_NAME = "http.loggerMiddleware"
)

func NewHTTPServer(cfg *config.Configuration, router *mux.Router, lc fx.Lifecycle, logger zerolog.Logger) *http.Server {
	ml, ok := cfg.Log.PackagesLevel[HTTP_PACKAGE_NAME]
	if ok {
		logger = logger.Level(zerolog.Level(ml))
	}

	// CORS setup
	corslogger := logger.Level(zerolog.InfoLevel)
	co := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.Origins,
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
		Logger:           &corslogger,
	})
	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: co.Handler(router)}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			logger.Info().Str("addr", srv.Addr).Msg("starting HTTP server")
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

type MuxRouterParams struct {
	fx.In

	Cfg        *config.Configuration
	Controller *RestController
	Logger     zerolog.Logger
}

func NewMuxRouter(p MuxRouterParams) *mux.Router {
	ml, ok := p.Cfg.Log.PackagesLevel[HTTP_PACKAGE_NAME]
	if ok {
		p.Logger = p.Logger.Level(zerolog.Level(ml))
	}

	router := mux.NewRouter()
	router.Use(loggerMiddleware(p.Logger))

	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/ping", p.Controller.Ping).Methods("GET")

	return router
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggerMiddleware(logger zerolog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqlogger := logger.With().
				Str("request_id", uuid.New().String()).
				Logger()

			lrw := newLoggingResponseWriter(w)
			defer func() {
				panicVal := recover()
				if panicVal != nil {
					lrw.statusCode = http.StatusInternalServerError // ensure that the status code is updated
					panic(panicVal)                                 // continue panicking
				}

				if lrw.statusCode == http.StatusOK {
					reqlogger.Debug().
						Str("func", HTTP_LOGGER_MIDDLEWARE_NAME).
						Str("method", r.Method).
						Str("url", r.URL.RequestURI()).
						Str("user_agent", r.UserAgent()).
						Int("status_code", lrw.statusCode).
						Dur("total_elapsed_ms", time.Since(start)).
						Send()
				} else {
					reqlogger.Error().
						Str("func", HTTP_LOGGER_MIDDLEWARE_NAME).
						Str("method", r.Method).
						Str("url", r.URL.RequestURI()).
						Str("user_agent", r.UserAgent()).
						Int("status_code", lrw.statusCode).
						Dur("total_elapsed_ms", time.Since(start)).
						Send()
				}
			}()

			ctx := reqlogger.WithContext(r.Context())
			next.ServeHTTP(lrw, r.WithContext(ctx))
		})
	}
}
