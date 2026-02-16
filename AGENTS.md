# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go backend service (`module myblogx`). Keep business code in package-oriented folders:
- `api/`: HTTP handlers grouped by domain (`article_api`, `user_api`, etc.).
- `router/`: route registration and middleware wiring (`/api` group).
- `service/`: domain services (ES, Redis, email, QQ, river sync).
- `models/`: GORM models, enums, and ES mapping/pipeline metadata.
- `core/` and `conf/`: initialization and config schema.
- `middleware/`, `utils/`, `common/`: shared infrastructure.
- `init/deploy/`: Docker and service bootstrap assets (MySQL, Redis, ES, Kafka).
- `test/`: runnable integration/demo scripts (not a full unit-test suite).

## Build, Test, and Development Commands
Use Go 1.25+.
- `go run . -f settings.yaml`: start API server (default port from config).
- `go build ./...`: compile all packages.
- `go test ./...`: run Go tests (currently limited coverage).
- `go run . -db -f settings.yaml`: run GORM auto-migration.
- `go run . -es -f settings.yaml`: initialize/delete ES index and pipeline (interactive).
- `go run . -t user -s create -f settings.yaml`: create a CLI user.
- `docker compose -f init/deploy/docker-compose.yml up -d`: bring up local dependencies.

## Coding Style & Naming Conventions
Follow standard Go style and always run `gofmt` before commit.
- Use tabs/`gofmt` formatting; keep imports grouped by `go fmt` defaults.
- Package names are lowercase; file names use `snake_case` (for example `article_model.go`).
- Exported symbols use `PascalCase`; internal helpers use `camelCase`.
- Keep handlers thin in `api/`; move business logic to `service/`.

## Testing Guidelines
Prefer table-driven tests with `*_test.go` and `TestXxx` naming near the package under test.
For existing integration scripts, run targeted checks via `go run ./test/<module>/enter.go`.
When adding features, include at least one deterministic test for service/model behavior and cover failure paths.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit-like prefixes: `feat:`, `fix:`, `refactor:` (sometimes scoped, such as `feat-es:`).
- Commit message format: `<type>(optional-scope): short summary`.
- Keep one logical change per commit.
- PRs should include: purpose, key changes, config/migration impact, test evidence (`go test` output or script used), and API examples when endpoints change.

## Security & Configuration Tips
`settings.yaml` contains runtime secrets and external endpoints. Do not commit real credentials for new environments.
Use sanitized/local overrides and keep sensitive values in deployment-specific config.
