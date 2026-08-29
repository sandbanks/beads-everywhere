.PHONY: all generate css build run install clean

all: generate css build

generate:
	templ generate

css:
	tailwindcss -i static/css/input.css -o static/css/output.css --minify

build: generate css
	go build -o beads-fleet ./cmd/beads-fleet

install: build
	@mkdir -p $(HOME)/.cargo/bin
	cp beads-fleet $(HOME)/.cargo/bin/beads-fleet
	@echo "🌐 beads-fleet installed to $(HOME)/.cargo/bin/beads-fleet"

clean:
	rm -f beads-fleet static/css/output.css
