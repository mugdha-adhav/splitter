.PHONY: build run clean docker-build docker-run help

build: ## Build the Go application
	mkdir -p backend/.build
	cd backend && CGO_ENABLED=1 go build -o .build/main main.go

run: build ## Run the application locally
	export ENV=local && cd backend && ./.build/main

clean: ## Clean build artifacts
	rm -rf backend/.build

docker-build: ## Build Docker image
	docker build -t splitter-app backend

docker-run: ## Run Docker container
	docker pull ghrc.io/mugdha-adhav/splitter:develop
	docker run -p 8080:8080 splitter-app

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
