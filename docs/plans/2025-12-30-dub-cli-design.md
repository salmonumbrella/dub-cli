# Dub CLI Design

Agent-friendly Go CLI mirroring the full Dub API.

## Decisions

- **Auth**: Browser-based UI (like airwallex-cli) with keyring storage
- **API scope**: Full API coverage (39 endpoints)
- **Repo**: `salmonumbrella/dub-cli`
- **Workspaces**: Multi-workspace support with `--workspace` flag

## Project Structure

```
dub-cli/
├── cmd/dub/main.go
├── internal/
│   ├── api/
│   │   ├── client.go          # HTTP client, retry logic, rate limiting
│   │   ├── endpoints.go       # Endpoint definitions
│   │   └── errors.go          # Error parsing
│   ├── auth/
│   │   ├── server.go          # Browser-based setup server
│   │   └── templates.go       # HTML templates
│   ├── cmd/
│   │   ├── root.go            # Root command, global flags
│   │   ├── auth.go            # login, logout, list, switch, status
│   │   ├── links.go           # CRUD + bulk operations
│   │   ├── analytics.go       # analytics, events
│   │   ├── domains.go         # CRUD + register + check
│   │   ├── partners.go        # partners + partners links + analytics
│   │   ├── customers.go       # CRUD
│   │   ├── commissions.go     # list, update
│   │   ├── track.go           # lead, sale
│   │   ├── tags.go            # CRUD
│   │   ├── folders.go         # CRUD
│   │   ├── version.go
│   │   ├── upgrade.go
│   │   └── completion.go
│   ├── config/
│   │   └── paths.go           # App name, config paths
│   ├── secrets/
│   │   └── store.go           # Keyring-based credential storage
│   ├── outfmt/
│   │   └── format.go          # Output formatting, jq filtering
│   ├── debug/
│   │   └── debug.go           # Debug logging
│   └── ui/
│       └── ui.go              # Terminal colors
├── .goreleaser.yaml
├── .github/workflows/ci.yml
├── Makefile
├── go.mod
└── README.md
```

## API Coverage (39 Endpoints)

### Links (10 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /links | `dub links create` |
| GET | /links | `dub links list` |
| GET | /links/info | `dub links get` |
| GET | /links/count | `dub links count` |
| PATCH | /links/{linkId} | `dub links update` |
| PUT | /links/upsert | `dub links upsert` |
| DELETE | /links/{linkId} | `dub links delete` |
| POST | /links/bulk | `dub links bulk create` |
| PATCH | /links/bulk | `dub links bulk update` |
| DELETE | /links/bulk | `dub links bulk delete` |

### Analytics (2 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| GET | /analytics | `dub analytics` |
| GET | /events | `dub events list` |

### Domains (5 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /domains | `dub domains create` |
| PATCH | /domains/{slug} | `dub domains update` |
| DELETE | /domains/{slug} | `dub domains delete` |
| POST | /domains/register | `dub domains register` |
| GET | /domains/status | `dub domains check` |

### Partners (7 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /partners | `dub partners create` |
| GET | /partners | `dub partners list` |
| POST | /partners/ban | `dub partners ban` |
| POST | /partners/links | `dub partners links create` |
| PUT | /partners/links/upsert | `dub partners links upsert` |
| GET | /partners/links | `dub partners links list` |
| GET | /partners/analytics | `dub partners analytics` |

### Customers (4 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| GET | /customers | `dub customers list` |
| GET | /customers/{id} | `dub customers get` |
| PATCH | /customers/{id} | `dub customers update` |
| DELETE | /customers/{id} | `dub customers delete` |

### Commissions (2 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| GET | /commissions | `dub commissions list` |
| PATCH | /commissions/{id} | `dub commissions update` |

### Conversions (2 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /track/lead | `dub track lead` |
| POST | /track/sale | `dub track sale` |

### Tags (3 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /tags | `dub tags create` |
| GET | /tags | `dub tags list` |
| PATCH | /tags/{id} | `dub tags update` |

### Folders (4 endpoints)
| Method | Path | CLI Command |
|--------|------|-------------|
| POST | /folders | `dub folders create` |
| GET | /folders | `dub folders list` |
| PATCH | /folders/{folderId} | `dub folders update` |
| DELETE | /folders/{folderId} | `dub folders delete` |

## Authentication

### Flow
1. User runs `dub auth login`
2. Local HTTP server starts on random port
3. Browser opens to `http://127.0.0.1:<port>/`
4. User enters workspace name + API key (`dub_xxxxx`)
5. CLI validates against `GET /workspaces` or similar health endpoint
6. On success, stores in keyring: `workspace:<name>` -> `{api_key, created_at}`

### Keyring Storage
```go
type Credentials struct {
    Name      string    `json:"name"`
    APIKey    string    `json:"-"`  // Not serialized in JSON output
    CreatedAt time.Time `json:"created_at"`
}
```

### Multi-Workspace Support
- `dub auth list` - Show all configured workspaces
- `dub auth switch <name>` - Set default workspace
- `dub auth status` - Show current workspace
- `--workspace <name>` flag or `DUB_WORKSPACE` env to override

## Agent-Friendly Features

### Global Flags
```
--workspace, -w    Workspace name (or DUB_WORKSPACE env)
--output, -o       Output format: text|json (default: text, or DUB_OUTPUT env)
--query            JQ filter for JSON output
--yes, -y          Skip confirmation prompts
--force            Alias for --yes
--limit            Limit results (0 = no limit)
--sort-by          Field to sort by
--desc             Sort descending (requires --sort-by)
--debug            Show API requests/responses
--color            Color output: auto|always|never
```

### Machine-Readable Output
- `--output json` returns structured JSON
- `--query '.items[].id'` applies jq filters
- Exit codes: 0=success, 1=error, 2=usage error

### Pagination
All list commands support:
- `--limit N` - Max results
- `--page N` - Page number (where applicable)

## API Client

### Base Configuration
```go
const (
    BaseURL = "https://api.dub.co"
    DefaultTimeout = 30 * time.Second
)
```

### Request Headers
```
Authorization: Bearer dub_xxxxx
Content-Type: application/json
```

### Retry Logic
- 429 (rate limit): Exponential backoff with jitter, respect Retry-After header
- 5xx: Single retry after 1s for idempotent methods (GET, HEAD)
- Circuit breaker: Open after 5 consecutive 5xx errors

### Error Handling
Parse Dub error responses:
```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded",
    "doc_url": "https://dub.co/docs/..."
  }
}
```

## GoReleaser Configuration

- macOS (darwin): CGO_ENABLED=1 for keychain
- Linux/Windows: CGO_ENABLED=0
- Formats: tar.gz (unix), zip (windows)
- Checksums: SHA256

## GitHub Actions

CI workflow:
- Lint: golangci-lint
- Test: go test ./...
- Build: go build ./...
- Release: goreleaser (on tag)

Using latest action versions from https://simonw.github.io/actions-latest/versions.txt

## Dependencies

```
github.com/99designs/keyring   # Secure credential storage
github.com/spf13/cobra         # CLI framework
github.com/itchyny/gojq        # JQ filtering
github.com/muesli/termenv      # Terminal colors
golang.org/x/term              # Terminal detection
golang.org/x/mod               # Version comparison for upgrades
```
