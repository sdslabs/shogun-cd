# AGENTS.md

## Project Overview

Shogun-CD is a self-hosted continuous delivery tool written in Go. It watches a Git repository for YAML-defined `Pipeline` and `Target` manifests, then deploys changes to servers via SSH/SFTP, mutates YAML files and pushes commits back, or triggers pipelines via webhooks. It runs as a single binary with an embedded HTTP API (Gin), a PostgreSQL-backed store (GORM), and a git polling loop.

## Tech Stack

- **Go 1.24** — language and toolchain
- **Gin** (`github.com/gin-gonic/gin`) — HTTP framework
- **GORM** (`gorm.io/gorm`) with PostgreSQL driver — ORM and migrations
- **golang-jwt** (`github.com/golang-jwt/jwt/v5`) — JWT auth
- **golang.org/x/crypto** — bcrypt, SSH, ed25519, PBKDF2
- **pkg/sftp** (`github.com/pkg/sftp`) — SFTP file transfers
- **doublestar** (`github.com/bmatcuk/doublestar/v4`) — glob matching for git trigger paths
- **go.yaml.in/yaml/v3** — YAML parsing with custom unmarshaling (uses `yaml.Node` for mutation)
- **Air** (`.air.toml`) — hot-reload for local development

## Build & Run

```sh
# One-time: start PostgreSQL container, then run with Air
make start

# Or run directly:
go run ./cmd

# Config: copy sample.config.yaml → config.yaml
```

The makefile uses `docker run` for a local PostgreSQL container. The app auto-migrates all models on startup via `db.AutoMigrate()`.

## Directory Structure

```
cmd/                  # Entry point (main.go + helper.go)
  main.go             # initServices() — wires everything together, blocks forever
  helper.go           # seedAdmin()

api/                  # HTTP API layer
  dto/                # Request DTOs (LoginInput, SecretInput, HookInput)
  http/
    apiutils/         # JWT helpers (GenerateToken, ParseToken), context key getters
    controllers/      # Gin handlers (auth, webhooks, secrets, pipelines, system, api-keys)
    middlewares/      # AuthRequired (JWT), VerifyAdmin
    response/         # Structured JSON response helper
    routes.go         # Route registration
    server.go         # StartAPIServer — Gin router + CORS + optional Scalar docs

docs/                 # OpenAPI spec + Scalar UI
  openapi.yaml        # Full API spec
  serve.go            # Embeds spec, serves Scalar HTML at /docs/

examples/             # Sample Pipeline and Target YAML

internal/             # Core application logic
  app/                # App struct — holds all services (currently unused after init)
  config/             # YAML config loading + types (Config, Git, Api, DB)
  encryption/         # AES-256-GCM via PBKDF2 derived key (Encrypt/Decrypt using AEAD)
  git/                # Git clone/pull/poll/commit/push + SSH deploy key generation (ed25519)
  models/             # GORM models: User, Webhook, Secret, PipelineRun, PipelineRunStep
  orchestrator/       # Orchestrates pipelines: indexer, poller, git watcher, RunPipeline
  pipeline/           # Pipeline YAML loader, validator, and execution engine
    steps/            # Step implementations: mutate, sync, exec, apply
  secrets/            # Encrypted secret store (DB-backed) + legacy in-memory impl
  sshclient/          # SSH connection manager (per-pipeline client cache)
  store/              # GORM data access: User, Webhook, Secret, Pipeline stores
  target/             # Target YAML loader and validator
  users/              # User service (bcrypt hashing, role management)
  utils/              # Colored logger with Shogun-styled pipeline log accumulation
  webhooks/           # Webhook CRUD, HMAC/token verification, registry, resolve
```

## Architecture & Key Abstractions

### Interfaces (contracts that matter)

Every service package exposes an interface consumed by other packages:

- **`git.GitService`** — Clone, Pull, CommitAndPushChanges, GetPullEvents, LockRepo/UnlockRepo/RLockRepo/RUnlockRepo
- **`orchestrator.Orchestrator`** — Start, RunIndexer, StartPoller, RunPipeline, PipelineExists
- **`pipeline.PipelineService`** — LoadPipeline, ExecutePipeline
- **`target.TargetService`** — LoadTarget
- **`secrets.SecretService`** — SetSecrets, FindMany, FetchSecret, ResolveSecrets, DeleteSecret
- **`webhooks.WebhookService`** — Create, Resolve, Delete, Find, SetStatus, Load
- **`users.Service`** — Create, MakeAdmin, FindOne, FindMany
- **`encryption.EncryptionService`** — Encrypt, Decrypt
- **`sshclient.SSHManager`** — GetClient, NewClient, CleanupSSHClients
- **`store.UserStore / WebhookStore / SecretStore / PipelineStore`** — DB access interfaces
- **`utils.Logger`** — Logging with severity levels and pipeline log accumulation

### Concurrency Model

- **Orchestrator** uses `atomic.Pointer` for `PipelineMap` and `TargetMap` — atomically swapped after indexing. An `sync.RWMutex` (`mu`) protects indexer vs. pull operations.
- **Git service** uses `sync.RWMutex` (`repo.mu`) for exclusive access during clone/pull/mutate and shared access during indexing.
- **Lock ordering rule**: Pipeline lock MUST be acquired before repo lock to avoid deadlocks (enforced in indexer and mutate steps).
- **Webhook service** uses `sync.RWMutex` on its in-memory registry.
- **SSH Manager** uses a plain `sync.Mutex` — one step executes at a time per pipeline, so no concurrent access.

### Startup Sequence (cmd/main.go → initServices)

1. Logger created
2. Config loaded from `config.yaml`
3. DB connected + auto-migrated (User, Webhook, Secret, PipelineRun, PipelineRunStep)
4. Encryption service initialized (PBKDF2 → AES-256-GCM)
5. Secret service, User service created
6. Git service created (clones repo, generates deploy key if configured)
7. Pipeline service, Target service created
8. Orchestrator started — clones repo → runs indexer → starts poller + git watcher goroutines
9. Admin seeded (ignores if exists)
10. Webhook service created + loaded from DB
11. API server started in a goroutine
12. Main blocks forever (`<-make(chan struct{})`)

### Pipeline Execution Flow

1. **Trigger sources**: Git poller detects file changes → matches `git_changes` triggers via doublestar globs, or webhook resolves with `ci_webhook` trigger.
2. **Orchestrator.RunPipeline** checks existence, launches goroutine.
3. **PipelineService.ExecutePipeline** validates trigger, creates `StepDeps` with SSHManager, iterates steps in order.
4. Each step type (`mutate`, `sync`, `exec`, `apply`) implements the `Step` interface with custom YAML unmarshaling via `StepWrapper.UnmarshalYAML`.
5. **mutate**: Locks repo → reads YAML via `yaml.Node` → traverses `field.path` (supports `[index]`) → writes temp file → renames → commits and pushes.
6. **sync**: Validates server-type target → fetches SSH key from secrets → SFTP client → atomic upload via tmp file rename.
7. **exec**: Validates server-type target → fetches SSH key → runs commands via `sh -s` heredoc with `set -eu`, interpolates `{{VARS}}`.
8. **apply** (Kubernetes): Not yet implemented — stub only.
9. Run results (steps + status) saved to DB via `store.PipelineStore.SavePipelineRunWithSteps`.

### YAML Processing Pattern

Pipeline and Target YAML both use `apiVersion: shogun.dev/v1` with custom unmarshaling. The indexer walks the repo, reads every `.yaml`/`.yml`, checks the `Meta{APIVersion, Kind}` header, then dispatches to `LoadPipeline` or `LoadTarget`. Pipeline steps use a custom `StepWrapper.UnmarshalYAML` that reads the first mapping key as the step type and decodes the value into the appropriate `Step` implementation.

### Mutate Step Field Path Syntax

Dot-separated path with optional bracket indices: `services.api.image` or `containers[0].image`.

### Secret Resolution

Two-level variable interpolation in pipeline steps:
- `{{VAR}}` — resolved from webhook JSON payload (flat key-value map)
- `{{SECRET_NAME}}` — resolved from encrypted DB secrets

### Webhook Authentication

Two auth methods via headers:
- `X-Shogun-Token` — raw secret token (constant-time comparison)
- `X-Shogun-HMAC` — `sha256=<hex>` HMAC-SHA256 signature of request body

### Database Models

| Model | Table | Key Columns |
|---|---|---|
| User | users | id (uuid), email (unique), password_hash, role, is_active |
| Webhook | webhooks | id (uuid), slug (unique), secret (encrypted), pipeline, alias, is_active |
| Secret | secrets | id (uuid), name (unique), value (encrypted) |
| PipelineRun | pipeline_runs | id, pipeline, trigger_kind, success, started_at, finished_at |
| PipelineRunStep | pipeline_run_steps | run_id + step_index (composite PK), step_type, status, logs |

### API Routes

| Group | Method | Path | Auth | Handler |
|---|---|---|---|---|
| auth | POST | /auth/login | — | Login |
| auth | POST | /auth/register | — | Register |
| hook | GET | /hook/:pipeline | JWT | ListWebhooks (filtered) |
| hook | GET | /hook | JWT | ListWebhooks (all) |
| hook | POST | /hook/:slug | — | HandleWebhook |
| admin/hook | POST | /admin/hook | JWT+Admin | CreateWebhook |
| admin/hook | PATCH | /admin/hook/:slug/pause | JWT+Admin | DeactivateWebhook |
| admin/hook | PATCH | /admin/hook/:slug/resume | JWT+Admin | ActivateWebhook |
| admin/hook | DELETE | /admin/hook/:slug | JWT+Admin | DeleteWebhook |
| admin/api-keys | * | /admin/api-keys/* | JWT+Admin | Stub (TODO) |
| admin/secrets | GET | /admin/secrets | JWT+Admin | ListAllSecrets |
| admin/secrets | POST | /admin/secrets | JWT+Admin | SetSecrets |
| admin/secrets | DELETE | /admin/secrets/:secret | JWT+Admin | DeleteSecret |
| system | GET | /system/targets | JWT | Stub (TODO) |
| system | GET | /system/pipelines | JWT | Stub (TODO) |
| pipelines | GET | /pipelines/:pipeline/runs/:run_id | JWT | ListPipelineRuns |
| pipelines | GET | /pipelines/:pipeline/runs | JWT | ListPipelineRuns |
| pipelines | GET | /pipelines/:pipeline/runs/:run_id/logs | JWT | Stub (TODO) |

### Key TODO Items (from codebase)

- Kubernetes `apply` step implementation
- API key management (stubs in `controllers/api-key.go`)
- Target and pipeline listing API (stubs in `controllers/system.go`)
- Live pipeline log streaming
- Server graceful shutdown via context
- Move pipeline enable check outside ExecutePipeline
- Host key verification for SSH
- Tighten CORS allowed origins
- CLI for shogun
- UI for shogun
- Remove legacy `InMemorySecretService` once DB secrets are stable

## Conventions

- **Error sentinel pattern**: Each package defines its own error variables (e.g. `users.ErrUserNotFound`, `webhooks.ErrHookNotFound`, `store.ErrRecordNotFound`).
- **Context propagation**: All DB operations accept `context.Context`, checked with `ctx.Err()` at the top.
- **Defer-based cleanup**: SSH clients closed via `defer deps.SSHManager.CleanupSSHClients()` after each pipeline execution.
- **Constant-based column references**: Every model defines column name constants used in GORM queries (e.g. `PipelineRunColID = "id"`).
- **GORM generic API**: Uses `gorm.G[T]` syntax throughout store layer.
- **YAML node-level mutation**: Mutate steps parse YAML into `*yaml.Node` trees and traverse them, preserving comments and formatting.
- **Atomic file writes**: Both mutate (temp file + rename) and sync (tmp file + rename) use atomic patterns.
- **Debug mode**: `config.debug: true` enables verbose logging and error details in API responses.
