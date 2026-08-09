# AGENTS.md

Go 1.26 app (module `github.com/fmotalleb/north_outage`) that scrapes electricity-outage data for Mazandaran (Iran), serves a React/Persian UI, and notifies via Telegram.

## Commands

All tooling is driven by the Makefile. `make ci` runs the whole pipeline: `mod gen build spell lint test vuln diff` — `diff` fails if the tree is dirty, so it doubles as a clean-tree gate.

- `make gen` — `go generate ./...`, which builds the frontend (`//go:generate pnpm ...` in `web/front/front.go`) and needs Node + pnpm.
- `make build` — `go tool goreleaser build --clean --single-target --snapshot`. Goreleaser's `before.hooks` also run `go generate ./...`.
- `make test` — `go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...`, then writes `coverage.html`. Use `-race` only runs with CGO enabled (default).
- `make spell` — misspell with `-w`, it rewrites `**/*.md` in place; re-running it should be a no-op.
- `make lint` is currently broken as written: the Makefile calls `go tool golangci-lint`, but golangci-lint is NOT in the `tool (...)` block of `go.mod`, and `.golangci.yml` uses the v2 config schema. Install golangci-lint v2 yourself if you need to lint.
- Focused verification: `go test ./telegram/handlers/...` (the only Go test package). `go vet ./...` works.

## Frontend gotchas

- The web UI lives in `web/front/` (Vite + React + Tailwind, Persian/RTL). Package manager is **pnpm** (CI and `go:generate` use it, despite `package.json` also having npm scripts).
- The build is inlined into a single `dist/index.html` (`vite-plugin-singlefile`) and embedded via `//go:embed dist/*` in `web/front/front.go`. Only `dist/.gitkeep` is committed — on a fresh clone the embedded FS is empty, so `go run .` serves nothing at `/` until you run `make gen` (or `cd web/front && pnpm i && pnpm build`).
- Frontend smoke tests are plain node scripts (`web/front/smoke-*.js`), not wired into the Makefile.

## Architecture

- Entry: `main.go` → `cmd.Execute()` (cobra) → `service.Serve(ctx)`. One shared context derived from OS signals is created in `cmd/root.go` and passed everywhere; services must not create their own cancellable contexts — a failing service reports to `errCh` and exits the process instead of cancelling the shared context (see the comments in `cmd/root.go` and `service/service.go`).
- Config: `config/reader.go` merges TOML config file → env vars → struct defaults. Env names come from struct tags (`HTTP_LISTEN`, `DATABASE`, `TELEGRAM_BOT`, ...). `.env` is autoloaded by godotenv. `collector` settings also live in the same Config (`collector.endpoint` etc.), not in a separate scrapper-go pipeline file — the README's `example/collector.toml`/scrapper-go sections describe a legacy approach.
- Collector: plain HTTP POST to `collector.endpoint` (default `https://khamooshi.maztozi.ir/api/outages`); request body is built from a template in `collector/api.go` (`internal/template` provides `jFormat`/`faNum`/etc. for Persian/Jalali dates). `ssl_verify` defaults to false on purpose (the site's TLS is currently broken). City IDs are mapped to Persian names in `defaultCityMap`.
- Storage: GORM. Both sqlite and postgres dialects are registered in `database/driver/*` init()s; the `//# go:build orm-sqlite` lines are plain comments, NOT real build constraints — all dialects are always compiled. SQLite (default `sqlite:///outage.db`) needs cgo.
- Web: Echo, routes registered via `init()` + `RegisterEndpoint` (see `web/root.go`). Endpoints: `/api/events`, `/api/updated_at`, `/api/up`, and the embedded frontend at `/`.
- Notifications: events flow collector → channel → `eventToNotificationTransformer` → broadcaster → Telegram subscribers. Integrations only start when their credentials are set (`telegram.key`/`TELEGRAM_BOT`).
- `memory.Memory[T]` is a small TTL cache (Pop/Put) used by bot handlers.

## Conventions

- All user-facing strings (UI, bot replies, city names) are Persian; dates use the Jalali calendar (`go-jalali`) and `Asia/Tehran`.
- `models.Event` and friends are AutoMigrated on startup; don't add migrations manually.
- Use conventional commit messages (`feat:`, `fix:`, `chore:`, `docs:`). Commit after every change before suggesting follow-ups.
