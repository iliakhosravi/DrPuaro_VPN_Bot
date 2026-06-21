# AGENTS.md

Guidance for LLM agents working in this repository. Read this before making changes.

## What this is

`techybat.org/go-vpn` (binary name `govpn`) is a Go Telegram bot that sells and manages VPN
configs. It supports two backend VPN panel types:

- **Sanaei / 3x-ui** — managed via the `panel/` package.
- **S-UI** — managed via the `sui/` package.

A MySQL database (via GORM) stores users, orders, packs, and configs. Subscription links are
fetched directly from the VPN panels (Sanaei/S-UI); the bot no longer self-hosts a subscription
endpoint. The bot also runs scheduled notification/maintenance jobs.

## Entry points / CLI

The binary uses `spf13/cobra`. `main.go` just wires subcommands into `cmd.RootCmd` and calls
`cmd.Execute()`.

- `govpn start` (`cmd/bot/start.go`) — loads `.env`, calls `database.Setup()`, builds the
  `go-telegram/bot.Bot`, registers cron jobs, then `b.Start(ctx)`. It no longer starts the
  `sub/` HTTP(S) subscription server — that's deprecated; subscription links are now fetched
  directly from the VPN panels (Sanaei/S-UI), not self-hosted. The `sub/` package and
  `vars.Get("env")` prod/dev split still exist in the repo but are unused dead weight from this
  path; don't reintroduce a call to `sub.ServeHttp`/`ServeHttps` from `start.go` without checking
  whether the deprecation is intentional project-wide.
- `govpn renew --expiring-until YYYY-MM-DD --new-end-date YYYY-MM-DD` (`cmd/renew/renew.go`) —
  batch-renews configs expiring in a date range, branching on `Order.Pack.Type` to call the
  correct panel API, then `Sync`s the config.

When adding a new CLI entry point, follow the existing pattern: a `*cobra.Command` in its own
`cmd/<name>/` package, registered in `main.go`.

## Config / environment variables

- `vars/vars.go` is the single accessor: `vars.Get(name)`. If `USE_ENV_FILE=true` (set in every
  shipped `.env*` file), it reads straight from `os.Getenv`; otherwise it falls back to a
  hardcoded defaults map in the same file.
- **Only `.env` is loaded by code** (`godotenv.Load(".env")` in `cmd/bot/start.go` and
  `cmd/renew/renew.go`, hardcoded filename). `.env.example`, `.env.techy`, `.env.ultra` are
  alternate deployment profiles a human copies to `.env` manually — they are not read directly by
  any Go code.
- `.env`, `.env.techy`, `.env.ultra` contain **live secrets** (Telegram bot tokens, panel
  passwords, DB credentials) and must never be committed or printed. They are gitignored; if you
  see one tracked in git, stop and flag it instead of fixing it silently — check whether history
  needs scrubbing too.
- When adding a new config value: add it to `.env.example` with a placeholder, and read it via
  `vars.Get("NEW_KEY")` — do not read `os.Getenv` directly elsewhere.

## Database

- `database.Setup()` = `database.GetDB()` (singleton MySQL connection via `gorm.io/driver/mysql`,
  pool tuned: 100 max open / 10 idle / 1h lifetime) + `MigrateAll(db)`.
- **No migration files** — every model implements `Migrate(db) error` (just `db.AutoMigrate(&T{})`)
  and `MigrateAll` calls them all. If you add a new model, add it to `database.MigrateAll` and give
  it a `Migrate` method matching the existing `Model` interface in `models/base.go`.
- Key models (`models/`): `Config`, `Order`, `Pack`, `User`, `Category`, `Card`, `Receipt`,
  `Guide`, `ChargeOrder`, `ChargeReceipt`, `InlineKeyboard`.
- `Pack.Type` is the central branch point: `CustomPack`, `SanaeiPack`, `SUIPack` (`models/pack.go`).
  `Pack.Period == 0` means an unlimited-duration pack (see `StringPeriod`, `Order.NormalStr`,
  `Config.SetupSanaei` for how `0` is special-cased — replicate this when touching period logic).
- `Config` (models/config.go) has per-panel-type method variants: `Sync`/`SyncSanaei`/`SyncSUI`,
  `SetupSanaei`/`SetupSUI`, `DepleteSanaei`/`DepleteSUI`, plus type-dispatching wrappers
  (`ShortLink`, `SubLink`, `PanelSubLink`, `Link`, `JSONLink`, `EndDate`, `RemainedTraffic`). When
  adding panel-type-specific behavior, follow this "wrapper dispatches on `Pack.Type`, per-type
  method does the work" pattern rather than adding `if` branches inline everywhere.

## Panel abstraction

- `panel/` (Sanaei/3x-ui): `panel.GetPanel()` singleton client; `Client`/`ClientForm` in
  `panel/client.go`; traffic constants `ONE_GB`/`ONE_MB`; sub-link helpers in `panel/v2ray.go`.
- `sui/` (S-UI): `sui.GetSui()` singleton, token-authenticated `resty` client; `Client`,
  `InitClient`, `GetClientByID`/`GetClientByName`, `NewClient`, `UpdateClient`, `GetSubUrl`,
  `GetShortLinks` in `sui/sui.go` / `sui/client.go`.
- Any feature touching configs/orders almost always needs a branch on
  `order.Pack.Type == models.SanaeiPack` vs `models.SUIPack` (see `crons/config/config.go`,
  `controllers/buy_controller/pack.go`, `controllers/admin/admin.go`, `cmd/renew/renew.go` for
  existing examples). Check both branches when changing shared logic — it's easy to fix one panel
  type and silently break the other.

## Telegram bot dialog framework

Conversational menus use the external library `github.com/sinasadeghi83/go-telegram-bot-ui`
(`dialog` package — not vendored in this repo, check go.sum/pkg cache for its source if you need
exact signatures). Pattern:

```go
nodes := []dialog.Node{
    {ID: "...", Text: "...", Keyboard: [][]dialog.Button{
        {{Text: "label", NodeID: "next-node"}},
        {{ID: "btn-id", Text: "label", CallbackHandler: someHandler, CallbackData: "..."}},
    }},
}
dialog.New(nodes, dialog.Inline(), dialog.WithPrefix(...)).Show(ctx, b, chatID, startNodeID)
```

- `tools/dialog/cat-pack-nodes.go` (`CreateCatPackNodes`) — category → period → pack node tree.
- `widgets/buttonpage/` — paginated list of buttons with prev/close navigation.
- `widgets/form/` — validated text-input forms (validators return `(bool, string)`, e.g.
  `models.PackValidator`, `models.DateValidator`).

Reuse these widgets for new menus instead of hand-rolling keyboards.

## Crons

`crons/config/config.go`: `NotifyAll` = `NotifySanaei` + `NotifySUI`, scheduled every 5 minutes
from `cmd/bot/start.go` via `robfig/cron/v3`. State (which notifications already fired) persists
to `notifs.json` in the working directory via `importNotifs`/`saveNotifs` — not in the database.
Also scheduled: `panel.Setup` every 30m, a backup-to-Telegram-channel job every 1h.

## Build / run

- `Dockerfile`: two-stage build (`golang:1.25.3-alpine` builder → `alpine:latest` runtime),
  entrypoint `/app/govpn start`.
- `docker-compose.yml`: services `db` (mysql:8.0), `bot` (this app), `proxy` (`sagernet/sing-box`,
  config mounted from `./sing-box`).
- No Makefile. Standard Go tooling: `go build ./...`, `go vet ./...`, `go run . start`.

## Testing

There are currently **no `*_test.go` files anywhere in this repo.** If you add non-trivial logic
(date math, traffic/period calculations, panel-type dispatch), prefer adding a test file next to
it rather than assuming "no tests" is the convention to preserve — but don't block a small change
on retrofitting full coverage for an untested package.

## Locale

Most user-facing strings — validation errors, order/pack descriptions, notification messages,
admin notes — are hardcoded **Persian/Farsi** throughout `models/`, `controllers/`,
`tools/dialog/`, and `crons/config/config.go`. When editing these strings:

- Preserve Persian phrasing and tone; don't translate to English unless explicitly asked.
- Be careful with RTL text mixed with `%s`/`%d` format verbs and Latin punctuation — check how
  existing `fmt.Sprintf` calls in the same file order them before changing wording.

## Git / commit conventions

This repo uses [gitmoji](https://gitmoji.dev) prefixes on commit subjects (e.g. `✨`, `🐛`, `🐳`,
`🔥`, `💄`, `🙈`). Match that style for new commits: one emoji, then a concise imperative summary.
Keep commits scoped to one logical change (e.g. don't mix a feature with an unrelated docker
config tweak) — split into multiple commits when a change touches unrelated concerns. Also DO NOT put Co-Authored by this and that at the end of the commit messages.

## Secrets hygiene

Never commit `.env`, `.env.techy`, `.env.ultra`, or any file containing tokens/passwords. If asked
to remove a tracked secret file, `git rm --cached` it (keep the local copy) and ensure it's listed
in `.gitignore` — but note the secret still exists in git history and may need a history rewrite if
the repo is ever made public or shared outside the current trusted group.
