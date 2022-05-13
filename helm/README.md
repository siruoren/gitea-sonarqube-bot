# Gitea SonarQube Bot

## Installation

```bash
helm repo add gitea-sonarqube-bot https://codeberg.org/justusbunsi/gitea-sonarqube-bot/raw/branch/charts/
helm repo update
helm install gitea-sonarqube-bot gitea-sonarqube-bot/gitea-sonarqube-bot
```

You have to modify the `app.configuration` values. Otherwise, the bot won't start as it tries to establish a connection
to your Gitea instance. See [config.example.yaml](https://codeberg.org/justusbunsi/gitea-sonarqube-bot/src/branch/main/config/config.example.yaml)
for full configuration options.

## Parameters

### Common parameters

| Name                 | Description                                                                                  | Value                             |
| -------------------- | -------------------------------------------------------------------------------------------- | --------------------------------- |
| `replicaCount`       | Number of replicas for the bot                                                               | `1`                               |
| `image.repository`   | Image repository                                                                             | `justusbunsi/gitea-sonarqube-bot` |
| `image.pullPolicy`   | Image pull policy                                                                            | `IfNotPresent`                    |
| `image.tag`          | Image tag (Overrides the image tag whose default is the chart `appVersion`)                  | `""`                              |
| `imagePullSecrets`   | Specify docker-registry secret names as an array                                             | `[]`                              |
| `nameOverride`       | String to partially override common.names.fullname template (will maintain the release name) | `""`                              |
| `fullnameOverride`   | String to fully override common.names.fullname template                                      | `""`                              |
| `resources.limits`   | The resources limits for the container                                                       | `{}`                              |
| `resources.requests` | The requested resources for the container                                                    | `{}`                              |
| `nodeSelector`       | Node labels for pod assignment. Evaluated as a template.                                     | `{}`                              |
| `tolerations`        | Tolerations for pod assignment. Evaluated as a template.                                     | `[]`                              |
| `affinity`           | Affinity for pod assignment. Evaluated as a template.                                        | `{}`                              |
| `podAnnotations`     | Pod annotations.                                                                             | `{}`                              |


### App parameters

| Name                                            | Description                                                                                                                               | Value |
| ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ----- |
| `app.configuration.gitea.url`                   | Endpoint of your Gitea instance. Must be expandable by '/api/v1' to form the API base path as shown in Swagger UI.                        | `""`  |
| `app.configuration.gitea.token.value`           | Gitea token as plain text. Can be replaced with `file` key containing path to file.                                                       | `""`  |
| `app.configuration.sonarqube.url`               | Endpoint of your SonarQube instance. Must be expandable by '/api' to form the API base path.                                              | `""`  |
| `app.configuration.sonarqube.token.value`       | SonarQube token as plain text. Can be replaced with `file` key containing path to file.                                                   | `""`  |
| `app.configuration.sonarqube.additionalMetrics` | Setting this option you can extend that default list by your own metrics.                                                                 | `[]`  |
| `app.configuration.projects[0].sonarqube.key`   | Project key inside SonarQube                                                                                                              | `""`  |
| `app.configuration.projects[0].gitea.owner`     | Repository owner inside Gitea                                                                                                             | `""`  |
| `app.configuration.projects[0].gitea.name`      | Repository name inside Gitea                                                                                                              | `""`  |
| `volumes`                                       | If token and webhook secrets shall be provided via file, volumes and volume mounts can be configured to setup the environment accordingly | `[]`  |
| `volumeMounts`                                  | If token and webhook secrets shall be provided via file, volumes and volume mounts can be configured to setup the environment accordingly | `[]`  |


### Security parameters

| Name                                     | Description                                                                                                            | Value  |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------ |
| `serviceAccount.create`                  | Specifies whether a service account should be created                                                                  | `true` |
| `serviceAccount.annotations`             | Annotations to add to the service account                                                                              | `{}`   |
| `serviceAccount.name`                    | The name of the service account to use. If not set and create is true, a name is generated using the fullname template | `""`   |
| `podSecurityContext.fsGroup`             | Group ID for the container                                                                                             | `1000` |
| `securityContext.readOnlyRootFilesystem` | Mounts the container's root filesystem as read-only                                                                    | `true` |
| `securityContext.runAsNonRoot`           | Avoid running as root user                                                                                             | `true` |
| `securityContext.runAsUser`              | User ID for the container                                                                                              | `1000` |


### Traffic exposure parameters

| Name                                 | Description                                                                           | Value                    |
| ------------------------------------ | ------------------------------------------------------------------------------------- | ------------------------ |
| `service.type`                       | Service type                                                                          | `ClusterIP`              |
| `service.port`                       | Service port                                                                          | `80`                     |
| `ingress.enabled`                    | Enable ingress controller resource                                                    | `false`                  |
| `ingress.className`                  | IngressClass that will be be used to implement the Ingress (Kubernetes 1.18+)         | `""`                     |
| `ingress.annotations`                | Additional annotations for the Ingress resource.                                      | `{}`                     |
| `ingress.hosts[0].host`              | Host for the ingress resource                                                         | `sqbot.example.com`      |
| `ingress.hosts[0].paths[0].path`     | The path to the bot endpoint                                                          | `/`                      |
| `ingress.hosts[0].paths[0].pathType` | Ingress path type                                                                     | `ImplementationSpecific` |
| `ingress.tls`                        | The tls configuration for additional hostnames to be covered with configured ingress. | `[]`                     |

