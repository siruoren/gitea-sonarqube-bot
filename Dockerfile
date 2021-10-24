###################################
# Build stages
###################################
FROM golang:1.17-alpine3.14 AS build-go

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
FROM alpine:3.14
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

VOLUME ["/home/bot/config/"]
RUN ["chmod", "+x", "/usr/local/bin/docker-entrypoint.sh"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
