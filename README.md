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

## Project Structure

- `main.go` - Main application code
- `Dockerfile` - Container configuration
- `Makefile` - Build and run automation
- `.build/` - Contains compiled binaries (gitignored)
