# Set up Golang build environment
FROM golang:alpine AS build-env
ARG TEMPORALVERSION
ENV BUILD_HOME=/go/src/github.com/RTradeLtd/Temporal_Payment-ETH \
    TEMPORALVERSION=${TEMPORALVERSION}

# Mount source code
ADD . ${BUILD_HOME}
WORKDIR ${BUILD_HOME}

# install gcc libs
RUN apk add build-base

# Build temporal binary
RUN go build -o /bin/temporal-payment \
    ./cmd/temporal-payment

# Copy binary into clean image
FROM alpine
LABEL maintainer "RTrade Technologies Ltd."
RUN mkdir -p /daemon
WORKDIR /daemon
COPY --from=build-env /bin/temporal-payment /usr/local/bin

# Set up directories
RUN mkdir /temporal \  
    mkdir -p /var/log/temporal

# Set default configuration
ENV CONFIG_DAG /temporal/config.json
COPY ./test/config.json /temporal/config.json

# Set default command
ENTRYPOINT [ "temporal-payment", "queue", "payment", "payment-confirmation"]
