#!/bin/bash
# Script to list Node.js-specific files that can be removed after Deno migration

echo "Node.js-specific files to remove after successful migration:"
echo "=============================================================="
echo ""

cd moq-web 2>/dev/null || exit 1

files_to_remove=(
    "package.json"
    "package-lock.json"
    "pnpm-lock.yaml"
    ".npmrc"
    ".npmignore"
    "node_modules/"
    "vitest.config.ts"
    "vitest.setup.ts"
    "tsconfig.json"
    "tsconfig.test.json"
    "tsconfig.browser.json"
    "eslint.config.js"
    "tslint.json"
    "api-extractor.json"
    ".pnp.cjs"
)

echo "Files/directories found:"
for file in "${files_to_remove[@]}"; do
    if [ -e "$file" ]; then
        if [ -d "$file" ]; then
            size=$(du -sh "$file" 2>/dev/null | cut -f1)
            echo "  [DIR]  $file ($size)"
        else
            size=$(du -h "$file" 2>/dev/null | cut -f1)
            echo "  [FILE] $file ($size)"
        fi
    fi
done

echo ""
echo "To remove these files after verifying Deno tests pass, run:"
echo "  cd moq-web"
echo "  rm -rf ${files_to_remove[*]}"
echo ""
echo "Note: Keep package.json if you still need npm publishing support"
