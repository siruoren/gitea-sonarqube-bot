#!/usr/bin/env bash

if [ $# -gt 0 ]; then
    exec "$@"
else
    exec /usr/local/bin/gitea-sonarqube-bot
fi
