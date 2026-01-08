#!/bin/bash
# Image Optimization Script
# Converts all images in content/images/ to optimized WebP format
# Automatically removes originals after successful conversion

set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default settings
QUALITY=80
REMOVE_ORIGINALS=true

# Recommended resolutions
CONTENT_MAX_WIDTH=1200
COVER_WIDTH=1200
COVER_HEIGHT=630
SMALL_MAX_SIZE=400

usage() {
    echo "Usage: $0 [image_directory]"
    echo ""
    echo "Converts all images to optimized WebP format and removes originals"
    echo ""
    echo "Arguments:"
    echo "  image_directory    Directory containing images (default: content/images/ and static/images/)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Process content/images/ and static/images/"
    echo "  $0 content/images/     # Process specific directory"
    exit 1
}

# Check dependencies
check_dependencies() {
    if ! command -v cwebp &> /dev/null; then
        echo -e "${RED}Error: cwebp not found${NC}"
        echo "Install with: sudo apt install webp"
        exit 1
    fi
    
    # ImageMagick is optional - only needed for formats cwebp doesn't support
    if ! command -v convert &> /dev/null && ! command -v magick &> /dev/null; then
        echo -e "${YELLOW}Warning: ImageMagick not found${NC}"
        echo "Some formats (GIF, BMP) may not be supported. Install with: sudo apt install imagemagick"
    fi
}

# Get ImageMagick command (optional fallback)
get_convert_cmd() {
    if command -v magick &> /dev/null; then
        echo "magick"
    elif command -v convert &> /dev/null; then
        echo "convert"
    else
        return 1
    fi
}

# Check if cwebp supports the input format
cwebp_supports_format() {
    local ext="${1,,}"
    case "$ext" in
        jpg|jpeg|png|tiff|tif|webp)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Detect image type from filename or path
detect_image_type() {
    local filepath="$1"
    local filename=$(basename "$filepath")
    
    # Check if it's a cover image
    if [[ "$filename" =~ ^cover- ]] || [[ "$filepath" =~ /covers/ ]]; then
        echo "cover"
    # Check if it's a small image
    elif [[ "$filename" =~ ^(icon|logo|small)- ]] || [[ "$filepath" =~ /(icons|logos)/ ]]; then
        echo "small"
    else
        echo "content"
    fi
}

# Process a single image
process_image() {
    local input_file="$1"
    local filename=$(basename "$input_file")
    local name_without_ext="${filename%.*}"
    local ext="${filename##*.}"
    local dir=$(dirname "$input_file")
    
    # Skip if already WebP
    if [ "${ext,,}" = "webp" ]; then
        echo -e "${YELLOW}Skipping $filename (already WebP)${NC}"
        return 0
    fi
    
    # Determine output path (same directory, .webp extension)
    local output_file="$dir/${name_without_ext}.webp"
    
    # Detect image type
    local img_type=$(detect_image_type "$input_file")
    
    local cwebp_opts="-q $QUALITY -quiet"
    local resize_opts=""
    
    # Set resize options based on image type
    case "$img_type" in
        content)
            resize_opts="-resize $CONTENT_MAX_WIDTH 0"
            ;;
        cover)
            resize_opts="-resize $COVER_WIDTH $COVER_HEIGHT"
            ;;
        small)
            resize_opts="-resize $SMALL_MAX_SIZE $SMALL_MAX_SIZE"
            ;;
    esac
    
    echo -e "Processing ${GREEN}$filename${NC} -> ${GREEN}${name_without_ext}.webp${NC} (type: ${YELLOW}$img_type${NC})"
    
    # Check if cwebp supports this format directly
    if cwebp_supports_format "$ext"; then
        # Use cwebp directly (fastest, best compression)
        if cwebp $cwebp_opts $resize_opts "$input_file" -o "$output_file" 2>/dev/null; then
            # Success
            :
        else
            echo -e "  ${YELLOW}Warning: cwebp failed, trying with ImageMagick preprocessing${NC}"
            # Fallback: convert to PNG first, then cwebp
            local convert_cmd=$(get_convert_cmd)
            if [ $? -eq 0 ]; then
                local temp_file=$(mktemp --suffix=.png)
                $convert_cmd "$input_file" -strip "$temp_file" 2>/dev/null
                if cwebp $cwebp_opts $resize_opts "$temp_file" -o "$output_file" 2>/dev/null; then
                    rm -f "$temp_file"
                else
                    rm -f "$temp_file"
                    echo -e "  ${RED}Error: Failed to convert $filename${NC}"
                    return 1
                fi
            else
                echo -e "  ${RED}Error: Cannot convert $filename (format not supported)${NC}"
                return 1
            fi
        fi
    else
        # Format not directly supported by cwebp, use ImageMagick to convert first
        local convert_cmd=$(get_convert_cmd)
        if [ $? -ne 0 ]; then
            echo -e "  ${RED}Error: Format .$ext not supported by cwebp and ImageMagick not available${NC}"
            return 1
        fi
        
        local temp_file=$(mktemp --suffix=.png)
        
        # Resize with ImageMagick based on type
        case "$img_type" in
            content)
                $convert_cmd "$input_file" -resize "${CONTENT_MAX_WIDTH}x>" -strip "$temp_file" 2>/dev/null
                ;;
            cover)
                $convert_cmd "$input_file" -resize "${COVER_WIDTH}x${COVER_HEIGHT}^" -gravity center -extent "${COVER_WIDTH}x${COVER_HEIGHT}" -strip "$temp_file" 2>/dev/null
                ;;
            small)
                $convert_cmd "$input_file" -resize "${SMALL_MAX_SIZE}x${SMALL_MAX_SIZE}>" -strip "$temp_file" 2>/dev/null
                ;;
        esac
        
        # Convert to WebP with cwebp
        if cwebp $cwebp_opts "$temp_file" -o "$output_file" 2>/dev/null; then
            rm -f "$temp_file"
        else
            rm -f "$temp_file"
            echo -e "  ${RED}Error: Failed to convert $filename${NC}"
            return 1
        fi
    fi
    
    # Show file size reduction
    local original_size=$(stat -f%z "$input_file" 2>/dev/null || stat -c%s "$input_file" 2>/dev/null)
    local new_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null)
    local reduction=$(awk "BEGIN {printf \"%.1f\", (1 - $new_size/$original_size) * 100}")
    
    echo -e "  ${GREEN}✓${NC} Reduced by ${GREEN}${reduction}%${NC} ($(numfmt --to=iec-i --suffix=B $original_size 2>/dev/null || echo "${original_size} bytes") -> $(numfmt --to=iec-i --suffix=B $new_size 2>/dev/null || echo "${new_size} bytes"))"
    
    # Remove original if conversion successful
    if [ "$REMOVE_ORIGINALS" = true ]; then
        rm -f "$input_file"
        echo -e "  ${GREEN}✓${NC} Removed original"
    fi
}

# Main processing
main() {
    # Parse arguments
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        usage
    fi
    
    # Check dependencies
    check_dependencies
    
    echo -e "Quality: ${YELLOW}$QUALITY${NC}, Removing originals: ${YELLOW}$REMOVE_ORIGINALS${NC}"
    echo ""
    
    local total_count=0
    local dirs=()
    local supported_exts=("jpg" "jpeg" "png" "tiff" "tif" "gif" "bmp")
    
    # Determine which directories to process
    if [ -n "$1" ]; then
        # Process specified directory
        dirs=("$1")
    else
        # Process both default directories
        dirs=("content/images" "static/images")
    fi
    
    # Process each directory
    for INPUT_DIR in "${dirs[@]}"; do
        # Validate input
        if [ ! -d "$INPUT_DIR" ]; then
            echo -e "${YELLOW}Warning: Directory not found: $INPUT_DIR${NC} (skipping)"
            echo ""
            continue
        fi
        
        echo -e "${GREEN}Optimizing images in: $INPUT_DIR${NC}"
        
        local image_count=0
    for ext in "${supported_exts[@]}"; do
        while IFS= read -r -d '' file; do
            if [ -n "$file" ]; then
                process_image "$file" || true  # Continue even if one fails
                ((image_count++))
            fi
        done < <(find "$INPUT_DIR" -type f -iname "*.${ext}" -print0 2>/dev/null)
    done
    
        total_count=$((total_count + image_count))
        echo ""
    done
    
    if [ $total_count -eq 0 ]; then
        echo -e "${YELLOW}No images found in specified directories${NC}"
    else
        echo -e "${GREEN}✓ Processed $total_count image(s) total${NC}"
        echo -e "All images converted to WebP format"
    fi
}

main "$@"
