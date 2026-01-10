.PHONY: generate clean serve setup optimize-images optimize deploy help

generate:
	@BASE_PATH="$${BASE_PATH:-/}" go run generate.go

clean:
	@echo "▓▓ CLEANING..."
	@rm -rf public
	@echo "▓▓ DONE"

serve:
	@go run serve.go || true

server: serve

setup:
	@echo "▓▓ INSTALLING DEPENDENCIES..."
	@./scripts/install-dependencies.sh
	@echo "▓▓ SETUP COMPLETE"

optimize-images:
	@echo "▓▓ OPTIMIZING IMAGES..."
	@./scripts/optimize-images.sh
	@echo "▓▓ OPTIMIZATION COMPLETE"

optimize: optimize-images

deploy:
	@echo "▓▓ DEPLOYING TO GITHUB PAGES..."
	@./scripts/deploy.sh
	@echo "▓▓ DEPLOYMENT COMPLETE"

help:
	@echo "▓▓ AVAILABLE COMMANDS:"
	@echo "  make setup       - Install dependencies (Go, WebP, ImageMagick)"
	@echo "  make generate    - Generate static site → public/"
	@echo "  make serve       - Dev server + hot reload (port 5174)"
	@echo "  make clean       - Remove public/ directory"
	@echo "  make optimize    - Optimize images to WebP"
	@echo "  make deploy      - Build + deploy to GitHub Pages"
	@echo "  make help        - Show this message"

