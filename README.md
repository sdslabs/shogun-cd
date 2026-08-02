# Shogun-CD

**A simple, self-hosted continuous delivery tool for Git-based deployments.**

Shogun-CD watches a Git repository and deploys changes to your servers via SSH and SFTP. Pipelines and targets are written in YAML, so your delivery setup stays close to the code and manifests it manages.

Shogun helps one Git repository coordinate deployments across many servers.

> **Status:** Active development. SSH and SFTP deployments are ready to use. Kubernetes support is under development.

## What it does

- Watches a Git repository for changes.
- Runs pipelines when matching files present in the repository change.
- Accepts webhooks from CI systems to run pipelines.
- Copies manifests to servers over SFTP.
- Runs commands on servers over SSH.
- Updates YAML files and pushes the change back to repository.
- Stores secrets encrypted in PostgreSQL.
- Keeps a history of pipeline runs.

## How it works

```text
                                      +-----------------------------+
                                      |  Manifest Git repository  |
                                      | Pipelines / Targets / files |
                                      +-----------------------------+
                                                            |
                                             clone, poll, and index
                                                            |
                            webhook                 v
  +-----------+   + build values  +--------------+
  | CI system |------------------->| Shogun-CD |
  +-----------+                            +--------------+
                                                           |
                                                match a pipeline
                                                           v
   +--------------------+        +-------------------------+
   | encrypted secrets |---->|   run steps in order:   |
   |   (PostgreSQL)    |        | sync / exec / mutate  |
   +--------------------+        +-------------------------+
                                                           |
                                       +-------------+-------------+
                                       |                                       |
                sync / exec     v                                     v    mutate
                               +----------------+           +--------------------+
                               | Your servers   |           | Git repository      |
                               | (SSH / SFTP) |           | (commit & push) |
                               +----------------+           +--------------------+
```

Shogun-CD clones your manifest repository and looks for `Pipeline` and `Target` YAML files to be indexed. When a matching Git change or webhook arrives, it runs the pipeline steps in order.

## Example

This pipeline sends a Compose file to a server and restarts the application when the API changes.

```yaml
apiVersion: shogun.dev/v1
kind: Pipeline
metadata:
  name: deploy-api
  enabled: true
spec:
  triggers:
    - type: git_changes
      paths: ["services/api/**"]
  steps:
    - sync:
        target: prod-api
        files:
          - src: services/api/compose.yaml
            dst: ~/api/docker-compose.yaml
    - exec:
        target: prod-api
        commands:
          - cd ~/api
          - docker compose up -d
```

See [examples](examples/) for pipeline and target definitions.

## Main features

| Feature | Description |
| --- | --- |
| Git triggers | Run pipelines only when selected files change. |
| CI webhooks | Starts a deployment from your CI system. |
| Exec commands | Run deployment commands on a server. |
| SFTP sync | Copy modified files to a server safely from the repo. |
| YAML mutation | Update values in YAML manifests, then commit and push it. |
| Secrets | Store SSH keys and other values encrypted at rest which can be accessed in pipelines. |
| Run history | View pipeline runs and step logs through the API. |

## Get started

1. Setup docker for postgres container or provide RDS.
2. Copy [sample.config.yaml](sample.config.yaml) to `config.yaml` and update it for your environment.
3. Add `Pipeline` and `Target` files to your Git repository.
4. Start Shogun-CD:

   ```sh
   $ go run ./cmd
   ```

For local development, `make start` starts PostgreSQL container and runs the app with Air.

When API docs are enabled, open `/docs/` in your browser.

## Roadmap forward 

- Kubernetes deployments with `apply`
- Target and pipeline listing API
- CLI for shogun
- UI for Shogun
- API keys
- Live pipeline logs

Licensed under [Apache License 2.0](LICENSE). See [NOTICE](NOTICE).
