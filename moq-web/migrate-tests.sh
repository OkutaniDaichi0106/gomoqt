#!/bin/bash

# Migration script to convert Vitest test files to Deno test files
# This script:
# 1. Renames .test.ts files to _test.ts
# 2. Converts Vitest imports to Deno standard library imports
# 3. Converts expect() assertions to assertEquals() style
# 4. Adds .ts extensions to relative imports

set -e

SRC_DIR="./src"
BACKUP_DIR="./migration-backup-$(date +%Y%m%d-%H%M%S)"

echo "Creating backup at $BACKUP_DIR..."
mkdir -p "$BACKUP_DIR"
cp -r "$SRC_DIR" "$BACKUP_DIR/"

echo "Starting migration..."

# Find all .test.ts files
find "$SRC_DIR" -name "*.test.ts" -type f | while read -r file; do
    echo "Processing: $file"
    
    # Create new filename
    new_file="${file%.test.ts}_test.ts"
    
    # Create temporary file for sed operations
    temp_file=$(mktemp)
    
    # Apply transformations
    sed -e "s/from 'vitest'/from \"..\/deps.ts\"/g" \
        -e "s/from \"vitest\"/from \"..\/deps.ts\"/g" \
        -e "s/import { describe, it, expect }/import { describe, it, assertEquals, assertExists }/g" \
        -e "s/import { describe, it, expect, beforeEach }/import { describe, it, beforeEach, assertEquals, assertExists }/g" \
        -e "s/import { describe, it, expect, beforeEach, afterEach }/import { describe, it, beforeEach, afterEach, assertEquals, assertExists }/g" \
        -e "s/import { describe, it, expect, beforeEach, afterEach, vi }/import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows }/g" \
        -e "s/expect(\([^)]*\))\.toBe(\([^)]*\))/assertEquals(\1, \2)/g" \
        -e "s/expect(\([^)]*\))\.toEqual(\([^)]*\))/assertEquals(\1, \2)/g" \
        -e "s/expect(\([^)]*\))\.toBeDefined()/assertExists(\1)/g" \
        -e "s/expect(\([^)]*\))\.toBeUndefined()/assertEquals(\1, undefined)/g" \
        -e "s/expect(\([^)]*\))\.toBeNull()/assertEquals(\1, null)/g" \
        -e "s/expect(\([^)]*\))\.toBeInstanceOf(\([^)]*\))/assertInstanceOf(\1, \2)/g" \
        -e "s/from '\.\/\([^']*\)'/from \".\/\1.ts\"/g" \
        -e "s/from \"\.\/\([^\"]*\)\"/from \".\\/\1.ts\"/g" \
        "$file" > "$temp_file"
    
    # Move temp file to new location
    mv "$temp_file" "$new_file"
    
    # Remove original if different
    if [ "$file" != "$new_file" ]; then
        rm "$file"
    fi
    
    echo "  ✓ Migrated to: $new_file"
done

echo ""
echo "Migration complete!"
echo "Backup saved to: $BACKUP_DIR"
echo ""
echo "Next steps:"
echo "1. Review the migrated files manually"
echo "2. Fix any TODO comments for mocks"
echo "3. Run: deno task test"
echo "4. If successful, delete the backup: rm -rf $BACKUP_DIR"
