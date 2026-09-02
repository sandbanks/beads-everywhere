# beads everywhere 🐝⚡️

[![Go Report Card](https://goreportcard.com/badge/github.com/sandbanks/beads-everywhere)](https://goreportcard.com/report/github.com/sandbanks/beads-everywhere)
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
* 📝 **One-Click Quick Capture** — File a bead directly into *any* project from the web header without switching directories.
* 🌓 **Zero-FOUC Dark & Light Themes** — Gorgeous amber-slate claymorphic UI designed for fast keyboard-and-mouse triage.
* 🦀 **Universal CLI Integration** — Auto-detects and seamlessly drives `br` or `bd` CLI binaries under the hood.
* 🚀 **Blazing Fast & Lightweight** — Single standalone Go binary with embedded HTML templates and minified Tailwind CSS. Starts in `< 5ms`.

---

## 📦 Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/sandbanks/beads-everywhere/cmd/be@latest
```

### Option 2: Build from Source

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
# List all discovered projects and open issue counts
be projects

# Show all unblocked/ready issues across all repositories
be ready

# Show all open issues across your fleet
be list

# Create an issue in a specific project from anywhere
be create --repo agentic_ssh --title "Add connection retry backoff" --priority 1
```

---

## ⚙️ Configuration

`beads everywhere` works out of the box with zero configuration. If you wish to customize search roots or filter repositories, create `~/.config/beads-everywhere/config.toml`:

```toml
# Search root directories (defaults to ~/projects)
roots = [
    "~/projects",
    "~/work"
]

# Optional whitelist (only show these projects)
# allowed_repos = ["agentic_ssh", "passbook", "sparks"]

# Optional blacklist (hide these projects from UI)
hidden_repos = ["archive", "scratch"]
```

---

## 💖 Sponsoring Sandbanks

`beads everywhere` is built and maintained as independent, sovereign open-source software.

If it keeps your multi-repo workflow organized:

👉 **[Sponsor @sandbanks on GitHub Sponsors](https://github.com/sponsors/sandbanks)**

---

## 📄 License

Dual-licensed under **MIT** and **Apache 2.0**.
