# beads everywhere 🐝⚡️

[![CI](https://github.com/sandbanks/beads-everywhere/actions/workflows/ci.yml/badge.svg)](https://github.com/sandbanks/beads-everywhere/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sandbanks/beads-everywhere.svg)](https://pkg.go.dev/github.com/sandbanks/beads-everywhere)
[![Sponsor](https://img.shields.io/badge/Sponsor-sandbanks-ea4aaa?style=flat-square&logo=githubsponsors&logoColor=white)](https://github.com/sponsors/sandbanks)
[![License](https://img.shields.io/badge/License-MIT%2FApache--2.0-blue.svg)](LICENSE-MIT)

> **A cross-platform web UI & hub for discovering and managing Beads issue databases across your local project folders.**

`beads everywhere` (`be`) gives you a bird's-eye view of all your [Beads (`br` / `bd`)](https://github.com/Dicklesworthstone/beads_rust) task databases across 5, 20, or 50+ local Git repositories.

---

## 🎯 Why `beads everywhere`?

[Beads](https://github.com/Dicklesworthstone/beads_rust) is the ultimate offline-first, Git-tracked issue tracker for developers and AI agents. But when your active projects are spread across dozens of folders, finding open tasks means constantly `cd`-ing around or forgetting where that P0 bug was filed.

`beads everywhere` solves this with a **zero-configuration local dashboard**:

* 🔍 **Automatic Project Discovery** — Scans your workspace (default: `~/projects`) and instantly registers every repo containing a `.beads/` database.
* ⚡️ **Global Task Stream** — View ready, in-progress, and open issues across your entire fleet in a single unified view.
* 🩺 **Fleet Health & Auto-Migration** — Audits workspace integrity and automatically detects and applies pending database schema migrations (`br doctor migrate-schema`) in parallel across every repository.
* 🔄 **Fleet Git Sync** — One-command flush and sync of `.beads/` databases to Git-tracked `issues.jsonl` across all repositories.
* 📝 **One-Click Quick Capture** — File a bead directly into *any* project from the web header without switching directories.
* 🌓 **Zero-FOUC Dark & Light Themes** — Gorgeous amber-slate claymorphic UI designed for fast keyboard-and-mouse triage.
* 🦀 **Universal CLI Integration** — Auto-detects and seamlessly drives `br` or `bd` CLI binaries under the hood.
* 🚀 **Blazing Fast & Lightweight** — Single standalone Go binary with embedded HTML templates and minified Tailwind CSS. Starts in `< 5ms`.

---

## 📦 Installation

### Option 1: Homebrew (macOS & Linux)

```bash
brew install sandbanks/tap/beads_everywhere
```

### Option 2: Go Install

```bash
go install github.com/sandbanks/beads-everywhere/cmd/be@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/sandbanks/beads-everywhere.git
cd beads-everywhere
just install
```

---

## 🚀 Quick Start

Launch the local web dashboard:

```bash
be web
```

Open `http://localhost:8425` in your browser.

To bind to a custom port or IP:

```bash
be web --port 8080 --host 0.0.0.0
```

---

## 💻 CLI Usage

`be` also doubles as a multi-repo command-line tool:

```bash
# Audit fleet health and automatically apply schema migrations across all repos
be doctor

# Alias for doctor: run schema migrations across all active and archive repos
be migrate

# Automatically repair degraded or recoverable workspaces
be doctor --repair  # or: be doctor -r

# Quiet mode: only display workspaces needing attention or migrations
be doctor -q

# List discovered active projects and issue counts
be scan

# Scan all repositories including archive roots
be scan --all

# Show all unblocked/ready issues across all active repositories
be ready

# Show all open issues across your fleet
be list

# Search issues across all repositories
be search "auth"

# Flush beads databases to issues.jsonl and sync Git repos across fleet
be sync

# Create an issue in a specific project from anywhere
be create --repo agentic_ssh --title "Add connection retry backoff" --priority 1
```

---

## ⚙️ Configuration

`beads everywhere` works out of the box with zero configuration. To customize search roots, separate active projects from archives, or filter repositories, create `~/.config/beads-everywhere/config.toml` (or `~/.config/beads-fleet/config.toml`):

```toml
# Active search roots: displayed in Web UI and daily CLI triage
scan_roots = [
    "~/projects",
    "~/.config/nix-config"
]

# Maintenance & archive roots: audited and auto-migrated by `be doctor`,
# but kept out of the daily web view and task streams to avoid clutter
archive_roots = [
    "~/archives",
    "~/bin"
]

# Optional whitelist (only expose these projects)
# allowed_repos = ["agentic_ssh", "passbook", "sparks"]

# Optional blacklist (hide these projects from UI and CLI)
hidden_repos = ["sparks"]

# Directories to ignore during scanning
ignored_dirs = [
    ".git", "node_modules", "target", "vendor",
    ".cache", "tmp", "dist", "build", ".idea",
    ".tokensave", ".doctor", "Library"
]

# Web server port
port = "8425"
```

---

## 💖 Sponsoring Sandbanks

`beads everywhere` is built and maintained as independent, sovereign open-source software.

If it keeps your multi-repo workflow organized:

👉 **[Sponsor @sandbanks on GitHub Sponsors](https://github.com/sponsors/sandbanks)**

---

## 📄 License

Dual-licensed under **MIT** and **Apache 2.0**.
