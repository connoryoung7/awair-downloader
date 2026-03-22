default: build

setup:
    go mod tidy
    go fix ./...

tidy:
    go mod tidy

fix:
    go fix ./...

build:
    go build ./...

run *args:
    go run . {{args}}
