# Virtual Spaces — Service Layer Design

## Goal

sprig creates fully isolated, offline-capable virtual spaces that mirror a developer's production environment. Each space runs real local equivalents of every service the application depends on — database, auth, messaging, cloud — with zero configuration duplication from production.

## Architecture

Five layers, all invisible to the developer:

```
┌─────────────────────────────────────┐
│         Developer's App             │
├─────────────────────────────────────┤
│   Service Layer                     │
│   Postgres · Kafka · LocalStack     │
│   GCP emulators · Clerk mock        │
│   Supabase · custom services        │
├─────────────────────────────────────┤
│   sprig orchestration daemon        │
│   lifecycle · health · env inject   │
├─────────────────────────────────────┤
│   Nix package layer                 │
│   exact reproducible versions       │
├─────────────────────────────────────┤
│   NixOS container (Linux)           │
│   Lima VM → NixOS container (macOS) │
└─────────────────────────────────────┘
```

- **Nix** installs exact reproducible versions of every tool (Postgres 16, Kafka 3.7, etc.)
- **sprig daemon** manages service lifecycle as systemd units — starts them, checks health, injects env vars, manages btrfs snapshots
- **The developer** only ever runs `sprig up`, `sprig down`, `sprig shell` — never sees Nix, NixOS, Lima, or systemd

## Source of Truth

Production config is the single source of truth. sprig reads what already exists in the project — it never asks the developer to redeclare what they've already declared elsewhere.

| Existing file | What sprig learns |
|---|---|
| `docker-compose.yml` | Services, versions, ports |
| `.env` / `.env.example` | External services (CLERK_SECRET_KEY → Clerk mock, AWS_* → LocalStack) |
| `terraform/` / `pulumi/` | AWS/GCP resources to emulate |
| `package.json` | SDKs imported (@supabase/supabase-js, kafkajs, aws-sdk) |
| `go.mod` | SDKs imported (aws-sdk-go-v2, confluentinc/confluent-kafka-go) |

`.sprig.toml` lives at the project root and contains only sprig-specific settings that have no home in existing config:

```toml
[seed]
strategy = "prod"   # "prod" | "manual" | "synthetic"

# Only needed to override an auto-detected value
[overrides.postgres]
version = "15"
```

## Services

### Built-in integrations

| Service | Local implementation | Env vars injected |
|---|---|---|
| **Postgres** | NixOS postgresql service, exact version from docker-compose | `DATABASE_URL` |
| **Supabase** | supabase-cli local stack (Postgres + PostgREST + GoTrue + Studio) | `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY` |
| **Kafka** | Apache Kafka via Nix, version from docker-compose | `KAFKA_BROKERS` |
| **Clerk** | Local OIDC server (dex) mimicking Clerk's API surface | `CLERK_SECRET_KEY`, `CLERK_PUBLISHABLE_KEY`, `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` |
| **AWS** | LocalStack Community (s3, sqs, dynamodb, sns, lambda) | `AWS_ENDPOINT`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| **GCP** | GCP emulators (Pub/Sub, Firestore, Cloud Storage) | `PUBSUB_EMULATOR_HOST`, `FIRESTORE_EMULATOR_HOST`, `STORAGE_EMULATOR_HOST` |

### Extension model — three tiers

**Tier 1: Extra Nix packages**
Any nixpkgs package available inside the space. Declared in `.sprig.toml`:
```toml
[packages]
extra = ["nodejs_20", "python311", "redis"]
```

**Tier 2: Custom service definitions**
Developer defines a service inline. sprig runs it as a systemd unit:
```toml
[[services.custom]]
name    = "payment-mock"
command = "/run/current-system/sw/bin/payment-mock-server"
port    = 9000
env     = { MOCK_MODE = "strict" }
health  = "http://localhost:9000/health"
```

**Tier 3: Plugin registry**
Community-published plugins hosted as a GitHub-indexed registry. Each plugin bundles a service definition + any required Nix packages:
```bash
sprig plugin install kafka-ui
sprig plugin install mailhog
```

## Data Seeding

Three strategies, set once in `.sprig.toml`:

| Strategy | How it works |
|---|---|
| `prod` | sprig connects to the production DATABASE_URL, dumps schema + a safe row-limited subset, imports into the space |
| `manual` | Developer provides SQL files or seed scripts in `sprig/seeds/`, sprig runs them in order |
| `synthetic` | sprig reads the schema (tables, types, constraints, foreign keys) and generates realistic fake data |

## CLI Reference

### Commands

```
sprig init              Scan project, detect services, create .sprig.toml
sprig up                Start the space and all services
sprig down              Stop the space and all services
sprig status            Show running services and health
sprig shell             Open a shell inside the space (env vars injected)
sprig run <cmd>         Run a command inside the space (env vars injected)
sprig logs <service>    Tail logs from a specific service
sprig env               Show all env vars that will be injected
sprig seed              Re-run the data seed strategy
sprig doctor            Diagnose common issues before they cause errors
sprig open <service>    Open a service UI in the browser
sprig update            Update sprig to the latest version
sprig plugin install    Install a plugin from the registry
sprig plugin list       List installed plugins
```

### Multi-space

All commands accept `--space <name>` for running multiple named spaces simultaneously (one per branch, one per developer, etc.):

```bash
sprig up --space feature-payments
sprig run --space feature-payments go test ./...
sprig down --space feature-payments
```

## Terminal Output

### `sprig init`
```
  sprig  Scanning project...

  Detected services

  ●  postgres    16.2    docker-compose.yml
  ●  kafka       3.7.0   docker-compose.yml
  ●  clerk       —       .env.example  (CLERK_SECRET_KEY)
  ●  aws         —       .env.example  (AWS_ACCESS_KEY_ID)
     └─ s3, sqs, dynamodb
  ●  supabase    —       package.json  (@supabase/supabase-js)

  Created .sprig.toml — set your seed strategy and run sprig up
```

### `sprig up`
```
  sprig  Starting space myapp...

  ✓  postgres    ready   1.2s
  ✓  kafka       ready   3.4s
  ✓  localstack  ready   2.1s
  ✓  clerk-mock  ready   0.8s

  Space myapp is ready — run sprig shell or sprig run <cmd>
```

### `sprig down`
```
  sprig  Stopping space myapp...

  ✓  clerk-mock  stopped
  ✓  localstack  stopped
  ✓  kafka       stopped
  ✓  postgres    stopped

  Space myapp stopped.
```

### `sprig status`
```
  sprig  myapp

  SERVICE       STATUS    PORT    UPTIME
  postgres      running   5432    2h 14m
  kafka         running   9092    2h 14m
  localstack    running   4566    2h 14m
  clerk-mock    running   3001    2h 14m
```

### `sprig env`
```
  sprig  Environment — myapp

  DATABASE_URL          postgres://localhost:5432/myapp
  KAFKA_BROKERS         localhost:9092
  AWS_ENDPOINT          http://localhost:4566
  AWS_ACCESS_KEY_ID     test
  AWS_SECRET_ACCESS_KEY test
  CLERK_SECRET_KEY      sk_test_local_xxxxxxxxxxxx
```

### `sprig doctor`
```
  sprig  System check

  ✓  Lima VM      running
  ✓  Nix          installed (2.18)
  ✗  Port 5432    in use by another process
     └─ Fix: kill $(lsof -ti:5432)
  ✓  Disk space   48GB free
```

### Update notifications
Shown at the end of `sprig up` when a newer version is available:
```
  Space myapp is ready — run sprig shell or sprig run <cmd>

  A new version of sprig is available: v0.4.0 → v0.5.0
  Run: sprig update
```

### Errors
Errors always include an actionable fix:
```
  ✗  postgres failed to start
     Port 5432 is already in use.
     Run: kill $(lsof -ti:5432)  or  sprig up --port postgres=5433
```

## UI Libraries

| Feature | Library |
|---|---|
| Spinners during `sprig up` / `sprig logs` | `bubbletea` + `bubbles/spinner` |
| Status tables, env output, doctor results | `lipgloss` |
| Ambiguous auto-detection prompts | `huh` |
| Shell auto-complete | cobra built-in |
| `sprig open` | stdlib `os/exec` (`open` macOS, `xdg-open` Linux) |
| Update check | stdlib `net/http` → GitHub Releases API |

## Implementation Plan Breakdown

This design requires four sequential implementation plans:

1. **Core runtime** — Lima VM setup, NixOS container lifecycle, systemd unit management, btrfs snapshots, `sprig up` / `sprig down` / `sprig shell`
2. **Service integrations** — Postgres, Supabase, Kafka, Clerk, AWS (LocalStack), GCP emulators, env var injection
3. **Auto-detection engine** — docker-compose parser, .env scanner, package.json / go.mod dependency scanner, Terraform reader
4. **Extension system** — extra Nix packages, custom service definitions, plugin registry (index format, install/list commands)
