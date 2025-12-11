#!/bin/bash

# Deploy script for BRiSK Calculator
# Usage: ./deploy.sh /path/to/destination

set -e

if [ -z "$1" ]; then
    echo "Usage: ./deploy.sh <destination_folder>"
    echo "Example: ./deploy.sh /var/www"
    exit 1
fi

DEST="$1"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist"

# Build the project
echo "Building project..."
cd "$SCRIPT_DIR"
npm run build

# Remove old brisk folder and recreate
echo "Removing old brisk folder..."
rm -rf "$DEST/brisk"

# Copy entire dist folder to brisk
echo "Copying dist to brisk..."
cp -r "$DIST_DIR" "$DEST/brisk"

# Update asset paths in index.html from /assets/ to brisk/
echo "Updating asset paths in index.html..."
sed 's|/assets/|brisk/|g' "$DEST/brisk/index.html" > "$DEST/brisk/index.html.tmp" && mv "$DEST/brisk/index.html.tmp" "$DEST/brisk/index.html"

# Move assets up and remove assets folder
echo "Moving assets..."
mv "$DEST/brisk/assets"/* "$DEST/brisk/"
rmdir "$DEST/brisk/assets"

# Rename index.html to _index.html (for blot.im)
echo "Renaming index.html to _index.html..."
mv "$DEST/brisk/index.html" "$DEST/brisk/_index.html"

echo "Deploy complete!"
echo "  Files deployed to: $DEST/brisk/"
