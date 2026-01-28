
.PHONY: all test lint docs verify amux-test

all: verify

test:
	go test ./...

lint:
	go vet ./...

docs:
	go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...

docs-check:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Repo dirty before docs check, please commit changes first"; \
		exit 1; \
	fi
	$(MAKE) docs
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Docs generation produced changes. Please commit updated README.md files."; \
		git status; \
		exit 1; \
	fi

build:
	go build ./cmd/amux
	go build ./cmd/amux-node

amux-test: build
	./amux test

verify: test lint docs-check amux-test
