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

# Get base path from environment or default to "/" for custom domain
BASE_PATH="${BASE_PATH:-/}"

# Ensure BASE_PATH ends with / (unless it's just "/")
if [ "$BASE_PATH" != "/" ] && [ "${BASE_PATH: -1}" != "/" ]; then
    BASE_PATH="${BASE_PATH}/"
fi

echo -e "${BLUE}Base path: ${YELLOW}$BASE_PATH${NC}"
echo ""

# Step 1: Build the site from an empty output directory so stale or sensitive
# files can never ride along from an earlier run.
echo -e "${BLUE}[1/6] Building site...${NC}"
if BASE_PATH="$BASE_PATH" make clean generate; then
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

# Step 2: Snapshot the generated site before stashing changes. The public/
# directory may contain staged or otherwise tracked files, so stashing first
# can remove the build that we are about to deploy.
echo -e "${BLUE}[2/6] Preparing deployment files...${NC}"
TEMP_DIR=$(mktemp -d)
if ! cp -a public/. "$TEMP_DIR/"; then
    echo -e "${RED}Error: Failed to copy public folder${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Refuse to deploy environment files even if one somehow reaches public/.
SENSITIVE_FILE=$(find "$TEMP_DIR" -type f \( -name '.env' -o -name '.env.*' \) -print -quit)
if [ -n "$SENSITIVE_FILE" ]; then
    echo -e "${RED}Error: Refusing to deploy sensitive file: ${SENSITIVE_FILE#"$TEMP_DIR"/}${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi
echo -e "${GREEN}✓${NC} Files prepared"
echo ""

# Step 3: Get the current branch name
CURRENT_BRANCH=$(git branch --show-current)
echo -e "${BLUE}[3/6] Current branch: ${YELLOW}$CURRENT_BRANCH${NC}"
echo ""

# Step 4: Check for uncommitted changes and stash if needed
echo -e "${BLUE}[4/6] Checking repository status...${NC}"
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

# Step 5: Create a fresh orphan deployment commit. Keeping deployment history
# can leave deleted secrets reachable through older gh-pages commits.
echo -e "${BLUE}[5/6] Setting up gh-pages branch...${NC}"
DEPLOY_BRANCH="gh-pages-deploy-$$"
git checkout --orphan "$DEPLOY_BRANCH"
git rm -rf . 2>/dev/null || true
echo -e "${GREEN}✓${NC} Created clean deployment branch"

# Copy files from temp directory to root
cp -a "$TEMP_DIR"/. .
rm -rf "$TEMP_DIR"

# Create CNAME file for custom domain
REPO_NAME=$(git remote get-url origin 2>/dev/null | sed -E 's/.*\/([^\/]+)\.git$/\1/' || echo "thisiskarthik.com")
echo "$REPO_NAME" > CNAME
echo -e "${GREEN}✓${NC} Created CNAME file for custom domain"

# Add all files
git add -A

# Commit
DEPLOY_TIME=$(date '+%Y-%m-%d %H:%M:%S')
git commit -m "Deploy site: $DEPLOY_TIME"
git branch -f gh-pages HEAD
echo -e "${GREEN}✓${NC} Changes committed"
echo ""

# Step 6: Push to GitHub
echo -e "${BLUE}[6/6] Pushing to GitHub...${NC}"
if git push origin HEAD:gh-pages --force; then
    echo -e "${GREEN}✓${NC} Pushed to GitHub Pages"
else
    echo -e "${RED}✗${NC} Failed to push to GitHub"
    echo -e "${YELLOW}Returning to $CURRENT_BRANCH branch...${NC}"
    git checkout "$CURRENT_BRANCH" 2>/dev/null || true
    git branch -D "$DEPLOY_BRANCH" 2>/dev/null || true
    
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
git branch -D "$DEPLOY_BRANCH" 2>/dev/null || true

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
