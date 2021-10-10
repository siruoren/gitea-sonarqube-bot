# Gitea SonarQube PR Bot

_Gitea SonarQube PR Bot_ is a bot that receives messages from both SonarQube and Gitea to help developers 
being productive. The idea behind this project is the missing ALM integration of Gitea in SonarQube. Unfortunately, 
this [won't be added in near future](https://github.com/SonarSource/sonarqube/pull/3248#issuecomment-701334327). 
_Gitea SonarQube PR Bot_ aims to fill the gap between working on pull requests and being notified on quality changes. 
Luckily, both endpoints have a proper REST API to communicate with each others.

- [Gitea SonarQube PR Bot](#gitea-sonarqube-pr-bot)
  - [Workflow](#workflow)
  - [Setup](#setup)
  - [Bot configuration](#bot-configuration)
  - [Contributing](#contributing)
  - [TODOs](#todos)
    - [Possible improvements](#possible-improvements)
  - [License](#license)

## Workflow

![Workflow](docs/workflow.png)

**Insights**

- Bot activities
    - extract data from SonarQube
        - [x] Read payload from hook post to receive project,branch/pr,quality-gate
        - [x] Load "api/measures/component"
    - [x] comment PR in Gitea (/repos/{owner}/{repo}/issues/{index}/comments)
    - [x] updates status check (either failing/success)
    - [ ] listen on "/sq-bot review" comments
      - [ ] comment PR in Gitea (/repos/{owner}/{repo}/issues/{index}/comments)
      - [ ] updates status check (either failing/success)

## Setup

**SonarQube**  
- Create a user and grant permissions to "Browse on project" for the desired project
- Create a token for this user that will be used by the bot.
- Create a webhook pointing to `https://<bot-url>/sonarqube`. Consider securing it with a secret.

**Gitea**  
- Create a user and grant permissions to "Read project" for the desired projects including access to "Pull Requests"
- Create a token for this user that will be used by the bot.
- Create a project/organization/system webhook pointing to `https://<bot-url>/gitea`. Consider securing it with a secret.

## Bot configuration

See [config.example.yaml](config/config.example.yaml) for a full configuration specification and description.

## Contributing

Expected workflow is: Fork -> Patch -> Push -> Pull Request

NOTES:

- **Please read and follow the [CONTRIBUTORS GUIDE](CONTRIBUTING.md).**

## TODOs

- [ ] Validate configuration on startup
- [ ] Verify webhook secrets
- [ ] Only post status-check (Opt-in/out)
- [ ] Maybe drop `PRBOT_CONFIG_PATH` environment variable in favor of `--config path/to/config.yaml` cli attribute
- [ ] Configure SonarQube PR branch naming pattern for more flexibility (currently focused on Jenkins with [Gitea Plugin](https://github.com/jenkinsci/gitea-plugin))
- [ ] Configuration live reloading
- [ ] _Caching_ of outgoing requests in case the target is not available
- [ ] Parsable logging for monitoring
- [ ] Official image for containerized hosting
- [ ] Helm chart for Kubernetes

### Possible improvements

- Reuse existing posted comment for updates via SonarQube webhook or `/sq-bot` comments  
Therefore storing or dynamically retrieving the previous comment id and modify content (/repos/{owner}/{repo}/issues/comments/{id})
- Add more information to posted comment
  - Read "api/project_pull_requests" to get current issue counts and current state
  - Load "api/issues/search" to get detailed information for unresolved issues
- Maybe directly show issues via review comments

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
