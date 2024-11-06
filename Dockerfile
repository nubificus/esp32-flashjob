FROM docker.io/library/python:3.9.20-bookworm AS cbuilder
RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y git make build-essential libssl-dev
RUN pip install --upgrade pip
RUN pip install jsonschema jinja2


COPY ./ota-agent /ota-agent
WORKDIR /ota-agent
RUN git submodule update --init
WORKDIR /ota-agent/mbedtls
RUN git submodule update --init
WORKDIR /ota-agent
RUN make

FROM docker.io/library/busybox:latest 
COPY --from=cbuilder /ota-agent/ota-agent /ota-agent