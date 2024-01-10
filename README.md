# A very simple Golang REST API

Key aspects:

- **Hexagonal architecture**

- **uber-go/fx** dependency injection framework

- **uber-go/config** injection-friendly YAML configuration

- **net/http** with **CORS** security

- **gorilla/mux** router and URL matcher. With a **subrouter** to use specific middleware for specific routes.

- **rs/zerolog** logging library. **Per request contextual logging**.

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

Compile it by means of the included Makefile:

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

Open another terminal and using `curl `try to access the relative URI `/api/v1/ping`:

```shell
$ curl -X GET http://localhost:8080/api/v1/ping
{"result":"ping"} 
```
