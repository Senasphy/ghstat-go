# ghstat

`ghstat` is an interactive terminal application for exploring a GitHub contribution calendar.
It opens a live view where you can move day by day, inspect totals, and switch year windows without leaving the terminal.

## What You Get

- Contribution grid with keyboard navigation
- Day detail panel with week, month, streak, and best-day stats
- Window switching across available years
- Loading and error states that keep context visible

## Requirements

- Go 1.25 or newer
- A GitHub token in `GITHUB_TOKEN` or passed with `--token`

## Run

```bash
go run . <github-username>
```

## Install from Releases

Prebuilt binaries are published with every release.
Download the archive for your platform from the GitHub Releases page in the **Assets** section.

### Quick Install Scripts

You can install directly from releases using the scripts in this repository.

Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/senasphy/ghstat-go/main/scripts/install.sh | sh
```

Windows PowerShell:

```powershell
iwr -useb https://raw.githubusercontent.com/senasphy/ghstat-go/main/scripts/install.ps1 | iex
```

Install a specific version:

Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/senasphy/ghstat-go/main/scripts/install.sh | sh -s -- v1.2.3
```

Windows PowerShell:

```powershell
$tmp = Join-Path $env:TEMP "ghstat-install.ps1"; iwr -useb https://raw.githubusercontent.com/senasphy/ghstat-go/main/scripts/install.ps1 -OutFile $tmp; & $tmp -Version v1.2.3
```

Run with explicit token:

```bash
go run . --token "$GITHUB_TOKEN" <github-username>
```

Run with a specific end-year window:

```bash
go run . --year 2025 <github-username>
```

## Command Options

- `--year`: End year for the rolling 12-month window
- `--token`: GitHub token (falls back to `GITHUB_TOKEN`)

## Navigation

- `←/h` and `→/l`: previous or next week
- `↑/k` and `↓/j`: move up or down weekday rows
- `g` and `G`: first day or last day in the loaded window
- `0` and `$`: row start or row end
- `H` and `L`: previous or next month
- `[` and `]`: select previous or next available window
- `Enter`: load the selected window
- `t`: jump to today if it exists in the current window
- `?`: toggle help
- `q`, `Esc`, `Ctrl+C`: quit

## Behavior Note

- The selected year chip is only a pending selection until you press `Enter`.

## Verification

Run tests:

```bash
go test ./...
```

## Development

Regenerate GraphQL types after query or schema updates:

```bash
go run github.com/Khan/genqlient genqlient.yaml
```

## Troubleshooting

- If the app cannot load data, verify your token and username first.
- If private contributions are missing, check token permissions and GitHub profile settings.

## Repository Layout

```text
.
├── .github/
│   └── workflows/
│       ├── release-validate.yml  # PR and main snapshot validation
│       └── release.yml           # tag-driven release pipeline
├── cmd/
│   └── ghstat-go/
│       └── main.go               # CLI entrypoint
├── internal/
│   ├── contrib/                  # calendar model, stats, and navigation
│   ├── githubapi/                # GitHub GraphQL client and mapping
│   └── ui/                       # terminal model, keys, styles, rendering
├── queries/                      # GraphQL operations
├── scripts/                      # install scripts for release binaries
├── .goreleaser.yaml              # release packaging and publishing config
├── genqlient.yaml                # GraphQL code generation config
└── schema.graphql                # pinned GraphQL schema
```
