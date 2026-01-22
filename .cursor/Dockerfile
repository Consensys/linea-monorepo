# Base image with common development tools
FROM ubuntu:24.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    && rm -rf /var/lib/apt/lists/*

# Install Java 21 (JDK) - required for Kotlin/Java projects
RUN apt-get update && apt-get install -y \
    openjdk-21-jdk \
    && rm -rf /var/lib/apt/lists/*

# Install Gradle (matching project version)
RUN curl -fsSL https://services.gradle.org/distributions/gradle-8.12.1-bin.zip -o gradle.zip && \
    unzip gradle.zip -d /opt && \
    rm gradle.zip && \
    ln -s /opt/gradle-8.12.1/bin/gradle /usr/local/bin/gradle

# Install Node.js (LTS version)
RUN curl -fsSL https://deb.nodesource.com/setup_lts.x | bash - && \
    apt-get install -y nodejs

# Install pnpm
RUN npm install -g pnpm

# Install Go (required for prover)
RUN curl -fsSL https://go.dev/dl/go1.21.0.linux-amd64.tar.gz -o go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# Install Foundry (for Solidity development)
RUN curl -L https://foundry.paradigm.xyz | bash && \
    /root/.foundry/bin/foundryup

ENV PATH=$PATH:/root/.foundry/bin

# Create a non-root user
RUN useradd -m -s /bin/bash ubuntu

# Set user and working directory
USER ubuntu
WORKDIR /home/ubuntu

# Set up Go environment for the ubuntu user
ENV GOPATH=/home/ubuntu/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Verify installations
RUN java -version && \
    gradle -v && \
    node --version && \
    npm --version && \
    pnpm --version && \
    go version && \
    forge --version || echo "Foundry may need initialization"
