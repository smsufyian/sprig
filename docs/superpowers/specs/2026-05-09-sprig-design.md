# sprig — Design Spec
**Date:** 2026-05-09
**Status:** Approved

---

## Overview

sprig is a Go CLI tool that creates fully isolated virtual spaces for developers, AI coding agents, and CI pipelines. Each space is a complete, self-contained environment: the application stack, all its service dependencies (databases, caches, brokers), and a copy of production data — all isolated from every other space on the machine. The user interacts only with `sprig` commands and an optional minimal `sprig.toml`. No Nix, no containers, no btrfs — none of the underlying machinery is ever exposed.

---

## Goals

- Any developer, AI agent, or CI pipeline can get a fully isolated, production-data-shaped environment in seconds
- Works entirely offline after initial setup
- Runs on macOS and Linux
- Complexity is fully abstracted: the user never touches Nix expressions, container configs, or filesystem primitives
- Extensible: new runtimes and services are self-contained additions
- No runtime reflection anywhere in the codebase — all encoding/decoding is explicit or code-generated

## Non-Goals

- Windows support (out of scope for now; WSL2 path may be added later)
- Cloud-hosted spaces (local-only for now)
- Exposing Nix, btrfs, or container primitives to the user

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    sprig CLI (Go)                       │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ Detector │  │ Override │  │  Space Registry    │   │
│  │ (auto-   │  │ (sprig.  │  │  (~/.sprig/)       │   │
│  │  detect) │  │  toml)   │  │                    │   │
│  └────┬─────┘  └────┬─────┘  └─────────┬──────────┘   │
│       └─────────────┴──────────────────┘               │
│                      │                                  │
│              ┌───────▼────────┐                        │
│              │  Space Engine  │                        │
│              │ (Nix generator │                        │
│              │  + lifecycle)  │                        │
│              └───────┬────────┘                        │
│                      │                                  │
│              ┌───────▼────────┐                        │
│              │   DB Manager   │                        │
│              │ prod/seed/synth│                        │
│              └───────┬────────┘                        │
└──────────────────────┼──────────────────────────────────┘
                       │
          ┌────────────▼────────────────┐
          │       Platform Layer        │
          │                             │
          │  Linux          macOS       │
          │  ───────        ──────────  │
          │  NixOS          Lima VM     │
          │  containers  →  (hidden)    │
          │  directly       then NixOS  │
          │                 containers  │
          └─────────────────────────────┘
                       │
          ┌────────────▼────────────────┐
          │     btrfs host filesystem   │
          │  ~/.sprig/snapshots/        │
          │    base/   ← production DB  │
          │    spaces/<name>/  ← CoW    │
          └─────────────────────────────┘
```

---

## Isolation Model

A space provides four simultaneous layers of isolation:

| Layer | Mechanism | What it means |
|---|---|---|
| **Process** | NixOS container (PID + mount namespaces) | Space cannot see host processes or other spaces |
| **Network** | Private VLAN per space + port-forward table | Ports never collide between spaces |
| **Filesystem** | btrfs CoW snapshot | Writes in one space never affect another |
| **Database** | Independent data dir (CoW clone of base) | Every space has its own full DB copy, instantaneously |

The project directory is **bind-mounted** into the space. Code lives on the host; the developer edits files locally; changes are immediately live inside the space. All services (DB, Redis, Kafka, etc.) run inside the space and are isolated per space.

```
HOST MACHINE
  ~/projects/my-app  (your code, editable locally)
        │
        │ bind-mount
        ▼
  ┌─────────────────────────────────────┐
  │         SPACE: my-feature           │
  │                                     │
  │   App ── Postgres ── Redis ── Kafka │
  │              │                      │
  │       private network               │
  │       172.16.x.0/24                 │
  └──────────────┬──────────────────────┘
                 │ port-forward
  API   → localhost:52100
  DB    → localhost:52101
  Redis → localhost:52102
  Kafka → localhost:52103
```

---

## Platform Layer

### Linux
sprig talks directly to the NixOS container tooling (systemd-nspawn under the hood). btrfs must be available on the host filesystem. `sprig init` checks and provides a clear setup guide if not.

### macOS
sprig manages a single hidden Lima VM running NixOS with a btrfs-formatted disk. The VM is downloaded once (~500MB) on `sprig init` — the only unavoidable network operation. After that, all operations are offline. The VM starts automatically as a background daemon when sprig runs. All container operations are proxied transparently over the Lima SSH socket. The user never interacts with the VM directly.

### Platform Driver Interface
```
PlatformDriver (interface)
  ├── LinuxDriver   → native NixOS container commands
  └── MacOSDriver   → same interface, proxied through Lima SSH
```

New platforms (e.g. WSL2) are added by implementing this interface.

---

## CLI Reference

### Global flags (all commands)
```bash
--output text|json    # default: text. json emits machine-readable JSON for AI agents + scripts
--timeout <duration>  # max time for the command (e.g. 120s, 5m). default: 60s
--debug               # enable verbose logging to ~/.sprig/sprig.log
```

### Setup
```bash
sprig init                                   # bootstrap: Lima VM (macOS) or btrfs check (Linux)
sprig doctor                                 # diagnose platform issues
sprig version                                # print version, commit, build date
sprig update                                 # self-update to latest release (verifies SHA256)
sprig update --version 1.2.0                 # pin to a specific version
sprig completion bash|zsh|fish               # print shell completion script
```

### Space Lifecycle
```bash
sprig create <name>                          # auto-detect stack, create space
sprig create <name> --from production        # create with prod DB clone
sprig create <name> --from staging           # create from named seed
sprig create <name> --timeout 120s           # override default timeout
sprig list                                   # list all spaces across all projects
sprig status <name>                          # status of a single space
sprig start <name>
sprig stop <name>
sprig destroy <name>
sprig prune                                  # interactive cleanup of stopped/stale spaces
sprig prune --days 7                         # auto-remove spaces stopped for >7 days
sprig prune --dry-run                        # show what would be removed without acting
```

### Working in a Space
```bash
sprig shell <name>                           # interactive shell with env vars pre-set
sprig run <name> -- <command>                # run a command inside the space
sprig open <name>                            # print all service URLs and ports
sprig open <name> --service api              # print a specific service URL
```

### Database Operations
```bash
sprig db pull --from <connection-string>     # snapshot production → base
sprig db pull --from <named-connection>      # use a saved named connection
sprig db seed <name> --file ./dump.sql       # manual seed from file
sprig db seed <name> --file ./dump.sql --set-as-base
sprig db generate <name>                     # synthetic data generation
sprig db generate <name> --rows 10000
sprig db reset <name>                        # re-clone space from base
sprig db snapshot <name> --as <label>        # save named snapshot
sprig db restore <name> --from <label>       # restore named snapshot
```

### Configuration
```bash
sprig config edit                            # open sprig.toml in $EDITOR (pre-filled with detected values as comments)
```

### Telemetry
```bash
sprig telemetry status                       # show whether telemetry is on/off + anonymous ID
sprig telemetry disable                      # opt out (writes to ~/.sprig/telemetry.json)
sprig telemetry enable                       # re-enable
```

### CI Mode
```bash
eval $(sprig create ci-$PR_NUMBER --from production --ci)
# exports: SPRIG_SPACE_NAME, SPRIG_API_URL, SPRIG_DB_URL, SPRIG_REDIS_URL, etc.

sprig destroy ci-$PR_NUMBER
```

---

## Stack Auto-Detection

Detection runs in order; later steps enrich earlier ones:

```
Project directory
  │
  ├── go.mod / package.json / pyproject.toml / Cargo.toml / Gemfile / build.gradle
  │     → runtime + version
  │
  ├── docker-compose.yml (read as hints only — Docker is never used)
  │     → service definitions (postgres, redis, kafka, etc.)
  │
  ├── Migration directories
  │     db/migrations, alembic/, prisma/schema.prisma, flyway/
  │     → DB engine + schema location
  │
  ├── .env.example / config.yaml / app.yaml / .env.sample
  │     → required env vars (DATABASE_URL, REDIS_URL, KAFKA_BROKERS, etc.)
  │
  ├── Procfile / Makefile (web:, api:, worker: targets)
  │     → how to start the application
  │
  └── Internal manifest → NixOS module expressions (never shown to user)
```

### Runtime Priority

**V1 (first-class):**
- Python (pip, poetry, pdm, uv)
- Node.js / Bun / Deno (npm, yarn, pnpm, bun)

**V2 (planned extensions):**
- Go, Rust, Java, Kotlin, Ruby

Each runtime is a self-contained detector implementing a `RuntimeDetector` interface. Adding a new runtime does not touch existing code.

### Detected Services (V1)

| Category | Services |
|---|---|
| Databases | PostgreSQL, MySQL, MongoDB, SQLite |
| Cache | Redis, Memcached |
| Message brokers | Kafka, RabbitMQ, NATS |
| Other | Elasticsearch, MinIO, ClickHouse |

All services are treated uniformly: detected, provisioned inside the space, ports allocated from the host port registry, connection strings injected as env vars automatically.

---

## sprig.toml — Override Format

Present only when the user needs to correct or extend detection. `sprig config edit` opens the file pre-filled with detected values as comments so the user knows exactly what was auto-detected.

```toml
[project]
name = "my-app"
run = "python manage.py runserver"    # override app start command

[services.postgres]
version = "15"

[services.redis]
enabled = true

[services.kafka]
version = "3.7"
partitions = 3

[services.elasticsearch]
enabled = false

[services.app]
port = 8080
env = { LOG_LEVEL = "debug", FEATURE_X = "true" }

[db]
migrations = "./db/migrations"        # where to find and auto-run migrations
```

sprig merges detected config with overrides. Only keys present in `sprig.toml` are applied; everything else comes from detection.

---

## Database Layer

### Base Snapshot Model

```
~/.sprig/snapshots/
  base/
    postgres/          ← btrfs subvolume (the source of truth)
  spaces/
    feature-payment/
      postgres/        ← btrfs CoW clone of base (instant, zero-copy)
    ci-pr-456/
      postgres/        ← another CoW clone
  named/
    v1-before-migration/
      postgres/        ← user-named snapshot via `sprig db snapshot`
```

Cloning a 50GB base snapshot into a new space takes ~200ms regardless of data size. Writes in each space stay in that space's CoW layer. `sprig db reset` deletes the space's snapshot and re-clones from base.

### Mode 1 — Production Pull

```bash
sprig db pull --from postgresql://user:pass@prod-host/mydb
sprig db pull --from prod    # named connection from ~/.sprig/connections.toml
```

Internally: pg_dump (or equivalent) → restore into base subvolume → btrfs snapshot. Connection strings stored encrypted (AES-256, key from system keychain). Subsequent pulls update base; running spaces are re-snapshotted lazily on next `sprig db reset`.

### Mode 2 — Manual Seed

```bash
sprig db seed <name> --file ./dump.sql
sprig db seed <name> --file ./dump.sql --set-as-base
```

Accepts `.sql`, `.dump` (pg_dump custom format), `.gz` compressed. `--set-as-base` promotes the imported data to the base snapshot all future spaces clone from.

### Mode 3 — Synthetic Data Generation

```bash
sprig db generate <name>
sprig db generate <name> --rows 10000
```

1. **Parse schema** from migration files or live DB introspection
2. **Build dependency graph** — FK parents before children
3. **Generate data** with type-aware semantic inference:
   - Column `email` → valid email format
   - Column `created_at` → realistic timestamp range
   - `VARCHAR(50) UNIQUE` → guaranteed unique values
   - FK column → valid PK reference from parent table
4. **Insert in constraint-safe order**

Pure Go implementation, no external LLM required. Designed for LLM enhancement later (richer domain-specific values) as an optional layer.

---

## Internal Components

### Space Engine

Per-space directory at `~/.sprig/spaces/<name>/`:

```
~/.sprig/spaces/feature-payment/
  space.nix      ← generated NixOS container config (internal, never shown)
  state.json     ← status, ports, created-at, stack summary
  ports.json     ← host port allocations
```

`space.nix` is generated from merged detected+override config. Re-generated and applied whenever `sprig.toml` changes. Port allocation is deterministic via a host-wide registry at `~/.sprig/ports.json` — no two spaces ever share a port.

### File Locking

Every read-modify-write cycle on `registry.json` and `ports.json` acquires an exclusive `syscall.Flock` file lock before reading and releases it after writing. This prevents corruption when multiple sprig processes run simultaneously (common in CI where several jobs may call `sprig create` in parallel).

```
internal/lock/lock.go   ← WithLock(path string, fn func() error) error
```

Lock files are adjacent to the data files (`.registry.lock`, `.ports.lock`). A process that cannot acquire the lock within 5 seconds fails with a clear error: `"another sprig process is running — retry in a moment"`.

### Signal Handling + Context Threading

All engine operations accept a `context.Context`. The CLI root initialises a cancellable context tied to `SIGINT` / `SIGTERM`:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

When the user presses Ctrl+C during a long `sprig create`:
1. The context is cancelled
2. The engine rolls back: destroys the partially created container, releases allocated ports, removes the registry entry
3. Exits cleanly with a message: `"interrupted — space cleaned up"`

Every engine method signature: `func (e *Engine) Create(ctx context.Context, ...) (*Space, error)`

### Input Validation

Space names are validated before any operation:
- Allowed characters: `[a-z0-9-]`
- Maximum length: 40 characters
- Must start with a letter
- Reserved names rejected: `base`, `default`, `snapshots`

Violation produces a clear error: `error: space name "My Feature" must match [a-z][a-z0-9-]{0,39}`

### Space Registry

`~/.sprig/registry.json` — tracks all spaces across all projects:

```json
{
  "spaces": {
    "feature-payment": {
      "project": "/Users/jane/projects/my-app",
      "status": "running",
      "created": "2026-05-09T10:00:00Z",
      "ports": { "app": 52100, "db": 52101, "redis": 52102 },
      "stack": { "runtime": "python", "services": ["postgres", "redis"] }
    }
  }
}
```

`sprig list` reads this globally — shows all spaces on the machine regardless of current directory.

---

## User Experience Examples

### Human developer
```
$ sprig create feature-payment

  Detected stack:
    runtime   Python 3.12  (poetry)
    database  PostgreSQL 16  (migrations: ./db/migrations)
    cache     Redis 7
    broker    Kafka 3.7

  Creating space "feature-payment"...
    ✓ Database cloned from base snapshot
    ✓ Migrations applied (14 migrations)
    ✓ Services started

  Ready in 6s
    API   →  http://localhost:52100
    DB    →  postgresql://localhost:52101/mydb
    Redis →  redis://localhost:52102
    Kafka →  localhost:52103
```

### AI coding agent
```bash
sprig create agent-task-123 --from production
# SPRIG_API_URL, SPRIG_DB_URL, SPRIG_REDIS_URL injected as env vars
# Agent modifies code, runs tests — all against real prod-shaped data
# No risk to production or other spaces
sprig destroy agent-task-123
```

### CI pipeline (GitHub Actions)
```yaml
- name: Create space
  run: eval $(sprig create ci-${{ github.event.number }} --from production --ci)

- name: Run integration tests
  run: go test ./... -tags integration

- name: Destroy space
  if: always()
  run: sprig destroy ci-${{ github.event.number }}
```

---

## Error Handling

| Scenario | Behaviour |
|---|---|
| btrfs not available (Linux) | `sprig init` detects this, prints clear setup guide. Never silently falls back. |
| Lima VM fails to start (macOS) | Clear error + `sprig doctor` for diagnostics |
| Port conflict | Auto-reassigned from registry; never fails silently |
| Production DB unreachable during pull | Fails cleanly; existing base snapshot untouched |
| Orphaned space after crash | `sprig init` detects on startup, offers cleanup |
| Migration failure during create | Space created but marked `degraded`; DB accessible for debugging; error shown |
| Detection wrong | Every error message suggests `sprig config edit` as the fix |

---

## Logging

Structured logging via `log/slog` (stdlib, Go 1.21+). No external dependency.

- Written to `~/.sprig/sprig.log` — never to stdout/stderr (never pollutes command output)
- `DEBUG` level: off by default; enabled via `SPRIG_DEBUG=1` or `--debug` flag
- `INFO` level: key lifecycle events (space created, container started, migration applied)
- `WARN` level: recoverable issues (port reassigned, migration had warnings)
- `ERROR` level: failures with full context

Sensitive data (connection strings, project paths, space names) is never written to the log. Users can inspect the log with `sprig logs` (tails `~/.sprig/sprig.log`).

---

## Telemetry

**Library:** `github.com/posthog/posthog-go`

Privacy rules (non-negotiable):
- Opt-out by default — disclosed on first `sprig init` with disable instructions
- `SPRIG_NO_TELEMETRY=1` or `sprig telemetry disable` turns it off permanently
- Anonymous device ID generated once, stored in `~/.sprig/telemetry.json` — never tied to identity
- **Never collected:** space names, project paths, connection strings, file contents, any user data

| Event | Properties collected |
|---|---|
| `command_run` | command, subcommand, duration_ms, success, error_type (enum), os, arch, version |
| `space_created` | services (list of names), runtime, seeded (bool), duration_ms |
| `db_operation` | operation, db_engine, duration_ms, success |
| `error` | command, error_type (enum — never the message), os, version |

All events are fired asynchronously in a goroutine and flushed on process exit — zero latency impact on commands.

---

## No-Reflection Policy

No `reflect` package usage anywhere in sprig — not directly, not via library. Rationale: reflection bypasses compile-time type safety, produces opaque runtime errors, and makes code harder to trace.

| Concern | Approach |
|---|---|
| JSON marshaling | `github.com/mailru/easyjson` — code-generated at `go generate` time |
| TOML parsing (sprig.toml) | Hand-rolled recursive descent parser (~200 lines) |
| YAML parsing (docker-compose.yml) | `gopkg.in/yaml.v3` Node API — explicit tree traversal, no struct mapping |
| Synthetic data generation | Explicit `switch` on SQL column types — no Go struct introspection |
| Test assertions | stdlib `testing` only — no testify, no go-cmp |

---

## Dependencies

```
# Runtime
github.com/spf13/cobra              v1.8.1   CLI framework
github.com/mailru/easyjson          v0.7     Reflection-free JSON (code-generated)
gopkg.in/yaml.v3                    v3       YAML Node API (docker-compose hints)
github.com/jackc/pgx/v5             v5       PostgreSQL driver (schema introspection + pull)
github.com/zalando/go-keyring       v0.2     OS keychain for production DB credentials
github.com/posthog/posthog-go       v1       Usage telemetry

# Hand-rolled (no library)
sprig.toml parser                            ~200 lines, replaces TOML library
SQL schema data generator                    Explicit type-switch, replaces gofakeit

# Test (no reflection)
stdlib testing only

# Dev tooling (not in go.mod)
golangci-lint   v1.57    Linting
goreleaser      v2       Cross-platform releases + Homebrew tap
easyjson        v0.7     go generate tool for JSON code generation
```

---

## Testing Strategy

| Layer | Approach |
|---|---|
| Stack detector | Unit tests per ecosystem: fixture project dirs → assert detected manifest |
| Nix generator | Golden file tests: manifest → expected `.nix` output |
| DB synthetic engine | Property tests: generated data must satisfy all schema constraints |
| Space lifecycle | Integration tests against real Linux host (CI runs on Linux; macOS Lima layer skipped in CI) |
| CLI commands | End-to-end tests using a fixture project |
| Platform driver | Interface-level mocks for unit tests; real driver in integration tests |

---

## Directory Structure (Go project)

```
sprig/
  cmd/
    sprig/
      main.go
  internal/
    cli/               ← cobra command definitions
      db/              ← sprig db subcommand group
    detect/            ← stack auto-detection (one file per ecosystem)
    engine/            ← space lifecycle: create, start, stop, destroy (context-threaded)
    nix/               ← NixOS expression generator
    platform/          ← PlatformDriver interface + Linux/macOS implementations
    db/
      pull/            ← production snapshot
      seed/            ← manual seed
      synth/           ← synthetic data generation (explicit SQL type switch)
      snapshot/        ← btrfs subvolume management
    registry/          ← space registry (registry.json, ports.json)
    config/            ← hand-rolled sprig.toml parser and merger
    lock/              ← syscall.Flock file locking for registry + ports
    log/               ← slog initialisation and helpers
    telemetry/         ← PostHog client wrapper (async, opt-out)
    validate/          ← space name validation
    version/           ← version, commit, build date (set via ldflags)
  .github/
    workflows/
      ci.yml           ← test + lint on every PR
      release.yml      ← goreleaser on git tag v*
      nightly.yml      ← full integration suite on cron
  .goreleaser.yml      ← cross-platform builds + Homebrew tap
  .golangci.yml        ← linter config
  Makefile             ← build, test, lint, generate targets
  docs/
    superpowers/
      specs/
        2026-05-09-sprig-design.md
      plans/
        2026-05-09-plan1-foundation.md
```

---

## CI/CD + Release Pipeline

### GitHub Actions workflows

| Workflow | Trigger | Steps |
|---|---|---|
| `ci.yml` | Every PR + push to master | `go vet`, `golangci-lint`, `go test -race -coverprofile`, build for linux-amd64 + darwin-arm64 |
| `release.yml` | `git tag v*` | `goreleaser release` → binaries, checksums, GitHub Release, Homebrew tap update |
| `nightly.yml` | Cron 02:00 UTC | Full integration suite (including Lima VM on macOS runner) |

### GoReleaser

Produces on every tagged release:
- `sprig_<version>_linux_amd64.tar.gz`
- `sprig_<version>_linux_arm64.tar.gz`
- `sprig_<version>_darwin_amd64.tar.gz`
- `sprig_<version>_darwin_arm64.tar.gz`
- `checksums.txt` (SHA256)
- GitHub Release with auto-generated changelog (filters `docs:`, `test:`, `ci:` commits)
- Homebrew formula updated in `github.com/smsufyian/homebrew-tap`

### Makefile targets

```makefile
build:     go build -ldflags "$(LDFLAGS)" -o dist/sprig ./cmd/sprig
test:      go test -race -coverprofile=coverage.out ./...
lint:      golangci-lint run
generate:  go generate ./...          # runs easyjson code generation
clean:     rm -rf dist/ coverage.out
install:   go install -ldflags "$(LDFLAGS)" ./cmd/sprig
```

---

## Open Questions (deferred)

- **Data anonymisation:** Option to strip/mask PII from production snapshots before storing locally. Planned for a future iteration.
- **LLM-enhanced synthetic data:** An optional pass using an LLM to produce more realistic domain-specific values (product names, realistic addresses). Synthetic engine is designed to accept this as a post-processing step.
- **Space sharing:** Exposing a space's services over the network to teammates (e.g. ngrok-style tunnels). Out of scope for V1.
- **Windows / WSL2:** Adding a `WSL2Driver` implementing the `PlatformDriver` interface is the extension point.
- **btrfs hard requirement (Linux):** If the user's Linux host cannot provide btrfs, V1 will fail with a clear message and setup guide. A plain-copy fallback (slower DB cloning) is a potential V2 addition.
- **Multi-engine DB pull:** V1 implements production pull via pg_dump (PostgreSQL). MySQL and MongoDB pull modes require their own dump/restore drivers behind the same `DBPuller` interface — planned for V2.
