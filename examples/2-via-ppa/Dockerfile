FROM ubuntu:22.04

# Prevent interactive prompts during apt operations
ENV DEBIAN_FRONTEND=noninteractive

# Install prerequisites and apt-bundle via install.sh
# (PPAs work with either installation method, but install.sh keeps it simple)
# Note: software-properties-common is required for add-apt-repository (used by PPAs)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        software-properties-common && \
    rm -rf /var/lib/apt/lists/* && \
    curl -fsSL https://raw.githubusercontent.com/apt-bundle/apt-bundle/main/install.sh | bash

# Copy Aptfile and install dependencies
COPY Aptfile /app/Aptfile
WORKDIR /app
RUN apt-bundle install --file /app/Aptfile

# Clean up apt cache to reduce image size
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Your application code
COPY . /app
CMD ["/bin/bash"]

