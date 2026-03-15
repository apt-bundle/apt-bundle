FROM ubuntu:22.04

ARG APT_BUNDLE_REPO_ENTRY="deb [arch=amd64,arm64,armhf,i386 trusted=yes] https://apt-bundle.org/repo/ stable main"

ENV DEBIAN_FRONTEND=noninteractive

# Install prerequisites and apt-bundle from APT repository
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates

RUN echo "$APT_BUNDLE_REPO_ENTRY" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends apt-bundle && \
    rm -rf /var/lib/apt/lists/*

# Install GitHub CLI via apt-bundle
COPY Aptfile /app/Aptfile
WORKDIR /app
RUN apt-bundle install --file /app/Aptfile

# Clean up apt cache to reduce image size
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

CMD ["/bin/bash"]
