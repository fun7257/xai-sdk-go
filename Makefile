.PHONY: proto test vet fmt examples race lint vuln check integration

# Regenerate xai/api/v1 from third_party protos. See docs/PROTO.md.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc
# PROTOC_INCLUDE: extra -I path for well-known types (optional).
#   macOS Homebrew: export PROTOC_INCLUDE=/opt/homebrew/include
#   Linux packages often need nothing extra if protoc finds google/protobuf.
proto:
	@set -e; \
	INCLUDES="-I third_party"; \
	if [ -n "$$PROTOC_INCLUDE" ]; then INCLUDES="$$INCLUDES -I $$PROTOC_INCLUDE"; \
	elif [ -d /opt/homebrew/include ]; then INCLUDES="$$INCLUDES -I /opt/homebrew/include"; \
	elif [ -d /usr/local/include ]; then INCLUDES="$$INCLUDES -I /usr/local/include"; \
	fi; \
	protoc $$INCLUDES \
		--go_out=. --go_opt=module=github.com/fun7257/xai-sdk-go \
		--go-grpc_out=. --go-grpc_opt=module=github.com/fun7257/xai-sdk-go \
		third_party/xai/api/v1/*.proto

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './xai/api/v1/*')

examples:
	go build -o /dev/null ./examples/...

# Race on packages with intentional concurrency.
race:
	go test -race ./files/ ./collections/ ./chat/ ./internal/conn/ ./batch/

# golangci-lint: prefer installed binary (v2 config); fallback matches CI pin.
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; using go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6"; \
		go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./...; \
	fi

vuln:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed; using go run golang.org/x/vuln/cmd/govulncheck@latest"; \
		go run golang.org/x/vuln/cmd/govulncheck@latest ./...; \
	fi

# Full local gate (matches CI intent). Offline only — no live API.
check: vet test examples

# Optional live smoke (requires XAI_API_KEY). Not part of default CI.
# Without a key, tests Skip. See docs/RELEASE.md and integration/.
integration:
	go test -tags=integration ./integration/ -count=1 -v -timeout 5m
