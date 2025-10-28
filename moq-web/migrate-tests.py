#!/usr/bin/env python3
"""
Migration script to convert Vitest test files to Deno test files.
This script handles complex transformations that are difficult with sed/regex.
"""

import os
import re
import sys
from pathlib import Path
from datetime import datetime
import shutil

# Mapping of Vitest assertions to Deno equivalents
ASSERTION_PATTERNS = [
    # Simple assertions
    (r'expect\(([^)]+)\)\.toBe\(([^)]+)\)', r'assertEquals(\1, \2)'),
    (r'expect\(([^)]+)\)\.toEqual\(([^)]+)\)', r'assertEquals(\1, \2)'),
    (r'expect\(([^)]+)\)\.toStrictEqual\(([^)]+)\)', r'assertStrictEquals(\1, \2)'),
    (r'expect\(([^)]+)\)\.toBeDefined\(\)', r'assertExists(\1)'),
    (r'expect\(([^)]+)\)\.toBeUndefined\(\)', r'assertEquals(\1, undefined)'),
    (r'expect\(([^)]+)\)\.toBeNull\(\)', r'assertEquals(\1, null)'),
    (r'expect\(([^)]+)\)\.toBeTruthy\(\)', r'assertEquals(!!\1, true)'),
    (r'expect\(([^)]+)\)\.toBeFalsy\(\)', r'assertEquals(!!\1, false)'),
    (r'expect\(([^)]+)\)\.toBeInstanceOf\(([^)]+)\)', r'assertInstanceOf(\1, \2)'),
    (r'expect\(([^)]+)\)\.toContain\(([^)]+)\)', r'assertArrayIncludes(\1, [\2])'),
    (r'expect\(([^)]+)\)\.not\.toBe\(([^)]+)\)', r'assertNotEquals(\1, \2)'),
    (r'expect\(([^)]+)\)\.not\.toEqual\(([^)]+)\)', r'assertNotEquals(\1, \2)'),
    
    # Throw assertions - handle various patterns
    (r'expect\(\(\) => ([^)]+)\)\.toThrow\(\)', r'assertThrows(() => \1)'),
    (r'expect\(\(\) => ([^)]+)\)\.toThrow\(([^)]+)\)', r'assertThrows(() => \1, Error, \2)'),
    (r'expect\(\(\) => \{\s*([^}]+)\s*\}\)\.toThrow\(\)', r'assertThrows(() => { \1 })'),
    (r'expect\(\(\) => \{\s*([^}]+)\s*\}\)\.toThrow\(([^)]+)\)', r'assertThrows(() => { \1 }, Error, \2)'),
    
    # Async assertions
    (r'expect\(([^)]+)\)\.resolves\.toBe\(([^)]+)\)', r'assertEquals(await \1, \2)'),
    (r'expect\(([^)]+)\)\.rejects\.toThrow\(\)', r'assertRejects(async () => await \1)'),
]

IMPORT_PATTERNS = [
    # Replace vitest imports
    (r"from\s+['\"]vitest['\"]", 'from "../deps.ts"'),
    
    # Common import combinations
    (r"import\s+\{\s*describe,\s*it,\s*expect\s*\}",
     'import { describe, it, assertEquals, assertExists }'),
    (r"import\s+\{\s*describe,\s*it,\s*expect,\s*beforeEach\s*\}",
     'import { describe, it, beforeEach, assertEquals, assertExists }'),
    (r"import\s+\{\s*describe,\s*it,\s*expect,\s*afterEach\s*\}",
     'import { describe, it, afterEach, assertEquals, assertExists }'),
    (r"import\s+\{\s*describe,\s*it,\s*expect,\s*beforeEach,\s*afterEach\s*\}",
     'import { describe, it, beforeEach, afterEach, assertEquals, assertExists }'),
    (r"import\s+\{\s*describe,\s*it,\s*expect,\s*beforeEach,\s*afterEach,\s*vi\s*\}",
     'import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows }'),
    (r"import\s+\{\s*describe,\s*it,\s*expect,\s*vi\s*\}",
     'import { describe, it, assertEquals, assertExists, assertThrows }'),
]

def add_ts_extension(content: str) -> str:
    """Add .ts extensions to relative imports that don't have them."""
    # Match relative imports without .ts extension
    def replace_import(match):
        import_path = match.group(1)
        # Don't modify if it already has an extension or is a directory import
        if '.' in import_path.split('/')[-1] or import_path.endswith('/'):
            return match.group(0)
        return f'from "{import_path}.ts"'
    
    # Single quotes
    content = re.sub(r"from\s+'(\.\/[^']+)'", lambda m: f"from '{m.group(1)}.ts'" if '.' not in m.group(1).split('/')[-1] else m.group(0), content)
    # Double quotes
    content = re.sub(r'from\s+"(\.\/[^"]+)"', lambda m: f'from "{m.group(1)}.ts"' if '.' not in m.group(1).split('/')[-1] else m.group(0), content)
    
    return content

def calculate_relative_deps_path(file_path: Path, src_dir: Path) -> str:
    """Calculate the relative path from a file to deps.ts."""
    # Count directory depth
    relative_to_src = file_path.relative_to(src_dir)
    depth = len(relative_to_src.parts) - 1  # -1 because we don't count the file itself
    
    if depth == 0:
        return "../deps.ts"
    else:
        return "../" * depth + "../deps.ts"

def migrate_test_file(file_path: Path, src_dir: Path) -> str:
    """Migrate a single test file from Vitest to Deno."""
    content = file_path.read_text()
    
    # Calculate correct path to deps.ts
    deps_path = calculate_relative_deps_path(file_path, src_dir)
    
    # Replace imports
    for pattern, replacement in IMPORT_PATTERNS:
        content = re.sub(pattern, replacement, content)
    
    # Replace 'from "../deps.ts"' with the correct relative path
    content = content.replace('from "../deps.ts"', f'from "{deps_path}"')
    
    # Replace assertions
    for pattern, replacement in ASSERTION_PATTERNS:
        content = re.sub(pattern, replacement, content)
    
    # Add .ts extensions to imports
    content = add_ts_extension(content)
    
    # Remove or comment out vi.mock calls
    content = re.sub(r"vi\.mock\([^;]+\);?\s*\n", "// TODO: Migrate mock to Deno compatible pattern\n", content)
    
    # Remove unused vi import if present
    content = re.sub(r',\s*vi\s*(?=\})', '', content)
    
    return content

def main():
    src_dir = Path("./src")
    
    if not src_dir.exists():
        print("Error: src directory not found")
        sys.exit(1)
    
    # Create backup
    backup_dir = Path(f"./migration-backup-{datetime.now().strftime('%Y%m%d-%H%M%S')}")
    print(f"Creating backup at {backup_dir}...")
    shutil.copytree(src_dir, backup_dir / "src")
    
    # Find all .test.ts files
    test_files = list(src_dir.rglob("*.test.ts"))
    
    print(f"\nFound {len(test_files)} test files to migrate\n")
    
    migrated_count = 0
    for test_file in test_files:
        try:
            print(f"Processing: {test_file.relative_to(src_dir)}")
            
            # Read and migrate content
            migrated_content = migrate_test_file(test_file, src_dir)
            
            # Determine new filename
            new_file = test_file.with_name(test_file.stem.replace('.test', '_test') + test_file.suffix)
            
            # Write migrated content to new file
            new_file.write_text(migrated_content)
            
            # Remove old file if different
            if new_file != test_file:
                test_file.unlink()
            
            print(f"  ✓ Migrated to: {new_file.relative_to(src_dir)}")
            migrated_count += 1
            
        except Exception as e:
            print(f"  ✗ Error migrating {test_file}: {e}")
    
    print(f"\n{'='*60}")
    print(f"Migration complete!")
    print(f"Successfully migrated {migrated_count}/{len(test_files)} files")
    print(f"Backup saved to: {backup_dir}")
    print(f"{'='*60}")
    print("\nNext steps:")
    print("1. Review the migrated files manually")
    print("2. Address any TODO comments for mocks")
    print("3. Fix any complex assertions that weren't auto-converted")
    print("4. Run: deno task test")
    print(f"5. If successful, delete backup: rm -rf {backup_dir}")

if __name__ == "__main__":
    main()
