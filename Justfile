set dotenv-load

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

align:
    go tool betteralign ./...

verify-work:
    go build ./...
    go vet ./...

run *args:
    go run . {{args}}

docker-build:
    docker build -t awair-downloader .

docker-build-dev:
    docker build -f Dockerfile.dev -t awair-downloader-dev .

docker-run *args:
    docker run --env-file .env awair-downloader {{args}}

docker-run-dev *args:
    docker run --env-file .env awair-downloader-dev {{args}}

db-tail db="awair.db":
    sqlite3 -column -header {{db}} "SELECT * FROM readings ORDER BY timestamp DESC LIMIT 10"
