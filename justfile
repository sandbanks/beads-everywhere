# Beads Everywhere (be) Justfile

default:
    @just --list

# Generate templ Go code from .templ files
generate:
    templ generate

# Minify Tailwind CSS
css:
    tailwindcss -i static/css/input.css -o static/css/output.css --minify

# Build be binary
build: generate css
    go build -o be ./cmd/be

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
        launchctl kickstart -k "gui/$(id -u)/org.nix-community.home.beads_fleet" || true
        echo "🔄 beads launchd service restarted"
    elif command -v systemctl >/dev/null 2>&1; then
        systemctl --user restart beads-everywhere || systemctl --user restart beads-fleet || true
        echo "🔄 beads systemd service restarted"
    fi

# Rebuild, install, and restart the live service
deploy: install restart
    @echo "🚀 beads-everywhere deployed successfully to be.flat9.uk"

# Tail service logs
logs:
    tail -f -n 50 "$HOME/logs/beads-everywhere.log" || tail -f -n 50 "$HOME/logs/beads-fleet.log"

# Run locally in development mode
dev: build
    ./be web --port 8425

# Clean build artifacts
clean:
    rm -f be beads-fleet static/css/output.css
