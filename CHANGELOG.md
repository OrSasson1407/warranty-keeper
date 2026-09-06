# Changelog

## v2.0.0 — 2026-09-06

The full v2.0 milestone (see [docs/warranty-tracker-v2-scope.md](docs/warranty-tracker-v2-scope.md)) — 6 of 7 scoped issues shipped and live-verified against real accounts/credentials; the 7th intentionally left open ([#7](https://github.com/OrSasson1407/warranty-keeper/issues/7), see below).

### Features

- **Google Sign-In** ([#6](https://github.com/OrSasson1407/warranty-keeper/issues/6)) — sign in or register with a Google account as an alternative to email/password. Verified end-to-end with a real Google Cloud OAuth client and a real account: new-account creation, household creation, and token issuance all confirmed live.
- **Gmail integration** ([#3](https://github.com/OrSasson1407/warranty-keeper/issues/3)) — optional, revocable Gmail connection that scans for order-confirmation emails from a small retailer allowlist (Amazon, KSP, איקאה) and feeds matches into the existing receipt-confirm flow. Tokens are encrypted at rest. Verified live: real OAuth connection, real encrypted token storage, and a real scan run against a connected inbox with zero errors.
- **Premium / freemium tier** ([#5](https://github.com/OrSasson1407/warranty-keeper/issues/5)) — free tier capped at 20 products; in-app upgrade unlocks unlimited products.
- **Multi-tier expiry notifications** ([#4](https://github.com/OrSasson1407/warranty-keeper/issues/4)) — warnings at 30/14/3 days before expiry (previously a single 30-day warning), plus an annual per-household summary.
- **Advanced search & filtering** ([#8](https://github.com/OrSasson1407/warranty-keeper/issues/8)) — filter the product list by room, category, warranty status, and price range, combinable with the existing text search.
- **Total-cost-of-ownership tracking** ([#9](https://github.com/OrSasson1407/warranty-keeper/issues/9)) — log repair/maintenance costs against a product and see a running total alongside the purchase price.
- **Community warranty-rule correction** ([#7](https://github.com/OrSasson1407/warranty-keeper/issues/7), partial) — a "does this warranty period look wrong?" report action on the product detail screen, landing in a reviewable queue. The other half of #7 — expanding `warranty_rules` coverage — is intentionally deferred: it needs real MVP usage data (which categories most often fall back to the uncertain 12-month default) that doesn't exist yet, so #7 stays open rather than shipping guessed-at categories.

## v1.0.2 — 2026-09-06

Real OCR, several new mobile features, and server hardening — all previously tracked as Backlog items ([#10](https://github.com/OrSasson1407/warranty-keeper/issues/10), [#12](https://github.com/OrSasson1407/warranty-keeper/issues/12)–[#15](https://github.com/OrSasson1407/warranty-keeper/issues/15)).

### Features

- **Real OCR providers** — the OCR stub can now be swapped for a real vision-model-backed provider: `GeminiProvider` (free tier via Google AI Studio, no credit card required) or `AnthropicProvider` (Claude). Gemini takes priority when both are configured. Verified end-to-end against a real receipt image with the free Gemini tier.
- **Dashboard analytics** — a summary card showing total value of covered products, count of warranties expiring within 30 days, and a per-category breakdown, computed client-side from existing data.
- **Offline mode (read-only)** — the dashboard caches the last successful product list and falls back to it with a banner when the network is unavailable.
- **Server-managed manufacturer contacts** — claim-screen contact info now comes from a `manufacturer_contacts` table via the API instead of a static bundled file, with a fallback to the receipt's OCR-parsed vendor when the product's brand field is blank.
- **Calendar sync (tier 1)** — a button on the product detail screen adds a one-time all-day reminder to the device's calendar on the warranty expiry date.

### Hardening

- Push notification sends now retry with backoff (3 attempts) before falling through to the existing next-day retry.
- The OCR call is now bounded by a 20-second timeout so a slow provider can't hang a receipt upload indefinitely.
- OCR failures are now logged with the actual error instead of failing silently.

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
