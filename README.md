# 🔗 Dub CLI — Link management in your terminal.

Dub in your terminal. Create short links, view analytics, generate QR codes, manage domains, tags, folders, and more.

## Features

- **Authentication** - browser-based login or API key via environment variable
- **Links** - create, update, delete, and bulk manage short links
- **Analytics** - view clicks, leads, and sales with filtering and grouping
- **QR Codes** - generate customizable QR codes for any URL
- **Domains** - manage custom domains for branded links
- **Tags & Folders** - organize links with tags and folder hierarchies
- **Partners** - manage affiliate partners, referral links, and commissions
- **Multiple workspaces** - manage multiple Dub workspaces
- **Webhooks** - configure webhook endpoints for events

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/dub-cli
```

### From Source

```bash
go install github.com/salmonumbrella/dub-cli/cmd/dub@latest
```

## Quick Start

### 1. Authenticate

Choose one of two methods:

**Browser:**
```bash
dub auth login
```

**Environment variable:**
```bash
export DUB_API_KEY=dub_xxxx
```

### 2. Test Authentication

```bash
dub auth status
```

### 3. Create Your First Link

```bash
dub links create --url https://example.com --key my-link
```

## Configuration

### Workspace Selection

Specify the workspace using either a flag or environment variable:

```bash
# Via flag
dub links list --workspace my-workspace

# Via environment
export DUB_WORKSPACE=my-workspace
dub links list
```

### Environment Variables

- `DUB_API_KEY` - API key for authentication (bypasses browser login)
- `DUB_WORKSPACE` - Default workspace name to use
- `DUB_OUTPUT` - Output format: `text` (default) or `json`

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Rate Limiting

The Dub API enforces rate limits to ensure service stability. The CLI automatically handles rate limiting with:

- **Exponential backoff** - Retries with increasing delays plus jitter
- **Retry-After header respect** - Honors the API's suggested retry timing
- **Maximum retry attempts** - Up to 3 retries on 429 (Too Many Requests) responses
- **Circuit breaker** - After 5 consecutive server errors (5xx), requests are blocked for 30 seconds

## Commands

### Authentication

```bash
dub auth login                 # Authenticate via browser
dub auth logout <workspace>    # Remove workspace credentials
dub auth list                  # List configured workspaces
dub auth switch <workspace>    # Set default workspace
dub auth status                # Show authentication status
```

### Links

```bash
dub links create --url <url> [--key <key>] [--domain <domain>]
dub links list [--search <query>] [--domain <domain>]
dub links get --id <id> | --domain <domain> --key <key>
dub links count
dub links update --id <id> [--url <url>] [--key <key>]
dub links upsert --url <url> [--key <key>] [--domain <domain>]
dub links delete --id <id>

# Bulk operations (read JSON from stdin)
dub links bulk create < links.json
dub links bulk update < updates.json
dub links bulk delete < ids.json
```

### Analytics

```bash
dub analytics [--event <type>] [--group-by <property>] [--interval <interval>] \
              [--domain <domain>] [--link-id <id>] [--start <date>] [--end <date>] \
              [--country <code>] [--city <city>] [--device <type>] [--browser <browser>] \
              [--os <os>] [--referer <referer>] [--timezone <tz>]
```

**Event types:** `clicks`, `leads`, `sales`

**Group by:** `count`, `timeseries`, `countries`, `cities`, `devices`, `browsers`, `os`, `referers`

**Intervals:** `1h`, `24h`, `7d`, `30d`, `90d`, `all`

### Events

```bash
dub events list [--event <type>] [--domain <domain>] [--link-id <id>] \
                [--interval <interval>] [--start <date>] [--end <date>] \
                [--country <code>] [--city <city>] [--device <type>] \
                [--browser <browser>] [--os <os>] [--referer <referer>] [--page <n>]
```

### Domains

```bash
dub domains create --slug <domain> [--placeholder <url>] [--expired-url <url>] [--archived]
dub domains list [--archived] [--search <query>] [--page <n>]
dub domains update --slug <domain> [--placeholder <url>] [--expired-url <url>] [--archived]
dub domains delete --slug <domain>
dub domains register --domain <domain>
dub domains check --slug <domain>
```

### Tags

```bash
dub tags create --name <name> [--color <color>]
dub tags list [--search <query>] [--page <n>]
dub tags update --id <id> [--name <name>] [--color <color>]
```

### Folders

```bash
dub folders create --name <name> [--parent-id <id>]
dub folders list [--search <query>] [--page <n>]
dub folders update --id <id> [--name <name>] [--parent-id <id>]
dub folders delete --id <id>
```

### Partners

```bash
# Partner management
dub partners create --program-id <id> --email <email> [--name <name>]
dub partners list --program-id <id> [--search <query>] [--status <status>]
dub partners ban --program-id <id> --partner-id <id> [--reason <reason>]

# Partner links
dub partners links create --program-id <id> --partner-id <id> --url <url>
dub partners links upsert --program-id <id> --partner-id <id> --url <url>
dub partners links list --program-id <id> [--partner-id <id>]

# Partner analytics
dub partners analytics --program-id <id> [--partner-id <id>] [--interval <interval>]
```

### Customers

```bash
dub customers list [--search <query>] [--page <n>]
dub customers get --id <id>
dub customers update --id <id> [--name <name>] [--email <email>]
dub customers delete --id <id>
```

### Commissions

```bash
dub commissions list --program-id <id> [--partner-id <id>] [--status <status>]
dub commissions update --id <id> [--status <status>] [--amount <amount>]
```

### Conversion Tracking

```bash
# Track leads
dub track lead --click-id <id> --event-name <name> \
               [--external-id <id>] [--customer-id <id>] [--email <email>]

# Track sales
dub track sale --click-id <id> --event-name <name> --amount <value> \
               [--external-id <id>] [--customer-id <id>] [--currency <code>]
```

### Workspaces

```bash
dub workspaces get --id <id>
dub workspaces update --id <id> [--name <name>] [--slug <slug>]
```

### QR Codes

```bash
dub qr --url <url> [--size <pixels>] [--level <L|M|Q|H>] \
        [--fg-color <hex>] [--bg-color <hex>] [-O <file>]
```

> **Note:** Use `-O` (capital O) to save to a file. Output is PNG format.

### Embed Tokens

```bash
dub embed create-referral-token --program-id <id> --partner-id <id>
```

## Output Formats

### Text

Human-readable output with formatting:

```bash
$ dub links list --limit 3
ID                              KEY           URL                    CLICKS
link_abc123...                  my-link       https://example.com    42
link_def456...                  promo         https://sale.com       128
```

### JSON

Machine-readable output:

```bash
$ dub links list --limit 1 --output json
[
  {
    "id": "link_abc123",
    "key": "my-link",
    "url": "https://example.com",
    "clicks": 42
  }
]
```

Data goes to stdout, errors to stderr for clean piping.

## Examples

### Create a branded short link

```bash
dub links create \
  --url "https://example.com/long-page-path" \
  --key "launch" \
  --domain "brand.link"
```

### View analytics by country

```bash
dub analytics --event clicks --group-by countries --interval 30d
```

### Generate a custom QR code

```bash
dub qr --url "https://dub.sh/my-link" \
  --size 800 \
  --level H \
  --fg-color 1a1a1a \
  --bg-color ffffff \
  -O branded-qr.png
```

### Bulk create links from JSON

```bash
echo '[{"url":"https://a.com"},{"url":"https://b.com"}]' | dub links bulk create
```

### Pipeline: get all link IDs

```bash
dub links list --output json --query '.[].id'
```

### JQ Filtering

Filter JSON output with JQ expressions:

```bash
# Get links with zero clicks
dub links list --output json --query '[.[] | select(.clicks == 0)]'

# Extract short links only
dub links list --output json --query '.[].shortLink'

# Get total click count
dub analytics --output json --query '.clicks'
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
dub --debug links list
```

## Global Flags

All commands support these flags:

- `--workspace <name>`, `-w` - Workspace to use (overrides DUB_WORKSPACE)
- `--output <format>`, `-o` - Output format: `text` or `json` (default: text)
- `--query <expr>` - JQ filter expression for JSON output
- `--yes`, `-y` - Skip confirmation prompts
- `--force` - Alias for `--yes`
- `--limit <n>` - Limit number of results returned
- `--sort-by <field>` - Sort results by field name
- `--desc` - Sort descending (requires `--sort-by`)
- `--page <n>` - Page number for pagination
- `--debug` - Enable debug output
- `--color <mode>` - Color mode: `auto`, `always`, or `never`
- `--help` - Show help for any command

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
dub completion bash > $(brew --prefix)/etc/bash_completion.d/dub

# Linux:
dub completion bash > /etc/bash_completion.d/dub

# Or source directly:
source <(dub completion bash)
```

### Zsh

```zsh
# Save to fpath:
dub completion zsh > "${fpath[1]}/_dub"

# Or add to .zshrc:
echo 'source <(dub completion zsh)' >> ~/.zshrc
```

### Fish

```fish
dub completion fish > ~/.config/fish/completions/dub.fish
```

### PowerShell

```powershell
dub completion powershell | Out-String | Invoke-Expression
```

## Development

After cloning, install git hooks:

```bash
make setup
```

## License

MIT

## Links

- [Dub API Documentation](https://dub.co/docs/api-reference)
