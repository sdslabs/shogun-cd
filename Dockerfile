# syntax=docker/dockerfile:1

FROM golang:1.24-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO is disabled since every dependency here (pgx, x/crypto/ssh) is pure Go.
RUN CGO_ENABLED=0 go build -o /out/shogun ./cmd

FROM debian:bookworm-slim

# git + openssh-client are runtime deps: the app shells out to them to clone/pull
# manifest repos over SSH using generated deploy keys.
RUN apt-get update \
    && apt-get install -y --no-install-recommends git openssh-client ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --system shogun && useradd --system --gid shogun --create-home shogun

WORKDIR /app

COPY --from=builder /out/shogun /app/shogun

RUN mkdir -p /var/lib/shogun && chown -R shogun:shogun /var/lib/shogun /app

USER shogun

EXPOSE 7007

ENTRYPOINT ["/app/shogun"]
