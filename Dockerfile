FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /awair-downloader .

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /awair-downloader /awair-downloader

ENTRYPOINT ["/awair-downloader"]
