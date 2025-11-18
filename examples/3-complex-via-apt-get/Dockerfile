# Build stage
FROM ubuntu:22.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

# Install prerequisites and apt-bundle from APT repository
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl \
        ca-certificates && \
    echo "deb [arch=amd64,arm64,armhf,i386] [trusted=yes] https://apt-bundle.org/repo/ stable main" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends apt-bundle && \
    rm -rf /var/lib/apt/lists/*

# Install build dependencies
COPY Aptfile.build /app/Aptfile.build
WORKDIR /app
RUN apt-bundle install --file /app/Aptfile.build

# Build your application
COPY . /app
RUN make build

# Runtime stage
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install prerequisites and apt-bundle from APT repository
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl \
        ca-certificates && \
    echo "deb [arch=amd64,arm64,armhf,i386] [trusted=yes] https://apt-bundle.org/repo/ stable main" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends apt-bundle && \
    rm -rf /var/lib/apt/lists/*

# Install runtime dependencies only
COPY Aptfile.runtime /app/Aptfile.runtime
WORKDIR /app
RUN apt-bundle install --file /app/Aptfile.runtime

# Copy built binary from builder
COPY --from=builder /app/bin/myapp /app/myapp

CMD ["/app/myapp"]

