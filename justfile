# Run all appropriate linters in a fixing mode.
lint: && generate
    uvx zizmor .
    actionlint
    go mod tidy
    golangci-lint run --fast-only --fix
    goimports -w .

# Generate code.
generate:
    go generate ./...
    goimports -w .

# Run all tests.
test:
    go test ./...

# Run all tests, always. This skips the test cache.
test-all:
    go test -count=1 ./...

# Build the binary. This produces a production-ready result.
build:
    CGO_ENABLED=0 go build -o ./emailsub

# Builds distribution ZIPs for AWS Lambda, intended for arm64 Amazon Linux 2 Lambdas.
dist:
    rm -rf dist
    find cmd/ -maxdepth 1 -type d | tail -n+2 | parallel -j4 --no-notice --bar --progress "just _dist {}"

# Build one command for distribution.
_dist dir:
    mkdir -p dist/{{dir}}
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/{{dir}}/bootstrap -ldflags="-s -w" ./{{dir}}
    cd dist/{{dir}} && zip $(basename {{dir}}).zip bootstrap
    mv dist/{{dir}}/$(basename {{dir}}).zip dist/

# Run the binary.
run *ARGS: build
    ./emailsub {{ARGS}}

#  Update all dependencies.
update: && lint 
    go get -u ./...

# Set up a Git pre-commit hook to run (fast-only) linters before committing.
setup-precommit:
    echo "#!/bin/sh" > .git/hooks/pre-commit
    echo "set -e" >> .git/hooks/pre-commit
    echo "just _precommit" >> .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit

# Run by the pre-commit Git hook.
_precommit:
    uvx zizmor .
    actionlint
    goimports -l .
    go mod tidy -diff
