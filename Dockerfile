# ---- Build stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Static binary: CGO disabled so it runs on the scratch/alpine runtime image
# without needing glibc/musl compatibility shims.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/api ./api

USER app

EXPOSE 8080

ENTRYPOINT ["./api"]