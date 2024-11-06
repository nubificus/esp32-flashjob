FROM docker.io/library/python:3.9.20-bookworm AS cbuilder
RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y git make build-essential libssl-dev
RUN pip install --upgrade pip
RUN pip install jsonschema jinja2
COPY ./ota-agent /ota-agent
WORKDIR /ota-agent
RUN make

FROM cgr.dev/chainguard/go:latest AS gobuilder
COPY ./cmd /sota/cmd
COPY ./pkg /sota/pkg
COPY ./internal /sota/internal
COPY ./go.mod /sota
WORKDIR /sota
RUN go mod tidy
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-extldflags=-static" -o ./esp32-sota-bin ./cmd/esp32-sota

FROM cgr.dev/chainguard/static:latest
COPY --from=gobuilder /sota/esp32-sota-bin /esp32-sota
COPY --from=cbuilder /ota-agent/ota-agent /ota-agent