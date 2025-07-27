test:
    go test -v ./...

format:
    find . -name "*.go" -not -path "./.go/*" -exec go fmt {} \;

install-devenv:
    go install golang.org/x/tools/gopls@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/go-delve/delve/cmd/dlv@latest

install-lint:
    go install github.com/kisielk/errcheck@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest

install-all: install-devenv install-lint

lint:
    #!/bin/bash
    set -e
    errcheck ./... || EXIT_CODE=1
    staticcheck ./... || EXIT_CODE=1
    exit ${EXIT_CODE:-0}
