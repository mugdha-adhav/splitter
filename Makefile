.PHONY: build run clean docker-build docker-run

# Build the Go application
build:
	mkdir -p .build
	go build -o .build/main

# Run the application locally
run: build
	./.build/main

# Clean build artifacts
clean:
	rm -rf .build

# Build Docker image
docker-build:
	docker build -t splitter-app .

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 splitter-app

# Default target
all: build
