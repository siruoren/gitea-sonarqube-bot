# Contribution Guidelines

## Table of Contents

- [Contribution Guidelines](#contribution-guidelines)
  - [Table of Contents](#table-of-contents)
  - [Setup development environment](#setup-development-environment)
  - [Build and Run](#build-and-run)
  - [Testing](#testing)
  - [Helm Chart](#helm-chart)
  - [Release](#release)
  - [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)

## Setup development environment

```bash
# Build docker environment
docker build -t gitea-sonarqube-bot/dev -f contrib/Dockerfile contrib

# Start the environment
docker run --rm -it -p 49182:3000 -v "$(pwd):/projects" gitea-sonarqube-bot/dev
```

## Build and Run

```bash
# Build the binary
make build
# Start the server
make run

# or all in once
make build run
```

## Testing

```bash
make test
# or
make coverage
```

## Helm Chart

The [Parameters section](helm/README.md#parameters) is auto-generated using [readme-generator-for-helm](https://github.com/bitnami-labs/readme-generator-for-helm).
When modifying anything in the `helm` directory, remember to update the documentation by running

```bash
make helm-params
```

## Release

For local purposes

```bash
docker build -t gitea-sonarqube-bot/prod .
```

**Docker image**

```bash
docker build -t justusbunsi/gitea-sonarqube-bot:$TAG .
docker push justusbunsi/gitea-sonarqube-bot:$TAG
```

**Helm Chart**

```bash
make helm-pack
```

Use the two files in `helm-releases` and push them to the `charts` branch.

## Developer Certificate of Origin (DCO)

I consider the act of contributing to the code by submitting a Pull Request as the "Sign off" or agreement to the 
certifications and terms of the [DCO](DCO) and [MIT license](LICENSE). No further action is required. Additionally, 
you could add a line at the end of your commit message.

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

If you set your `user.name` and `user.email` git configs, you can add the line to the end of your commit automatically 
with `git commit -s`.

I assume in good faith that the information you provide is legally binding.
