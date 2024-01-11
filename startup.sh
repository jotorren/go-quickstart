#! /bin/sh -x

/usr/bin/socat TCP-LISTEN:8090,fork,bind=127.0.0.1 TCP:${KEYCLOAK_CONTAINER}:8080 &

/app/myapp
