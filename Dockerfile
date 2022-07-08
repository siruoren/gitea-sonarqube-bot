###################################
# Build stages
###################################
FROM golang:1.18-alpine3.16@sha256:7cc62574fcf9c5fb87ad42a9789d5539a6a085971d58ee75dd2ee146cb8a8695 AS build-go

ARG GOPROXY
ENV GOPROXY ${GOPROXY:-direct}

RUN apk update \
    && apk --no-cache add build-base git bash

COPY . ${GOPATH}/src/bot
WORKDIR ${GOPATH}/src/bot

RUN go build ./cmd/gitea-sonarqube-bot

###################################
# Production image
###################################
FROM alpine:3.16@sha256:686d8c9dfa6f3ccfc8230bc3178d23f84eeaf7e457f36f271ab1acc53015037c
LABEL maintainer="justusbunsi <sk.bunsenbrenner@gmail.com>"

RUN apk update \
    && apk --no-cache add ca-certificates bash \
    && rm -rf /var/cache/apk/*

RUN addgroup -S -g 1000 bot \
    && adduser -S -D -H -h /home/bot -s /bin/bash -u 1000 -G bot bot

RUN mkdir -p /home/bot/config/
RUN chown bot:bot /home/bot/config/

COPY --chown=bot:bot docker /
COPY --from=build-go --chown=bot:bot /go/src/bot/gitea-sonarqube-bot /usr/local/bin/gitea-sonarqube-bot

#bot:bot
USER 1000:1000
WORKDIR /home/bot
ENV HOME=/home/bot

EXPOSE 3000
ENV GIN_MODE "release"
ENV GITEA_SQ_BOT_CONFIG_PATH "/home/bot/config/config.yaml"
ENV GITEA_SQ_BOT_PORT "3000"

VOLUME ["/home/bot/config/"]
RUN ["chmod", "+x", "/usr/local/bin/docker-entrypoint.sh"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
