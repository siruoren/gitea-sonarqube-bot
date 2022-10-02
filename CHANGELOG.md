# Changelog

## v0.2.2

### ☸️ Helm Chart

- Bump default image tag to `v0.2.2`

### 👻 Maintenance

- Remove `fvbock/endless` dependency
- Require Golang 1.19 for builds
- Update base Docker images
- Update dependencies to newest versions

## v0.2.1

### 🤖 Application

- Allow configuring listening port
- Allow changing naming pattern for Pull Requests
- Improve error handling for SonarQube communication

### 🐳 Docker image

- Add `GITEA_SQ_BOT_PORT` environment variable

### ☸️ Helm Chart

- Add `.Values.app.listeningPort` parameter
- Add `.Values.app.configuration.namingPattern` parameters

## v0.2.0

### 🤖 Application

- Add webhook secret validation
- Improve configuration file flexibility
- Stop log output for `/ping` and `/favicon.ico` endpoints

### 🐳 Docker image

- Add `GITEA_SQ_BOT_CONFIG_PATH` environment variable

### ☸️ Helm Chart

- Add `.Values.app.configLocationOverride` parameter
- Bump default image tag to `v0.2.0`

## v0.1.1

### ☸️ Helm Chart

- Bump default image tag to `v0.1.1`

### 👻 Maintenance

- Bump Golang version to 1.18
- Update dependencies to newest versions

## v0.1.0

Initial release
