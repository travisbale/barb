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

Campaign management console for [Mirage](https://github.com/travisbale/mirage). Barb handles the operational side of phishing engagements — target lists, email templates, SMTP delivery, and campaign tracking — while Mirage handles the reverse proxy and session capture.

## Architecture

```
Operator's browser → Barb (campaign management, email delivery)
                        ↓ Mirage API (mTLS)
                     miraged (reverse proxy, session capture)
```

Barb is a single Go binary with an embedded Vue frontend. It communicates with `miraged` over its mTLS API to create lures and monitor captured sessions.

## Requirements

- Go 1.26+
- Node.js 18+ (for building the frontend)

## Building

```bash
make build
# Produces build/barb
```

## Running

```bash
./barb --debug
# Starts on :8080 by default
```

Open `http://localhost:8080` in your browser.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:8080` | Listen address |
| `--db` | `barb.db` | SQLite database path |
| `--debug` | `false` | Enable debug logging |

## Development

Run the Go backend and Vue dev server separately for hot reload:

Terminal 1:
```bash
make dev-backend
```

Terminal 2:
```bash
make dev-frontend
```

Then open `http://localhost:5173`. The Vite dev server proxies `/api` requests to the Go backend.

## Testing

```bash
make test    # all tests
make unit    # unit tests only
```

Integration tests start the full server in-process with an in-memory SQLite database.

## Features

- **Target lists** — manage recipients manually or import from CSV
- **Email templates** — compose phishing emails with HTML and plain text bodies
- **SMTP profiles** — configure mail relay servers for delivery
- **Campaigns** — tie targets, templates, and SMTP profiles together into operations
- **Dark theme** — tactical operations console aesthetic with light mode support via CSS variables

## Project Structure

```
cmd/barb/          # entry point, embeds frontend
internal/
  api/                # HTTP handlers
  phishing/           # domain types, services, validation
  server/             # HTTP server, SPA routing
  store/sqlite/       # SQLite persistence
frontend/
  src/
    api/              # TypeScript API client
    components/       # reusable Vue components
    views/            # page-level views
    composables/      # Vue composables (theme, etc.)
sdk/                  # Go SDK (types, routes, client)
test/                 # integration tests
```

## License

Barb is licensed under the [GNU General Public License v3.0](LICENSE).
