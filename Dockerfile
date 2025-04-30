FROM harbor.nbfc.io/proxy_cache/library/python:3.13-alpine3.21 AS cbuilder
RUN pip install --upgrade pip && \
    pip install jsonschema jinja2 && \
    apk update && \
    apk add --no-cache git make build-base curl-dev perl openssl-dev

COPY ./ota-agent /ota-agent
WORKDIR /ota-agent
RUN make

FROM harbor.nbfc.io/proxy_cache/library/golang:1.24.2-alpine3.21 AS gobuilder
WORKDIR /app
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./internal ./internal
COPY go.mod go.mod
RUN apk update && \
    apk add --no-cache git && \
    go mod tidy && \
    go mod vendor && \
    go mod verify && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-extldflags=-static" -o ./esp32-flashjob ./cmd/esp32-flashjob

FROM scratch AS intermediate
WORKDIR /intermediate
COPY --from=gobuilder /app/esp32-flashjob /intermediate/esp32-flashjob
COPY --from=cbuilder /ota-agent/ota-agent /intermediate/ota-agent

FROM harbor.nbfc.io/proxy_cache/library/alpine:3.21
RUN apk update && \
    apk add --no-cache curl-dev openssl-dev
COPY --from=intermediate /intermediate /usr/local/bin/
COPY ./misc/certs /ota/certs
CMD ["/usr/local/bin/esp32-flashjob"]
