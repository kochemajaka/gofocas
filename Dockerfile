# Build stage — CGO enabled, x86-64 Linux
FROM golang:1.25 AS builder

WORKDIR /app

# Install gcc for CGO
RUN apt-get update && apt-get install -y gcc && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Place the shared library where CGO LDFLAGS -L. can find it at build time
RUN cp libfwlib32.so internal/fwlib32/libfwlib32.so

# Build all examples with real FOCAS CGO bindings
RUN mkdir -p /out && \
    for d in examples/*/; do \
        name=$(basename "$d"); \
        CGO_ENABLED=1 go build -tags=focas_cgo -o "/out/$name" "./$d"; \
    done

# Runtime stage — minimal, just needs libc
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y libc6 && rm -rf /var/lib/apt/lists/*

# Install shared library so the runtime linker finds it
COPY libfwlib32.so /usr/lib/libfwlib32.so.1
RUN ln -s /usr/lib/libfwlib32.so.1 /usr/lib/libfwlib32.so

COPY --from=builder /out /app/examples

# Default: run example 01-connect. Override CMD to run others.
ENV FOCAS_ADDR=192.168.1.1
CMD ["/app/examples/01-connect"]
