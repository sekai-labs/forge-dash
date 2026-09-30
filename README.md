# ForgeDash

The standalone multi-forge terminal dashboard for Pull Requests, Merge Requests, and Issues across GitHub, GitLab, and Codeberg / Forgejo.

## Features

- **Multi-Forge Aggregation**: Unified view combining GitHub (github.com & Enterprise), GitLab (gitlab.com & self-hosted), and Codeberg/Forgejo.
- **Keyboard-Driven**: Fast navigation, section tabs, and action shortcuts (`j`/`k`, `h`/`l`, `Enter`, `d`, `c`, `v`, `m`, `x`, `p`, `/`, `?`).
- **5-Tab Preview Pane**: Inspect Overview, Activity timeline, Commits, CI Checks, and Unified Diffs.
- **Offline Demo Mode**: Test the full dashboard without credentials using `--demo`.

## Installation

### Using `go install`

```bash
go install github.com/forgedash/forgedash/cmd/forgedash@latest
```

Make sure `$HOME/go/bin` is in your `PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### From Source

```bash
git clone https://github.com/forgedash/forgedash.git
cd forgedash
make build
./bin/forgedash
```

## Quick Start

### 1. Demo Mode

```bash
forgedash --demo
```

### 2. Authentication

Generate a Personal Access Token:
- **GitHub**: Scopes `repo`, `read:org`, `notifications`, `user:email` at [github.com/settings/tokens](https://github.com/settings/tokens)
- **GitLab**: Scopes `api`, `read_user` at GitLab Preferences -> Access Tokens
- **Codeberg**: Scopes `issue`, `repository`, `notification` at Codeberg Settings -> Applications

Log in:

```bash
forgedash auth login --provider github --token "<TOKEN>" --name "GitHub Personal"
forgedash auth login --provider gitlab --host gitlab.com --token "<TOKEN>" --name "GitLab Work"
forgedash auth login --provider forgejo --host codeberg.org --token "<TOKEN>" --name "Codeberg OSS"
```

Or set environment variables: `GITHUB_TOKEN`, `GITLAB_TOKEN`, `CODEBERG_TOKEN`.

### 3. Configuration

Generate the default configuration file at `~/.config/forgedash/config.yaml`:

```bash
forgedash config init
```

Example `~/.config/forgedash/config.yaml`:

```yaml
default_view: prs
default_mode: unified
theme: default
vim_mode: true
auto_refresh: true
refresh_interval: 60

preview:
  open: true
  position: right
  width: 50

accounts:
  - id: gh-personal
    name: GitHub Personal
    provider: github
    host: github.com
    auth_ref: github

pr_sections:
  - title: Needs Review
    filters: "is:open review-requested:@me"
    limit: 25
  - title: My PRs
    filters: "is:open author:@me"
    limit: 25
```

### 4. Health Diagnostics

Verify tool dependencies, network access, and stored tokens:

```bash
forgedash doctor
```

## Keybindings

| Key | Action |
|---|---|
| `j` / `k` or `↓` / `↑` | Move selection down / up |
| `g` / `G` or `Home` / `End` | Jump to top / bottom |
| `h` / `l` or `←` / `→` | Previous / next section tab |
| `1` - `5` | Switch directly to section tab index |
| `s` / `Tab` | Cycle views (`PRs` -> `Issues` -> `Notifications`) |
| `p` / `P` | Toggle preview pane / Toggle position (right / bottom) |
| `[` / `]` | Previous / next preview tab |
| `ctrl+d` / `ctrl+u` | Scroll preview pane down / up |
| `Enter` / `o` | Open selected item in browser |
| `y` / `Y` | Copy item number / URL to clipboard |
| `d` | Jump to Diff preview tab |
| `c` | Add comment |
| `v` | Approve PR / MR |
| `m` | Merge PR / MR |
| `x` / `X` | Close / Reopen item |
| `C` / `Space` | Checkout PR branch locally |
| `/` | Filter current section |
| `r` / `R` | Refresh section / Refresh all |
| `ctrl+a` | Switch account scope |
| `?` | Help cheat sheet |
| `q` / `ctrl+c` | Exit |

## License

MIT
