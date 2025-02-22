# Splitter

A simple Go application that prints "Hello, World!".

## Building and Running

### Local Development

```bash
# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean
```

### Docker

```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run
```

## CI/CD

Docker images are automatically built and published to GitHub Container Registry (ghcr.io) on:
- Every commit to the main branch
- Every commit in a Pull Request
- Every tag creation (v* tags)

The images can be pulled using:
```bash
docker pull ghcr.io/mugdha-adhav/splitter:latest
```

Or with a specific tag:
```bash
docker pull ghcr.io/mugdha-adhav/splitter:v1.0.0
```

## Project Structure

- `main.go` - Main application code
- `Dockerfile` - Container configuration
- `Makefile` - Build and run automation
- `.build/` - Contains compiled binaries (gitignored)
- `.github/workflows/` - GitHub Actions workflow definitions
