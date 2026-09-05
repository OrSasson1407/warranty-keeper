# Contributing to WarrantyKeeper

By participating in this project, you're expected to uphold the [Code of Conduct](CODE_OF_CONDUCT.md).

## Scope

Read [`docs/warranty-tracker-mvp-scope.md`](docs/warranty-tracker-mvp-scope.md) first — its "Out of Scope" table lists things that are intentionally deferred (Gmail integration, web dashboard, billing, multi-tier notifications, repair marketplace, B2B/multi-property, advanced permissions, insurance export, full NLP categorization). PRs for those will likely be declined, not because they're bad ideas but because they're out of scope for now.

## Setup

Follow the [Server setup](README.md#server-setup) and [Mobile setup](README.md#mobile-setup) sections of the README to get both halves running locally.

## Making a change

1. Fork the repo and create a branch off `master`.
2. Keep the server (Go) and mobile (React Native/Expo) changes in separate PRs where possible — they're tested and released independently (see the path-filtered CI workflows below).
3. Write or update tests for anything you touch. See [Testing](#testing) below.
4. Open a PR against `master`. CI must pass before it can be merged.

## Testing

**Server** (from `server/`):

```bash
go test ./...
```

Tests use an in-memory SQLite database as a stand-in for Postgres, so no running database is needed. Keep SQL portable between Postgres and SQLite (e.g. use `LOWER(x) LIKE LOWER(?)` instead of Postgres-only `ILIKE`).

**Mobile** (from `mobile/`):

```bash
npx tsc --noEmit
npx jest
```

Mobile tests use `@testing-library/react-native` **pinned to v13.3.3** — don't bump this without checking compatibility with `jest-expo`; v14+ introduces breaking changes that broke the setup used here.

Also run lint and format checks before opening a PR:

```bash
npx eslint .
npx prettier --check .    # or `npm run format` to auto-fix
```

## CI

- [`server-ci.yml`](.github/workflows/server-ci.yml) runs on changes under `server/**`: build, vet, test, coverage upload to Codecov (flag `server`).
- [`mobile-ci.yml`](.github/workflows/mobile-ci.yml) runs on changes under `mobile/**`: `tsc --noEmit`, ESLint, Prettier format check, `jest --coverage`, coverage upload to Codecov (flag `mobile`).

Both can also be triggered manually via `workflow_dispatch`. Coverage badges use Codecov's carryforward config (`codecov.yml`) so a commit that only touches one side doesn't zero out the other's badge.

## Code style

- Go: standard `gofmt`/`go vet` conventions, no linter config beyond that.
- TypeScript: ESLint (`eslint-config-expo`) + Prettier, both enforced in CI. Run `npx eslint .` and `npx prettier --check .` from `mobile/` before opening a PR.
- Commit messages: short, imperative, describe the *why* when it's not obvious from the diff.

## Reporting bugs / requesting features

Use the issue templates — they ask for the information needed to act on a report (area, repro steps, environment for bugs; problem/proposal for features).
