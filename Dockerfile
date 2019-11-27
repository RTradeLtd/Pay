############
## PART 1 ##
############

# declare build image and arguments
FROM golang:alpine3.10 as build-env
ARG XVERSION
ENV BUILD_HOME=/BUILD \
    VERSION=${VERSION}
# setup build directory
RUN mkdir -p ${BUILD_HOME}
RUN apk add build-base linux-headers
# mount source code into build directory
# this code must be vendored
# this takes code from the current working directory (build context)
# and adds it into the image
ADD . ${BUILD_HOME}
# set working directory
WORKDIR ${BUILD_HOME}

# build binary from vendored source code
RUN go mod download
RUN go build \
    -ldflags "-X main.Version=$VERSION" \
    -o /bin/pay \
    ./cmd/pay

############
## PART 2 ##
############

# create clean image for the binary
FROM alpine:3.10
LABEL maintainer "RTrade Technologies Ltd."
# copy built binary into the clean image
COPY --from=build-env /bin/pay /usr/local/pay
# setup directory for TemporalX stuff
RUN mkdir /temporal && mkdir -p /var/log/temporal

# set default configuration
ENV CONFIG_DAG /temporal/config.json
COPY ./test/config.json /temporal/config.json

# the default command to run when container starts
ENTRYPOINT ["/usr/local/pay", "--config", "/temporal/config.json"]
# default command to feed into entrypoint
CMD ["grpc", "server"]