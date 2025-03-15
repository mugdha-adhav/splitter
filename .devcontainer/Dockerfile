FROM node:20-slim

# Install basic development tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 5173
    
# Set working directory
WORKDIR /workspace 