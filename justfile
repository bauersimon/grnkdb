generate-mocks:
    mockery

format:
    find . -name "*.go" -not -path "./.go/*" -exec go fmt {} \;

install-devenv:
    go install golang.org/x/tools/gopls@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/go-delve/delve/cmd/dlv@latest

install-lint:
    go install github.com/kisielk/errcheck@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install gotest.tools/gotestsum@latest
    go install github.com/vektra/mockery/v2@latest

install: install-devenv install-lint

build:
    go build .

test: generate-mocks
    gotestsum -- -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    rm coverage.out

lint:
    #!/bin/bash
    set -e
    errcheck ./... || EXIT_CODE=1
    staticcheck ./... || EXIT_CODE=1
    exit ${EXIT_CODE:-0}
