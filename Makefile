.PHONY: generate build clean serve setup optimize-images optimize help

generate:
	@echo "Generating static site..."
	@go run generate.go

build: generate
	@echo "Build complete! Output in ./public/"

clean:
	@echo "Cleaning public directory..."
	@rm -rf public

serve:
	@echo "Serving site on http://localhost:5173 (with auto-rebuild)"
	@go run serve.go || true

setup:
	@echo "Setting up project dependencies..."
	@./scripts/install-dependencies.sh

optimize-images:
	@echo "Optimizing images..."
	@./scripts/optimize-images.sh

optimize: optimize-images

help:
	@echo "Available commands:"
	@echo "  make setup          - Install all dependencies (Go, WebP tools, ImageMagick)"
	@echo "  make optimize       - Optimize all images in content/images/ to WebP"
	@echo "  make generate       - Generate the static site"
	@echo "  make build          - Alias for generate"
	@echo "  make serve          - Generate, serve, and watch for changes (auto-rebuild)"
	@echo "  make clean          - Remove generated public directory"
	@echo "  make help           - Show this help message"

