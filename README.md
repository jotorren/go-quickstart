# A very simple Golang REST API

This is an opinionated *Go project template* you *can use* as a starting point for your project. Current version includes the following *key aspects*:

- **Hexagonal architecture**

- **uber-go/fx** dependency injection framework

- **uber-go/config** injection-friendly YAML configuration

- **net/http** with **CORS** security

- **gorilla/mux** router and a **subrouter** to use specific middleware for specific routes

- **rs/zerolog** logging library. **Per request contextual logging**

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
$ curl -X GET http://localhost:8080/api/v1/ping
{"result":"ping"} 
```

### Docker multistage build



```bash
go-quickstart$ more Dockerfile 
FROM golang:1.21 AS build
WORKDIR /app/
COPY ./src/ ./src/
WORKDIR /app/src/
RUN go env -w GOPROXY=direct
RUN CGO_ENABLED=0 go build -o ../myapp cmd/docker/main.go

FROM alpine:3.18 AS runtime
RUN addgroup -S nonroot \
    && adduser -S nonroot -G nonroot
COPY --from=build /app/myapp  /app/myapp
USER nonroot
CMD ["/app/myapp"]

```



```bash
go-quickstart$ cd deploy/

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
go-quickstart/deploy$ docker ps
CONTAINER ID   IMAGE                                 COMMAND                  CREATED              STATUS              PORTS                                                                                                                                             NAMES
acbfea5f0839   deploy_myapp-server                   "/app/myapp"             About a minute ago   Up About a minute   0.0.0.0:8080->8080/tcp                                                                                                                            deploy_myapp-server_1

go-quickstart/deploy$ docker logs acbfea5f0839   
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/cmd/docker/main.go:36 > application.yaml read
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/infrastructure/config/configuration.go:43 > {Log:{DefaultLevel:1 PackagesLevel:map[http:0]} Server:{Port:8080 Origins:[http://localhost:3000]}}
Wed, 10 Jan 2024 22:55:02 UTC INF app/src/infrastructure/transport/httpadapter.go:46 > starting HTTP server addr=:8080
```



```bash
go-quickstart/deploy$ docker image ls
REPOSITORY                         TAG               IMAGE ID       CREATED         SIZE
deploy_myapp-server                latest            9cef1edc154b   2 minutes ago   18.9MB
<none>                             <none>            4bcba143f2c7   2 minutes ago   984MB
```



```bash
go-quickstart/deploy$ docker container run --rm -it 4bcba143f2c7
root@3e5f7d2f4f89:/app/src# ls
application  cmd  domain  go.mod  go.sum  infrastructure  resources  resources.go
root@3e5f7d2f4f89:/app/src# ls ..
myapp  src
root@3e5f7d2f4f89:/app/src# exit
exit
go-quickstart/deploy$
```
