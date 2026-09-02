# Beads Everywhere (be) Justfile

default:
    @just --list

# Minify Tailwind CSS
css:
    tailwindcss -i static/css/input.css -o static/css/output.css --minify

# Build be binary
build: css
    go build -o be ./cmd/be

# Cross-compile for Mac (Apple Silicon), Linux (x86_64, arm64)
cross-build: css
    mkdir -p dist
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/be-darwin-arm64 ./cmd/be
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/be-linux-amd64 ./cmd/be
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/be-linux-arm64 ./cmd/be
    @echo "✅ Cross-compilation complete in dist/"

# Install binary to ~/go/bin and re-sign binary on macOS
install: build
    #!/usr/bin/env bash
    mkdir -p "$HOME/go/bin"
    cp be "$HOME/go/bin/be"
    ln -sfn "$HOME/go/bin/be" "$HOME/go/bin/beads-everywhere"
    if command -v codesign >/dev/null 2>&1; then
        codesign -s - -f "$HOME/go/bin/be" || true
    fi
    echo "🐝 be installed to $HOME/go/bin/be (and $HOME/go/bin/beads-everywhere)"

# Restart background service (launchd on macOS / systemd on Linux)
restart:
    #!/usr/bin/env bash
    if command -v launchctl >/dev/null 2>&1; then
        launchctl kickstart -k "gui/$(id -u)/org.nix-community.home.beads_everywhere" 2>/dev/null || \
        launchctl kickstart -k "gui/$(id -u)/org.nix-community.home.beads_fleet" 2>/dev/null || true
        echo "🔄 beads launchd service restarted"
    elif command -v systemctl >/dev/null 2>&1; then
        systemctl --user restart beads-everywhere 2>/dev/null || \
        systemctl --user restart beads-fleet 2>/dev/null || true
        echo "🔄 beads systemd service restarted"
    fi

# Rebuild, install, and restart the live service
deploy: install restart
    @echo "🚀 beads-everywhere deployed and service restarted successfully"

# Tail service logs
logs:
    tail -f -n 50 "$HOME/logs/beads-everywhere.log" || tail -f -n 50 "$HOME/logs/beads-fleet.log"

# Run locally in development mode
dev: build
    ./be web --port 8425

# Clean build artifacts
clean:
    rm -f be beads-fleet static/css/output.css
