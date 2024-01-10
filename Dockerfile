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
