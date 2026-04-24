<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="logo-dark.png" />
    <img src="logo-light.png" alt="Barb" width="350" />
  </picture>
</p>
<p align="center">
  <a href="https://github.com/travisbale/barb/actions/workflows/ci.yml"><img src="https://github.com/travisbale/barb/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://golang.org/doc/go1.26"><img src="https://img.shields.io/badge/go-1.26-blue?logo=go" alt="Go 1.26" /></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0"><img src="https://img.shields.io/badge/license-GPLv3-green.svg" alt="License: GPL v3" /></a>
</p>

Phishing operations console for [Mirage](https://github.com/travisbale/mirage). Barb handles the operational side of phishing engagements — target lists, email templates, phishlet management, SMTP delivery, and campaign tracking — while Mirage handles the reverse proxy and session capture.

## Architecture

Barb is a single Go binary with an embedded Vue frontend. It communicates with `miraged` over its mTLS API to push phishlets, create lures, and monitor captured sessions in real time.

```txt
┌──────────┐         ┌──────────────────────────┐         ┌──────────────┐
│ Operator │────────▶│          Barb            │────────▶│  SMTP Relay  │
│ Browser  │◀────────│  campaign mgmt, delivery │         └──────┬───────┘
└──────────┘  HTTP   └───────────┬──────────────┘                │
                           mTLS  │                               │ emails
                                 ▼                               ▼
                          ┌─────────────┐                  ┌────────────┐
                          │   miraged   │◀─────────────────│   Target   │
                          │ reverse     │  clicks lure URL │            │
                          │ proxy       │─────────────────▶│            │
                          └─────────────┘  proxied site    └────────────┘
```

## Features

- **Dashboard** — operations overview with campaign stats, active campaign progress, and recent captures
- **Campaigns** — tie targets, templates, SMTP profiles, and phishlets together; configurable send rate; start, cancel, and monitor from the UI
- **Campaign wizard** — guided setup for new campaigns
- **Target lists** — manage recipients manually or import from CSV, with inline renaming
- **Email templates** — compose phishing emails with Go template variables, preview rendered output before sending
- **Phishlet management** — store phishlet YAML configs with a syntax-highlighted editor; automatically pushed to miraged on campaign start
- **SMTP profiles** — configure mail relay servers with encrypted credential storage (AES-256-GCM)
- **Miraged connections** — enroll with miraged instances using invite tokens (automatic keypair generation and mTLS certificate enrollment); configure per-connection notification channels (webhook, Slack) with event-type filtering and one-click test delivery
- **Click tracking** — per-target lure URLs with encrypted tracking parameters for deterministic click attribution
- **Session monitoring** — real-time correlation of miraged session captures to campaign targets via SSE
- **Live updates** — campaign results stream to the UI in real time via server-sent events (no polling)
- **Result export** — download campaign results as CSV for reporting
- **Authentication** — session-based login with mandatory password change on first use
- **Dark/light theme** — terminal-inspired operations console aesthetic with theme toggle

## Quickstart

See the [Quickstart Guide](docs/quickstart.md) to connect Barb to the [Mirage quickstart environment](https://github.com/travisbale/mirage/blob/master/docs/quickstart.md) and run a full campaign end-to-end with the bundled target site.

## Requirements

- Go 1.26+
- Node.js 22+ (for building the frontend)

## Building

```bash
make build
# Produces build/barb
```

## Running

```bash
./barb serve --debug
# Starts on :443 by default
```

Open `https://localhost` in your browser.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:443` | Listen address |
| `--db` | `barb.db` | SQLite database path |
| `--debug` | `false` | Enable debug logging |

On first run, Barb generates an encryption key at `encryption.key` (next to the database file) used to encrypt SMTP passwords at rest.

## Development

Start the full development environment (Mailpit, miraged, Go backend, Vite dev server):

```bash
make dev
```

Then open `http://localhost:5173`. The Vite dev server proxies `/api` requests to the Go backend on `:4443`. Mailpit UI is at `http://localhost:8025`.

Requires Docker for Mailpit and miraged containers. Stop everything with Ctrl+C, or run `make dev-down` to clean up containers.

## Testing

```bash
make test    # all tests (integration + unit)
make unit    # unit tests only
```

Integration tests start the full server in-process with an in-memory SQLite database and a mock mailer.

## License

Barb is licensed under the [GNU General Public License v3.0](LICENSE).
