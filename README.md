# WarrantyKeeper

[![GitHub repo](https://img.shields.io/badge/GitHub-warranty--keeper-181717?logo=github)](https://github.com/OrSasson1407/warranty-keeper)
[![server coverage](https://codecov.io/gh/OrSasson1407/warranty-keeper/branch/master/graph/badge.svg?flag=server)](https://codecov.io/gh/OrSasson1407/warranty-keeper/tree/master/server)
[![mobile coverage](https://codecov.io/gh/OrSasson1407/warranty-keeper/branch/master/graph/badge.svg?flag=mobile)](https://codecov.io/gh/OrSasson1407/warranty-keeper/tree/master/mobile)

Mobile app for tracking home product purchases, receipts, and warranty expiration dates.

## Repo layout

```
server/   Go + Gin API (PostgreSQL via GORM)
mobile/   React Native + Expo app
docs/     Product & architecture docs (source of truth for scope)
```

Planning docs in [`docs/`](docs/):
- [`warranty-tracker-prd.md`](docs/warranty-tracker-prd.md) — full product vision (long-term; not all of it is being built now)
- [`warranty-tracker-architecture.md`](docs/warranty-tracker-architecture.md) — technical architecture
- [`warranty-tracker-ux-flow.md`](docs/warranty-tracker-ux-flow.md) — screen-by-screen UX flow
- [`warranty-tracker-mvp-scope.md`](docs/warranty-tracker-mvp-scope.md) — **what's actually in scope for this build** (see its "In Scope" table)

## Stack

| Layer | Choice |
|---|---|
| Backend | Go + Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Mobile | React Native + Expo |
| Auth | JWT + refresh tokens (email/password only for MVP — see mvp-scope doc) |
| File storage | S3-compatible object storage — stubbed with a local-disk implementation for now |
| OCR | External API (Google Vision / AWS Textract), stubbed behind an interface for now |
| Push | Expo push service (routes to FCM/APNs), stubbed with a log-only sender for now |

## Prerequisites

- Go 1.22+
- Node 20+
- Docker (for local Postgres) — or point `DATABASE_URL` at any Postgres 14+ instance

## Server setup

```bash
cd ..            # repo root
docker compose up -d db      # local Postgres on localhost:5434

cd server
cp .env.example .env
go mod tidy
go run ./cmd/migrate          # create tables
go run ./cmd/seed             # seed ~30 default warranty_rules
go run ./cmd/api
```

The API starts on `http://localhost:8080` by default. Verify it's up:

```bash
curl http://localhost:8080/health
```

Environment variables (see `.env.example`):

| Var | Purpose |
|---|---|
| `PORT` | HTTP port (default `8080`) |
| `APP_ENV` | `development` / `staging` / `production` |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Signing secret for access/refresh tokens |
| `UPLOADS_DIR` | Where receipt images are stored locally (default `./data/uploads`) |
| `PUBLIC_BASE_URL` | Base URL used to build image URLs returned to the client |

### Useful commands

| Command | Does |
|---|---|
| `go run ./cmd/migrate` | Creates/updates all tables (safe to re-run) |
| `go run ./cmd/seed` | Seeds default `warranty_rules` (idempotent, ~32 categories) |
| `go run ./cmd/notify-expiring` | The single scheduled job: pushes a warning for every product expiring in exactly 30 days. Run once a day via cron / Windows Task Scheduler — there's no in-process scheduler. |

### Swapping stubs for the real thing later

- **OCR**: implement `internal/ocr.Provider` for Google Vision/Textract and wire it in `cmd/api/main.go` instead of `ocr.NewStubProvider()`.
- **Object storage**: implement `internal/storage.Store` for S3 and swap `storage.NewLocalStore(...)`.
- **Push**: switch `notify.NewLogSender()` to `notify.NewExpoSender()` in `cmd/notify-expiring/main.go` once you're ready to actually deliver (no credentials needed for Expo's push service).

## Mobile setup

Requirements: the [Expo Go](https://expo.dev/go) app (or a simulator) for testing on device.

```bash
cd mobile
npm install
npm start
```

This opens the Expo dev tools — scan the QR code with Expo Go, press `a`/`i` for an Android/iOS emulator, or `w` for the web preview.

**Pointing the app at the API:** by default the app calls `http://localhost:8080` (or `http://10.0.2.2:8080` on the Android emulator). A physical device can't reach your computer's `localhost` — set `EXPO_PUBLIC_API_URL` to your machine's LAN IP before starting:

```bash
EXPO_PUBLIC_API_URL=http://192.168.1.20:8080 npm start
```

## Project status

All 5 build steps are in place and manually verified end-to-end (backend via curl, mobile via `tsc` + web preview):

1. Repo scaffolding
2. Database schema (GORM models, migrations, warranty_rules seed)
3. Core API (auth, households/invites, receipts + OCR stub, products, warranty rules engine, claims — all household-scoped)
4. Mobile screens (onboarding → add product → OCR confirm → dashboard → product detail → claim, search, settings)
5. Push notifications (device registration + the single 30-day-warning scheduled job)

Not built (intentionally, per mvp-scope doc's Out of Scope table): Gmail integration, web dashboard, Premium/billing, multi-tier notification schedules, repair marketplace, B2B/multi-property, advanced household permissions, insurance export, full NLP categorization.
