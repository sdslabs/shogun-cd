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

COPY sample.config.yaml /etc/shogun/config.yaml
COPY docker/shogun/entrypoint.sh /usr/local/bin/config-builder.sh
COPY --from=builder /build/shogun ./shogun

RUN chmod +x /usr/local/bin/config-builder.sh

EXPOSE 7007

ENTRYPOINT ["/usr/local/bin/config-builder.sh"]
CMD ["/app/shogun"]