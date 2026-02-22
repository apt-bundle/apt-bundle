FROM ubuntu:22.04

ARG APT_BUNDLE_REPO_ENTRY="deb [arch=amd64,arm64,armhf,i386 trusted=yes] https://apt-bundle.org/repo/ stable main"

ENV DEBIAN_FRONTEND=noninteractive

# Install prerequisites and apt-bundle from APT repository
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates && \
    echo "$APT_BUNDLE_REPO_ENTRY" | tee /etc/apt/sources.list.d/apt-bundle.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends apt-bundle && \
    rm -rf /var/lib/apt/lists/*

# Copy both Aptfile and Aptfile.lock so exact versions can be reproduced
COPY Aptfile Aptfile.lock /app/
WORKDIR /app

# --locked installs the exact versions recorded in Aptfile.lock
RUN apt-bundle install --file /app/Aptfile --locked

CMD ["/bin/bash"]
