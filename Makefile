.PHONY: generate clean serve setup optimize-images optimize deploy help

generate:
	@echo "Generating static site..."
	@BASE_PATH="$${BASE_PATH:-/}" go run generate.go

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

deploy:
	@echo "Deploying to GitHub Pages..."
	@./scripts/deploy.sh

help:
	@echo "Available commands:"
	@echo "  make setup          - Install all dependencies (Go, WebP tools, ImageMagick)"
	@echo "  make optimize       - Optimize all images in content/images/ to WebP"
	@echo "  make generate       - Generate the static site"
	@echo "  make serve          - Generate, serve, and watch for changes (auto-rebuild)"
	@echo "  make clean          - Remove generated public directory"
	@echo "  make deploy         - Build and deploy to GitHub Pages"
	@echo "  make help           - Show this help message"

