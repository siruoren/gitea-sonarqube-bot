# Contribution Guidelines

## Table of Contents

- [Contribution Guidelines](#contribution-guidelines)
  - [Table of Contents](#table-of-contents)
  - [Setup development environment](#setup-development-environment)
  - [Testing](#testing)
  - [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)

## Setup development environment

```bash
# Build docker environment
docker build -t gitea-sonarqube-pr-bot/dev -f contrib/Dockerfile contrib

# Start the environment
docker run --rm -it -p 49182:3000 -v "$(pwd):/projects" gitea-sonarqube-pr-bot/dev

# Build the binary
go build ./cmd/gitea-sonarqube-bot

# Start the server
./gitea-sonarqube-bot
```

## Testing

```bash
# generic test execution
go test ./...

# or with coverage report
go test -coverprofile cover.out ./...
```

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
