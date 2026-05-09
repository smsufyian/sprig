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

### Setup
```bash
sprig init                                   # bootstrap: Lima VM (macOS) or btrfs check (Linux)
sprig doctor                                 # diagnose platform issues
```

### Space Lifecycle
```bash
sprig create <name>                          # auto-detect stack, create space
sprig create <name> --from production        # create with prod DB clone
sprig create <name> --from staging           # create from named seed
sprig list                                   # list all spaces across all projects
sprig status <name>                          # status of a single space
sprig start <name>
sprig stop <name>
sprig destroy <name>
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
    detect/            ← stack auto-detection (one file per ecosystem)
    engine/            ← space lifecycle: create, start, stop, destroy
    nix/               ← NixOS expression generator
    platform/          ← PlatformDriver interface + Linux/macOS implementations
    db/
      pull/            ← production snapshot
      seed/            ← manual seed
      synth/           ← synthetic data generation
      snapshot/        ← btrfs subvolume management
    registry/          ← space registry (state.json, ports.json)
    config/            ← sprig.toml parsing and merging
  docs/
    superpowers/
      specs/
        2026-05-09-sprig-design.md
```

---

## Open Questions (deferred)

- **Data anonymisation:** Option to strip/mask PII from production snapshots before storing locally. Planned for a future iteration.
- **LLM-enhanced synthetic data:** An optional pass using an LLM to produce more realistic domain-specific values (product names, realistic addresses). Synthetic engine is designed to accept this as a post-processing step.
- **Space sharing:** Exposing a space's services over the network to teammates (e.g. ngrok-style tunnels). Out of scope for V1.
- **Windows / WSL2:** Adding a `WSL2Driver` implementing the `PlatformDriver` interface is the extension point.
- **btrfs hard requirement (Linux):** If the user's Linux host cannot provide btrfs, V1 will fail with a clear message and setup guide. A plain-copy fallback (slower DB cloning) is a potential V2 addition.
- **Multi-engine DB pull:** V1 implements production pull via pg_dump (PostgreSQL). MySQL and MongoDB pull modes require their own dump/restore drivers behind the same `DBPuller` interface — planned for V2.
