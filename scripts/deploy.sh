#!/bin/bash
# Deploy to GitHub Pages Script
# Builds the site and deploys the public folder to the gh-pages branch

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Deploying site to GitHub Pages...${NC}"
echo ""

# Get base path from environment or auto-detect
if [ -z "$BASE_PATH" ]; then
    # Try to detect repository name from git remote
    REPO_NAME=$(git remote get-url origin 2>/dev/null | sed -E 's/.*\/([^\/]+)\.git$/\1/' || echo "")
    
    # Check if it's a user/org page (repo name ends with .github.io)
    if [[ "$REPO_NAME" == *.github.io ]]; then
        BASE_PATH="/"
    # Check if repo name looks like a custom domain (contains dot but not .github.io)
    elif [[ "$REPO_NAME" == *.* ]] && [[ "$REPO_NAME" != *.github.io ]]; then
        # Likely a custom domain - use root path
        BASE_PATH="/"
    # Check if CNAME file exists (indicates custom domain = root path)
    elif [ -f "static/CNAME" ] || [ -f "CNAME" ] || git show origin/gh-pages:CNAME >/dev/null 2>&1; then
        # Custom domain detected - use root path
        BASE_PATH="/"
    elif [ -n "$REPO_NAME" ]; then
        # It's a project page, use repo name as base path
        BASE_PATH="/$REPO_NAME/"
    else
        # Fallback to root if we can't detect
        BASE_PATH="/"
    fi
fi

# Ensure BASE_PATH ends with / (unless it's just "/")
if [ "$BASE_PATH" != "/" ] && [ "${BASE_PATH: -1}" != "/" ]; then
    BASE_PATH="${BASE_PATH}/"
fi

echo -e "${BLUE}Base path: ${YELLOW}$BASE_PATH${NC}"
echo ""

# Step 1: Build the site
echo -e "${BLUE}[1/6] Building site...${NC}"
if BASE_PATH="$BASE_PATH" make generate; then
    echo -e "${GREEN}✓${NC} Site built successfully"
else
    echo -e "${RED}✗${NC} Build failed"
    exit 1
fi
echo ""

# Check if public directory exists and has content
if [ ! -d "public" ] || [ -z "$(ls -A public)" ]; then
    echo -e "${RED}Error: public directory is empty or doesn't exist${NC}"
    echo "Build may have failed. Check the output above."
    exit 1
fi

# Step 2: Get the current branch name
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${BLUE}[2/6] Current branch: ${YELLOW}$CURRENT_BRANCH${NC}"
echo ""

# Step 3: Check for uncommitted changes and stash if needed
echo -e "${BLUE}[3/6] Checking repository status...${NC}"
HAS_CHANGES=false
HAS_UNTRACKED=false
STASH_APPLIED=false

# Check for uncommitted changes
if ! git diff-index --quiet HEAD -- 2>/dev/null; then
    HAS_CHANGES=true
fi

# Check for untracked files
if [ -n "$(git ls-files --others --exclude-standard)" ]; then
    HAS_UNTRACKED=true
fi

if [ "$HAS_CHANGES" = true ] || [ "$HAS_UNTRACKED" = true ]; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    echo -e "${YELLOW}These will be stashed temporarily during deployment${NC}"
    
    # Stash changes (including untracked files)
    if git stash push --include-untracked -m "Temporary stash for deployment $(date '+%Y-%m-%d %H:%M:%S')" 2>/dev/null; then
        STASH_APPLIED=true
        echo -e "${GREEN}✓${NC} Changes stashed temporarily"
    else
        echo -e "${YELLOW}Note: No stash needed or stash failed${NC}"
    fi
else
    echo -e "${GREEN}✓${NC} Working directory clean"
fi
echo ""

# Step 4: Create temporary directory and copy public folder
echo -e "${BLUE}[4/6] Preparing deployment files...${NC}"
TEMP_DIR=$(mktemp -d)
cp -r public/* "$TEMP_DIR/" 2>/dev/null || {
    echo -e "${RED}Error: Failed to copy public folder${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
}
echo -e "${GREEN}✓${NC} Files prepared"
echo ""

# Step 5: Setup gh-pages branch
echo -e "${BLUE}[5/6] Setting up gh-pages branch...${NC}"
if git show-ref --verify --quiet refs/heads/gh-pages; then
    # Branch exists, checkout and reset it
    git checkout gh-pages
    git rm -rf . 2>/dev/null || true
    echo -e "${GREEN}✓${NC} Switched to existing gh-pages branch"
else
    # Branch doesn't exist, create orphan branch
    git checkout --orphan gh-pages
    git rm -rf . 2>/dev/null || true
    echo -e "${GREEN}✓${NC} Created new gh-pages branch"
fi

# Copy files from temp directory to root
cp -r "$TEMP_DIR"/* .
rm -rf "$TEMP_DIR"

# Add all files
git add -A

# Commit
if git diff --staged --quiet; then
    echo -e "${YELLOW}No changes to deploy (site unchanged)${NC}"
else
    DEPLOY_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    git commit -m "Deploy site: $DEPLOY_TIME" || {
        echo -e "${YELLOW}Nothing to commit${NC}"
    }
    echo -e "${GREEN}✓${NC} Changes committed"
fi
echo ""

# Step 6: Push to GitHub
echo -e "${BLUE}[6/6] Pushing to GitHub...${NC}"
if git push origin gh-pages --force; then
    echo -e "${GREEN}✓${NC} Pushed to GitHub Pages"
else
    echo -e "${RED}✗${NC} Failed to push to GitHub"
    echo -e "${YELLOW}Returning to $CURRENT_BRANCH branch...${NC}"
    git checkout "$CURRENT_BRANCH" 2>/dev/null || true
    
    # Restore stashed changes if we stashed them
    if [ "$STASH_APPLIED" = true ]; then
        echo -e "${BLUE}Restoring stashed changes...${NC}"
        git stash pop 2>/dev/null || true
    fi
    
    exit 1
fi
echo ""

# Return to original branch
echo -e "${BLUE}Returning to $CURRENT_BRANCH branch...${NC}"
git checkout "$CURRENT_BRANCH" 2>/dev/null || {
    echo -e "${YELLOW}Warning: Could not return to $CURRENT_BRANCH branch${NC}"
}

# Restore stashed changes if we stashed them
if [ "$STASH_APPLIED" = true ]; then
    echo -e "${BLUE}Restoring stashed changes...${NC}"
    if git stash pop 2>/dev/null; then
        echo -e "${GREEN}✓${NC} Changes restored"
    else
        echo -e "${YELLOW}Note: Stash restore had conflicts or was empty${NC}"
    fi
fi
echo ""

# Summary
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Deployment Summary${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓${NC} Site built successfully"
echo -e "${GREEN}✓${NC} Deployed to gh-pages branch"
echo -e "${GREEN}✓${NC} Pushed to GitHub"
echo ""
echo -e "${BLUE}Your site should be available at:${NC}"
if [ "$BASE_PATH" = "/" ]; then
    echo -e "${YELLOW}https://<your-username>.github.io/${NC}"
else
    echo -e "${YELLOW}https://<your-username>.github.io${BASE_PATH}${NC}"
fi
echo ""
echo -e "${YELLOW}Note:${NC} It may take a few minutes for GitHub Pages to update"
echo ""

