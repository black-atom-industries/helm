.PHONY: build install clean test coverage dev

BINARY_NAME=helm
INSTALL_DIR=$(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) ./cmd/helm/

install: build
	mkdir -p $(INSTALL_DIR)
	ln -sf $(CURDIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	ln -sf $(CURDIR)/hooks/helm-hook.sh $(INSTALL_DIR)/helm-hook
	@echo "Installed $(BINARY_NAME) and helm-hook to $(INSTALL_DIR) (symlinks)"

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./...

coverage:
	go test ./... -cover

# Development helpers
run: build
	./$(BINARY_NAME)

dev:
	find cmd internal -name '*.go' | entr -r make build

fmt:
	go fmt ./...

lint:
	golangci-lint run

tidy:
	go mod tidy
