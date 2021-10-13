# Gitea SonarQube Bot

_Gitea SonarQube Bot_ is a bot that receives messages from both SonarQube and Gitea to help developers 
being productive. The idea behind this project is the missing ALM integration of Gitea in SonarQube. Unfortunately, 
this [won't be added in near future](https://github.com/SonarSource/sonarqube/pull/3248#issuecomment-701334327). 
_Gitea SonarQube Bot_ aims to fill the gap between working on pull requests and being notified on quality changes. 
Luckily, both endpoints have a proper REST API to communicate with each others.

- [Gitea SonarQube Bot](#gitea-sonarqube-bot)
  - [Workflow](#workflow)
  - [Requirements](#requirements)
  - [Bot configuration](#bot-configuration)
  - [Setup](#setup)
    - [SonarQube](#sonarqube)
    - [Gitea](#gitea)
    - [CI system](#ci-system)
  - [TODOs](#todos)
    - [Possible improvements](#possible-improvements)
  - [Contributing](#contributing)
  - [License](#license)
  - [Screenshots](#screenshots)

## Workflow

![Workflow](docs/workflow.png)

**Insights**

- Bot activities
    - extract data from SonarQube
        - [x] Read payload from hook post to receive project,branch/pr,quality-gate
        - [x] Load "api/measures/component"
    - [x] comment PR in Gitea (/repos/{owner}/{repo}/issues/{index}/comments)
    - [x] updates status check (either failing/success)
    - [x] listen on "/sq-bot review" comments
      - [x] comment PR in Gitea (/repos/{owner}/{repo}/issues/{index}/comments)
      - [x] updates status check (either failing/success)

## Requirements

This bot is designed to interact with [SonarQube _Developer_ edition](https://www.sonarsource.com/plans-and-pricing/) and above due to its pull request features. It will most likely work with public SonarCloud because it includes that feature for open source projects.

## Bot configuration

See [config.example.yaml](config/config.example.yaml) for a full configuration specification and description.

## Setup

### SonarQube

- Create a user and grant permissions to "Browse on project" for the desired project
- Create a token for this user that will be used by the bot
- Create a webhook pointing to `https://<bot-url>/hooks/sonarqube`
- Consider securing it with a secret

### Gitea

- Create a user and grant permissions to "Read project" for the desired projects including access to "Pull Requests"
- Create a token for this user that will be used by the bot
- Create a project/organization/system webhook pointing to `https://<bot-url>/hooks/gitea`
- Consider securing the webhook with a secret

### CI system

Some CI systems may emulate a merge and therefore produce another, not yet existing commit hash that is promoted to SonarQube. 
This would cause the bot to fail to set the commit status in Gitea because the webhook sent by SonarQube contains that commit hash. 
To mitigate that situation, the bot will look inside the `properties` object for the key `sonar.analysis.sqbot`. If available, this 
key can contain the actual commit hash to use for updating the status in Gitea.  
See [SonarQube docs](https://docs.sonarqube.org/latest/project-administration/webhooks) for details.

## TODOs

- [ ] Validate configuration on startup
- [ ] Verify webhook secrets
- [ ] Only post status-check (Opt-in/out)
- [ ] Maybe drop `PRBOT_CONFIG_PATH` environment variable in favor of `--config path/to/config.yaml` cli attribute
- [ ] Configure SonarQube PR branch naming pattern for more flexibility (currently focused on Jenkins with [Gitea Plugin](https://github.com/jenkinsci/gitea-plugin))
- [ ] Configuration live reloading
- [ ] _Caching_ of outgoing requests in case the target is not available
- [ ] Parsable logging for monitoring
- [x] Official image for containerized hosting
- [x] Helm chart for Kubernetes
- [ ] Publish Helm chart + docker image
- [x] Respect `"action": "opened"` PR event for updating status check

### Possible improvements

- Reuse existing posted comment for updates via SonarQube webhook or `/sq-bot` comments  
Therefore storing or dynamically retrieving the previous comment id and modify content (/repos/{owner}/{repo}/issues/comments/{id})
- Add more information to posted comment
  - Read "api/project_pull_requests" to get current issue counts and current state
  - Load "api/issues/search" to get detailed information for unresolved issues
- Maybe directly show issues via review comments

## Contributing

Expected workflow is: Fork -> Patch -> Push -> Pull Request

NOTES:

- **Please read and follow the [CONTRIBUTORS GUIDE](CONTRIBUTING.md).**

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for the full license text.

---

## Screenshots

> Bot name and avatar depend on user configuration.

![Comment](./docs/slideshow/comment.png)
![Status](./docs/slideshow/status.png)
