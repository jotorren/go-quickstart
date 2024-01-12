# A very simple Golang REST API

This is an opinionated *Go project template* you *can use* as a starting point for your project. Current version includes the following *key aspects*:

- **Hexagonal architecture**
  
  ```textile
  go-quickstart/
   ├──src/
   │   ├──application/
   │   ├──cmd/             (main.go)
   │   │   ├──docker/
   │   │   └──local/
   │   ├──domain/
   │   ├──infrastructure/
   │   └──resources/       (application.yaml)
   └──deploy/
  ```

- **uber-go/fx** dependency injection framework
  
  ```shell
  go-quickstart$ more src/cmd/docker/main.go
  ```
  
  ```go
  ...
  fx.New(
          fx.Provide(
                  NewRootLogger,
                  config.NewConfiguration,
                  transport.NewHTTPServer,
                  transport.NewMuxRouter,
                  transport.NewRestController,
          ),
          fx.WithLogger(func(rootLogger zerolog.Logger) fxevent.Logger {
                  return fxlogger.WithZerolog(rootLogger.Level(zerolog.WarnLevel))()
          }),
          fx.Invoke(
                  func(cfg *config.Configuration, rootLogger zerolog.Logger) {
                          rootLogger.Info().Msg("application.yaml read")
                          cfg.Print(rootLogger)
                  },
                  func(*http.Server) {
                          // start server
                  },
          ),
  ).Run()
  ...
  ```

- **uber-go/config** injection-friendly YAML configuration
  
  ```shell
  go-quickstart$ more src/resources/application.yaml
  ```
  
  ```yaml
  log:
    defaultlevel: ${LOG_LEVEL:1}
    packageslevel:
      http: ${LOG_LEVEL_HTTP:1}
  server:
    port: ${SERVER_PORT:8080}
    origins: ${SERVER_ALLOWED_ORIGINS}
  ```
  
  Expands any environment variable references using the `os.LookupEnv` function. All characters between the opening curly brace and the first colon are used as the key, and all characters from the colon to the closing curly brace are used as the default value. If a variable isn't found, the default value is used.
  
  $$ is expanded to a literal \$
  
  YAML  configuration files are embedded into the application binary at compile-time:
  
  ```shell
  go-quickstart$ more src/resources.go
  ```
  
  ```go
  package src
  
  import _ "embed"
  
  //go:embed resources/application.yaml
  var ApplicationYaml []byte
  ```

- **net/http** with **CORS** security
  
  ```shell
  go-quickstart$ more src/infrastructure/transport/httpadapter.go
  ```
  
  ```go
  func NewHTTPServer(cfg *config.Configuration, router *mux.Router, lc fx.Lifecycle, logger zerolog.Logger) *http.Server {
      ml, ok := cfg.Log.PackagesLevel[HTTP_PACKAGE_NAME]
      if ok {
              logger = logger.Level(zerolog.Level(ml))
      }
  
      // CORS setup
      co := cors.New(cors.Options{
              AllowedOrigins:   cfg.Server.Origins,
              AllowCredentials: true,
              AllowedMethods:   []string{http.MethodGet, http.MethodPost},
              AllowedHeaders:   []string{"Authorization"},
              Debug:            true,
              Logger:           &logger,
      })
      srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: co.Handler(router)}
      ...
  }
  ```

- **gorilla/mux** router and a **subrouter** to use specific middleware for specific routes
  
  ```shell
  go-quickstart$ more src/infrastructure/transport/httpadapter.go
  ```
  
  ```go
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
  ```

- **rs/zerolog** logging library. **Per request contextual logging** (all traces within the same request will share the same unique id)
  
  ```shell
  go-quickstart$ more src/infrastructure/transport/httpadapter.go
  ```
  
  ```go
  func loggerMiddleware(logger zerolog.Logger) mux.MiddlewareFunc {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              ...
              reqlogger := logger.With().
                      Str("request_id", uuid.New().String()).
                      Logger()
              ...
              ctx := reqlogger.WithContext(r.Context())
              next.ServeHTTP(lrw, r.WithContext(ctx))
          })
      }
  }
  ```

## Build & Run

Get the source code:

```shell
$ git clone https://github.com/jotorren/go-quickstart.git
Cloning into 'go-quickstart'...
remote: Enumerating objects: 59, done.
remote: Counting objects: 100% (59/59), done.
remote: Compressing objects: 100% (38/38), done.
remote: Total 59 (delta 13), reused 51 (delta 9), pack-reused 0
Receiving objects: 100% (59/59), 13.93 KiB | 356.00 KiB/s, done.
Resolving deltas: 100% (13/13), done.
```

Use `make `to compile a final executable binary:

```shell
$ cd go-quickstart/
go-quickstart$ cd src
go-quickstart/src$ make build
********** Building binary...
go build -o myapp cmd/local/main.go

go-quickstart/src$
go-quickstart/src$ ls -l myapp
-rwxr-xr-x 1 jotorren jotorren 11433338 Jan 10 16:39 myapp
```

Start the service:

```shell
go-quickstart/src$ ./myapp
Wed, 10 Jan 2024 16:40:14 CET INF cmd/local/main.go:40 > application.yaml read
Wed, 10 Jan 2024 16:40:14 CET INF infrastructure/config/configuration.go:43 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]}}
Wed, 10 Jan 2024 16:40:14 CET INF infrastructure/transport/httpadapter.go:46 > starting HTTP server addr=:8080
```

Run the `curl `command followed by the target URL `/api/v1/ping`:

```shell
$ curl http://localhost:8080/api/v1/ping
{"result":"ping"} 
```

To test **CORS** configuration:

1. **Sending a regular request**. If the origin is allowed, the response will include the `Access-Control-Allow-Origin` header. Otherwise, that header will not appear.

```shell
$ curl -H "Origin: http://localhost:3000" --head -X GET http://localhost:8080/api/v1/ping
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://localhost:3000
Content-Type: application/json
Vary: Origin
Date: Fri, 12 Jan 2024 02:04:29 GMT
Content-Length: 18
```

Adding `--head` outputs only headers. Looking at the service log:

```log
Fri, 12 Jan 2024 02:04:29 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 > Handler: Actual request
Fri, 12 Jan 2024 02:04:29 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 >   Actual response added headers: map[Access-Control-Allow-Credentials:[true] Access-Control-Allow-Origin:[http://localhost:3000] Vary:[Origin]]
Fri, 12 Jan 2024 02:04:29 UTC INF app/src/infrastructure/transport/restcontroller.go:46 > request ends with no error func=http.RestController.Ping request_id=9f41f2b6-13af-413b-b29f-e9d96f8364e4
Fri, 12 Jan 2024 02:04:29 UTC DBG app/src/infrastructure/transport/httpadapter.go:119 > func=http.loggerMiddleware method=GET request_id=9f41f2b6-13af-413b-b29f-e9d96f8364e4 status_code=200 total_elapsed_ms=0.075355 url=/api/v1/ping user_agent=curl/7.74.0
```

```shell
$ curl -H "Origin: http://some.other" --head -X GET http://localhost:8080/api/v1/ping
HTTP/1.1 200 OK
Content-Type: application/json
Vary: Origin
Date: Fri, 12 Jan 2024 02:05:20 GMT
Content-Length: 18
```

```log
Fri, 12 Jan 2024 02:05:20 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 > Handler: Actual request
Fri, 12 Jan 2024 02:05:20 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 >   Actual request no headers added: origin 'http://some.other' not allowed
Fri, 12 Jan 2024 02:05:20 UTC INF app/src/infrastructure/transport/restcontroller.go:46 > request ends with no error func=http.RestController.Ping request_id=77f1fe29-410d-45bb-91f2-2f5b54f0a8a2
Fri, 12 Jan 2024 02:05:20 UTC DBG app/src/infrastructure/transport/httpadapter.go:119 > func=http.loggerMiddleware method=GET request_id=77f1fe29-410d-45bb-91f2-2f5b54f0a8a2 status_code=200 total_elapsed_ms=0.100864 url=/api/v1/ping user_agent=curl/7.74.0
```

```diff
Actual request no headers added: origin '[http://some.other'](http://some.other') not allowed
```

2. **Sending a preflight request**. If the preflight request is successful, the response should include the `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, and `Access-Control-Allow-Headers` headers. Otherwise, these headers shouldn't appear.

```shell
$ curl -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: GET" -X OPTIONS --head http://localhost:8080/api/v1/ping
HTTP/1.1 204 No Content
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET
Access-Control-Allow-Origin: http://localhost:3000
Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers
Date: Fri, 12 Jan 2024 01:59:51 GMT
```

```log
Fri, 12 Jan 2024 01:59:51 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 > Handler: Preflight request
Fri, 12 Jan 2024 01:59:51 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 >   Preflight response headers: map[Access-Control-Allow-Credentials:[true] Access-Control-Allow-Methods:[GET] Access-Control-Allow-Origin:[http://localhost:3000] Vary:[Origin, Access-Control-Request-Method, Access-Control-Request-Headers]]
```

```shell
$ curl -H "Origin: http://some.other" -H "Access-Control-Request-Method: GET" -X OPTIONS --head http://localhost:8080/api/v1/ping
HTTP/1.1 204 No Content
Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers
Date: Fri, 12 Jan 2024 01:59:37 GMT
```

```log
Fri, 12 Jan 2024 01:59:37 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 > Handler: Preflight request
Fri, 12 Jan 2024 01:59:37 UTC DBG go/pkg/mod/github.com/rs/cors@v1.10.1/cors.go:445 >   Preflight aborted: origin 'http://some.other' not allowed
```

```diff
Preflight aborted: origin '[http://some.other'](http://some.other') not allowed
```

### Docker multistage build

If  you don't have `Golang` installed, you don't need to install it, you can use `docker` to build the `Dockerfile` found in the project's root directory: 

```dockerfile
FROM golang:1.21 AS build
WORKDIR /app/
COPY ./src/ ./src/
WORKDIR /app/src/
RUN go env -w GOPROXY=direct
RUN CGO_ENABLED=0 go build -o ../myapp cmd/docker/main.go

FROM alpine:3.18 AS runtime
RUN addgroup -S nonroot  
&& adduser -S nonroot -G nonroot
COPY --from=build /app/myapp /app/myapp
USER nonroot
CMD ["/app/myapp"]
```

Just run:

```bash
go-quickstart$ docker build -t myapp_server .
Sending build context to Docker daemon  437.8kB
Step 1/11 : FROM golang:1.21 AS build
 ---> 4c88d2e04e7d
Step 2/11 : WORKDIR /app/
 ---> Running in 47fd8008d012
Removing intermediate container 47fd8008d012
 ---> dc0318615de3
Step 3/11 : COPY ./src/ ./src/
 ---> 4d4cf56daaba
Step 4/11 : WORKDIR /app/src/
 ---> Running in 8fade6aaec7f
Removing intermediate container 8fade6aaec7f
 ---> 6181aa91e9d5
Step 5/11 : RUN go env -w GOPROXY=direct
 ---> Running in 723f5251309f
Removing intermediate container 723f5251309f
 ---> f76f7f8578b3
Step 6/11 : RUN CGO_ENABLED=0 go build -o ../myapp cmd/docker/main.go
 ---> Running in 6b85eb1627f0
go: downloading go.uber.org/fx v1.20.1
go: downloading github.com/ipfans/fxlogger v0.2.0
go: downloading github.com/rs/zerolog v1.31.0
go: downloading go.uber.org/config v1.4.0
go: downloading github.com/google/uuid v1.5.0
go: downloading github.com/gorilla/mux v1.8.1
go: downloading github.com/rs/cors v1.10.1
go: downloading github.com/mattn/go-colorable v0.1.13
go: downloading go.uber.org/multierr v1.6.0
go: downloading golang.org/x/text v0.3.2
go: downloading gopkg.in/yaml.v2 v2.2.5
go: downloading go.uber.org/atomic v1.7.0
go: downloading github.com/mattn/go-isatty v0.0.19
go: downloading go.uber.org/zap v1.23.0
go: downloading golang.org/x/sys v0.12.0
go: downloading go.uber.org/dig v1.17.0
Removing intermediate container 6b85eb1627f0
 ---> 0c15d0f3703f
Step 7/11 : FROM alpine:3.18 AS runtime
 ---> 8ca4688f4f35
Step 8/11 : RUN addgroup -S nonroot     && adduser -S nonroot -G nonroot
 ---> Running in 65f4ecd034c4
Removing intermediate container 65f4ecd034c4
 ---> 8ba8f377ad6d
Step 9/11 : COPY --from=build /app/myapp  /app/myapp
 ---> 4dc3c2ce72d2
Step 10/11 : USER nonroot
 ---> Running in 0b5a89420c23
Removing intermediate container 0b5a89420c23
 ---> 93f8aec7447b
Step 11/11 : CMD ["/app/myapp"]
 ---> Running in fedabce904e1
Removing intermediate container fedabce904e1
 ---> 072fcdee271f
Successfully built 072fcdee271f
Successfully tagged myapp_server:latest

go-quickstart$ docker image ls
REPOSITORY                         TAG               IMAGE ID       CREATED          SIZE
myapp_server                       latest            072fcdee271f   45 seconds ago   18.9MB
<none>                             <none>            0c15d0f3703f   47 seconds ago   984MB
```

Once the image is built, to start the service:

```shell
go-quickstart$ docker run --env SERVER_ALLOWED_ORIGINS="['http://localhost:3000']" --env LOG_LEVEL_HTTP=0 -dp 127.0.0.1:8080:8080
myapp_server
ed97b8661cd80398d998a8d1ea91ad0e4ab4cb5e944d59587af2babe23148a47

go-quickstart$ docker ps
CONTAINER ID   IMAGE          COMMAND        CREATED         STATUS         PORTS                      NAMES
ed97b8661cd8   myapp_server   "/app/myapp"   4 seconds ago   Up 3 seconds   127.0.0.1:8080->8080/tcp   mystifying_dubinsky

go-quickstart$ docker logs ed97b8661cd8
Thu, 11 Jan 2024 13:27:39 UTC INF app/src/cmd/docker/main.go:36 > application.yaml read
Thu, 11 Jan 2024 13:27:39 UTC INF app/src/infrastructure/config/configuration.go:43 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]}}
Thu, 11 Jan 2024 13:27:39 UTC INF app/src/infrastructure/transport/httpadapter.go:46 > starting HTTP server addr=:8080

go-quickstart$ curl -X GET http://localhost:8080/api/v1/ping
{"result":"ping"}
```

Or if you prefer to use `docker-compose`, creating the image and running the container can be done in one single step:

```bash
go-quickstart$ cd deploy/

go-quickstart/deploy$ more docker-compose.yaml
version: "3.3"

services:
  myapp-server:
    build:
      dockerfile: Dockerfile
      context: ../
    restart: always
    environment:
      - SERVER_ALLOWED_ORIGINS=['http://localhost:3000']
      - LOG_LEVEL_HTTP=0
    ports:
      - "8080:8080"

go-quickstart/deploy$ docker-compose up -d --build myapp-server
Creating network "deploy_default" with the default driver
Building myapp-server
Step 1/11 : FROM golang:1.21 AS build
 ---> 4c88d2e04e7d
Step 2/11 : WORKDIR /app/
 ---> Running in 8488b127e5e4
Removing intermediate container 8488b127e5e4
 ---> 2991878ddf50
Step 3/11 : COPY ./src/ ./src/
 ---> 2f17f908dd92
Step 4/11 : WORKDIR /app/src/
 ---> Running in 84d82d8bc673
Removing intermediate container 84d82d8bc673
 ---> 8afbe33a4390
Step 5/11 : RUN go env -w GOPROXY=direct
 ---> Running in 0d48d880f77a
Removing intermediate container 0d48d880f77a
 ---> 331f37cf7e92
Step 6/11 : RUN CGO_ENABLED=0 go build -o ../myapp cmd/docker/main.go
 ---> Running in d8d25b860173
go: downloading go.uber.org/fx v1.20.1
go: downloading github.com/ipfans/fxlogger v0.2.0
go: downloading github.com/rs/zerolog v1.31.0
go: downloading go.uber.org/config v1.4.0
go: downloading github.com/google/uuid v1.5.0
go: downloading github.com/gorilla/mux v1.8.1
go: downloading github.com/rs/cors v1.10.1
go: downloading github.com/mattn/go-colorable v0.1.13
go: downloading gopkg.in/yaml.v2 v2.2.5
go: downloading go.uber.org/multierr v1.6.0
go: downloading golang.org/x/text v0.3.2
go: downloading golang.org/x/sys v0.12.0
go: downloading go.uber.org/dig v1.17.0
go: downloading go.uber.org/zap v1.23.0
go: downloading github.com/mattn/go-isatty v0.0.19
go: downloading go.uber.org/atomic v1.7.0
Removing intermediate container d8d25b860173
 ---> 4bcba143f2c7

Step 7/11 : FROM alpine:3.18 AS runtime
 ---> 8ca4688f4f35
Step 8/11 : RUN addgroup -S nonroot     && adduser -S nonroot -G nonroot
 ---> Running in 0138ef7c9882
Removing intermediate container 0138ef7c9882
 ---> 5f4f1ba4bcfa
Step 9/11 : COPY --from=build /app/myapp  /app/myapp
 ---> 057522715e5d
Step 10/11 : USER nonroot
 ---> Running in 5cf7f8a2158a
Removing intermediate container 5cf7f8a2158a
 ---> e731c3dd8232
Step 11/11 : CMD ["/app/myapp"]
 ---> Running in dfac91a33576
Removing intermediate container dfac91a33576
 ---> 9cef1edc154b

Successfully built 9cef1edc154b
Successfully tagged deploy_myapp-server:latest
Creating deploy_myapp-server_1 ... done
```

```bash
go-quickstart/deploy$ docker image ls
REPOSITORY                         TAG               IMAGE ID       CREATED         SIZE
deploy_myapp-server                latest            9cef1edc154b   2 minutes ago   18.9MB
<none>                             <none>            4bcba143f2c7   2 minutes ago   984MB

go-quickstart/deploy$ docker ps
CONTAINER ID   IMAGE                                 COMMAND                  CREATED              STATUS              PORTS                                                                                                                                             NAMES
acbfea5f0839   deploy_myapp-server                   "/app/myapp"             About a minute ago   Up About a minute   0.0.0.0:8080->8080/tcp                                                                                                                            deploy_myapp-server_1

go-quickstart/deploy$ docker logs acbfea5f0839   
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/cmd/docker/main.go:36 > application.yaml read
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/infrastructure/config/configuration.go:43 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]}}
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/infrastructure/transport/httpadapter.go:46 > starting HTTP server addr=:8080

go-quickstart$ curl http://localhost:8080/api/v1/ping
{"result":"ping"}
```

In both cases you can check the image used to compile the final executable binary:

```bash
go-quickstart/deploy$ docker run --rm -it 4bcba143f2c7
root@3e5f7d2f4f89:/app/src# ls
application  cmd  domain  go.mod  go.sum  infrastructure  resources  resources.go
root@3e5f7d2f4f89:/app/src# ls ..
myapp  src
root@3e5f7d2f4f89:/app/src# exit
exit
go-quickstart/deploy$
```

Where `4bcba143f2c7` is the ID of the image created during the **first docker stage**:

```dockerfile
FROM golang:1.21 AS build
WORKDIR /app/
COPY ./src/ ./src/
WORKDIR /app/src/
RUN go env -w GOPROXY=direct
RUN CGO_ENABLED=0 go build -o ../myapp cmd/docker/main.go
```

## Notes

The source code found in the `main` branch is the baseline on which different functionalities will be added (security, `ORM`, observability...), each of them available in a specific branch of this repository.

## Support, Questions, or Feedback

I'll accept pretty much everything so feel free to open a Pull-Request
