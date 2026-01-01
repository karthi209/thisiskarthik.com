# Scripts

Utility scripts for content management and optimization.

## install-dependencies.sh

Installs all required dependencies for the static site generator.

### Usage

**Recommended:** Use the make command:
```bash
make setup
```

**Alternative:** Run the script directly:
```bash
./scripts/install-dependencies.sh
```

### What it installs

- **Go (golang)**: Required for the static site generator
- **WebP tools (cwebp)**: Required for image optimization
- **ImageMagick**: Optional but recommended for advanced image processing
- **Go module dependencies**: Downloads Go packages from go.mod

### Supported OS

- Ubuntu/Debian
- Fedora/RHEL/CentOS
- Arch/Manjaro

### Manual installation

If the script doesn't work on your system:

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y golang-go webp imagemagick

# Fedora/RHEL
sudo dnf install -y golang libwebp-tools ImageMagick

# Arch/Manjaro
sudo pacman -S go libwebp imagemagick
```

## optimize-images.sh

Converts all images in `content/images/` to optimized WebP format and automatically removes originals.

### Requirements

- **cwebp** (required): Install with `make setup` or `sudo apt install webp`
- **ImageMagick** (optional): Install with `make setup` or `sudo apt install imagemagick` - Only needed for formats cwebp doesn't support (GIF, BMP, etc.)

### Usage

**Recommended:** Use the make command:
```bash
# Process all images in content/images/ (default)
make optimize
```

**Alternative:** Run the script directly:
```bash
# Process all images in content/images/ (default)
./scripts/optimize-images.sh

# Process specific directory
./scripts/optimize-images.sh content/images/2025/05/
```

### Features

- **Automatic type detection**: Detects image type from filename/path
  - `cover-*` or `/covers/` → Cover images (1200×630px)
  - `icon-*`, `logo-*`, `small-*` or `/icons/`, `/logos/` → Small images (max 400px)
  - Everything else → Content images (max 1200px width)
- **Automatic cleanup**: Removes original JPG/PNG files after successful conversion
- **Performance optimized**: Uses cwebp for best compression and speed
- **Shows progress**: Displays file size reduction for each image

### Image Types & Sizes

- **content** (default): Blog post images, max 1200px width, maintains aspect ratio
- **cover**: Featured/cover images, 1200×630px (1.9:1 ratio), crops to fit
- **small**: Icons/illustrations, max 400×400px, maintains aspect ratio

### Examples

```bash
# Recommended: Use make
make optimize

# Or run script directly
./scripts/optimize-images.sh

# Process specific folder (script only)
./scripts/optimize-images.sh content/images/2025/05/
```

### Output

- Converts all images to `.webp` format
- Removes original JPG/PNG files automatically
- Shows file size reduction percentage
- Skips files already in WebP format
- Preserves directory structure

