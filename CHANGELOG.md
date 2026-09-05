# Changelog

## v1.0.1 — 2026-09-05

Small post-release fixes and repo hygiene, no user-facing feature changes.

- Added `cmd/demo-data` to seed/reset a realistic demo dataset, replacing ad hoc curl + manual `psql` cleanup that had no safeguard against targeting the wrong database. Refuses to run unless `APP_ENV=development` and only ever touches its own `[DEMO]`-prefixed household.
- Added ESLint (`eslint-config-expo`) and Prettier to the mobile codebase, wired into CI alongside the existing type-check and test steps.

## v1.0.0 — 2026-09-05

First public release. Full MVP scope (see [`docs/warranty-tracker-mvp-scope.md`](docs/warranty-tracker-mvp-scope.md)) implemented end-to-end and tested.

### Features

- **Auth & households** — email/password signup with JWT access/refresh tokens; two-person households via invite code.
- **Receipts & OCR** — photograph a receipt; OCR stub extracts vendor/date/amount (swappable for Google Vision/Textract).
- **Products** — manual entry or receipt-confirmed, with room and category.
- **Warranty rules engine** — 3-step fallback: exact category+brand → general category → 12-month default (flagged `uncertain`).
- **Claims** — log a warranty claim against a product; manufacturer contact info surfaced per brand.
- **Push notifications** — device registration + a daily job warning on products expiring in exactly 30 days (stubbed via Expo push, swappable for real delivery).
- **Mobile app** — Hebrew RTL UI: onboarding, login/register, dashboard (soonest-to-expire first), product detail, add product / OCR confirm, claim, search, settings.

### Quality

- 110+ Go tests (server) and 123 Jest/RTL tests (mobile), both wired into CI on every push.
- Codecov coverage tracking with `server`/`mobile` flags and a combined badge.
- Custom app icon and branding (WarrantyKeeper), real README screenshots.

### Not included (intentionally out of MVP scope)

Gmail integration, web dashboard, billing/Premium tier, multi-tier notification schedules, repair marketplace, B2B/multi-property support, advanced household permissions, insurance export, full NLP categorization.
