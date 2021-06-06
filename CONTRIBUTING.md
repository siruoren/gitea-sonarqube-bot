# Contribution Guidelines

## Table of Contents

- [Contribution Guidelines](#contribution-guidelines)
    - [Setup development environment](#setup-development-environment)
    - [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)

## Setup development environment

```bash
# Build docker environment
docker build -t gitea-sonarqube-pr-bot/dev -f contrib/Dockerfile contrib

# start the environment
docker run --rm -it -p 9100:8080 -v "$(pwd):/projects" gitea-sonarqube-pr-bot/dev

# Start the server
go run main.go
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
