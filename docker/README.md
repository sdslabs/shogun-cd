# Shogun-CD — Docker Deployment
Run the whole stack — Shogun, PostgreSQL database, and the web UI — with Docker Compose.

| Service | Published port | What it is |
|---|---|---|
| `shogun` | none (7007, internal) | The CD engine and HTTP API |
| `db` | none (5432, internal) | PostgreSQL database |
| `ui` | `7006 → 7006` | nginx serving the web UI and proxying `/api` |

## Version note
`TODO : Add docker version advisory after tailscale setup is tested. `

## Quick start

```sh
cd docker
cp .env.sample .env
$EDITOR .env          # replace every placeholder — see the .env reference below
docker compose up -d
docker compose logs -f shogun
```

Then open <http://localhost:7006> and log in with `API_ADMIN_EMAIL` and `API_ADMIN_PASSWORD`.

If `GIT_CREATE_DEPLOY_KEY=true`, the first start generates a deploy key — you must add it to your
repository before Shogun can clone anything (see [The deploy key](#the-deploy-key)).

## The `.env` file

`.env` is the single source of truth for the deployment. It is read twice:

1. by **Docker Compose**, for values it needs itself (the database credentials, the bind-mount path), and
2. by the **container's entrypoint script**, which renders `/etc/shogun/config.yaml` from it on every start.

**You never edit `config.yaml`.** Edit `.env` and restart.

### Database

| Variable | Set it to | Notes |
|---|---|---|
| `DB_POSTGRES_USER` | the database role | Used for **both** the Postgres container and Shogun's config. Set before the first run. |
| `DB_POSTGRES_PASSWORD` | a strong password | Same dual use. Changing it later needs more than an edit — see [Changing the database password](#changing-the-database-password). |
| `DB_POSTGRES_DB` | the database name | Both sides. |
| `DB_HOST` | db/your rds url | Leave as `db` with the bundled container. Set your endpoint (e.g. an RDS host) to use an external database. |
| `DB_PORT` | port number the db is exposed on | The port shogun uses to reach the database. Leave 5432 with the bundled database container; set your external database's port otherwise. |
| `DB_SSL_MODE` | see notes | Leave `disable` for the bundled container. For an external database use one of `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` depending on your use case.|
| `DB_MASTER_KEY` | a long random string | Encrypts the secrets Shogun stores. **Set before the first run and never change it** — changing it makes every stored secret undecryptable. |
| `DB_ENCRYPTION_SALT` | a random string | Same rule as `DB_MASTER_KEY`. |

### API

| Variable | Set it to | Notes |
|---|---|---|
| `API_ADMIN_EMAIL` | your admin email | Seeded **once**, on first start. If the admin already exists this is ignored. |
| `API_ADMIN_PASSWORD` | a strong password | Also only used when seeding. Editing it later does **not** change an existing admin's password. |
| `API_JWT_SECRET` | a long random string | Signs login tokens. **Changing it logs everyone out.** |
| `API_JWT_EXPIRATION_HOURS` | see notes | Token lifetime in hours. There is no refresh token — users re-login when it expires. |
| `API_ENABLE_API_SPEC` | true/false | Serves the OpenAPI docs at `/api/docs/` if set to true. |

### Git

| Variable | Set it to | Notes |
|---|---|---|
| `GIT_CREATE_DEPLOY_KEY` | true | Makes Shogun generate an ed25519 key pair and use it. |
| `GIT_REPO` | your manifest repo url | Use only an ssh based url for the repo. |
| `GIT_BRANCH` | name of branch | The branch of your manifest repo which serves as the source of truth. |
| `GIT_POLLING_INTERVAL` | see notes | Number of seconds between git fetches. |

### Runtime

| Variable | Set it to | Notes |
|---|---|---|
| `DATA_DIR` | host path | The host path where you want shogun to store its data like SMR clone, deploy keys, etc. This path will be bind-mounted into the container.  |
| `DEBUG` | false | `true` adds verbose `[LOG]` lines. Setting it to true is recommended only for troubleshooting purposes. |
| `SHOGUN_HOST` | see notes | Only relevant if hosting shogun and ui on separate machines, provide the ip of the machine hosting shogun over here. Else leave it set to 'shogun' in the case of same machine setup. |
| `SHOGUN_PORT` | see notes | Relevant in the same scenario as `SHOGUN_HOST`, mention the published port on which shogun api is exposed on the shogun host. Else, leave unchanged (set to 7007). |

## Starting a subset of the stack
| You want | Command |
|---|---|
| Everything (default) | `docker compose up -d` |
| Shogun and db only, no web UI (hosting ui and api on separate machines) | `docker compose up -d shogun db` |
| UI only, no db or shogun (hosting ui and api on separate machines) | `docker compose up -d --no-deps ui` |
| No bundled database *(external DB)* | `docker compose up -d --no-deps shogun ui` |
| Shogun only, no UI, no bundled database | `docker compose up -d --no-deps shogun` |

## Accessing the UI and the API

- **Web UI:** <http://localhost:7006>
- **OpenAPI docs:** <http://localhost:7006/api/docs/>

## The deploy key

With `GIT_CREATE_DEPLOY_KEY=true`, Shogun generates an ed25519 key pair on first start:

```
${DATA_DIR}/ssh/deploy_id_ed25519       # private — never share
${DATA_DIR}/ssh/deploy_id_ed25519.pub   # public  — add this to your repo
```

Add the **public** key as a deploy key on the repository in `GIT_REPO` (with write access, if you use
the `mutate` step, since Shogun pushes commits back). The public key is also printed in the Shogun logs
at startup, tagged `[Deploy-Key]`.

Until the key is registered, every git operation fails and the service cannot do anything useful.


## Logging

```sh
docker compose logs -f shogun     # or db / ui, or omit the name for everything
```

Log rotation is configured per container via the `x-logging` anchor at the top of
`docker-compose.yaml`: `max-size: 10m`, `max-file: 3`, so each container's logs are capped at roughly
30 MB. Without this, Docker's default settings make them grow forever which can eventually fill
the disk.
To raise Shogun's own verbosity, set `DEBUG=true` in `.env` and restart. Note that `[ERROR]` and
`[INFO]` lines always print; `DEBUG` only controls the verbose `[LOG]` lines.

## Changing the database password

The `db` container applies `DB_POSTGRES_PASSWORD` **only when it initialises the database**. After
that, the password lives inside the `db-data` volume. So editing `.env` alone does nothing useful —
Shogun will re-render its config with the new value and then fail to authenticate. The full sequence:

```sh
# 1. update DB_POSTGRES_PASSWORD in .env

# 2. make sure the database container is running

# 3. change the password in the running database
# (substitute the values from your .env)
docker compose exec db psql -U shogun -c "ALTER ROLE shogun WITH PASSWORD 'new_password';"

# 4. restart Shogun so it re-renders config.yaml with the new password
docker compose restart shogun
```
  