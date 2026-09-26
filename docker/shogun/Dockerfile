FROM golang:1.26.7-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildmode=pie -ldflags "-s -w" -o shogun ./cmd

FROM debian:12.15-slim AS runner

RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        openssh-client \
        ca-certificates \
        && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/sample.config.yaml /etc/shogun/config.yaml
COPY --from=builder /build/shogun ./shogun

EXPOSE 7007

CMD ["./shogun"]