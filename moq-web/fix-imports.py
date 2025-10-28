#!/usr/bin/env python3
"""
Script to add .ts extensions to all relative imports in source files.
This is required for Deno to resolve module imports correctly.
"""

import os
import re
import sys
from pathlib import Path

def add_ts_extension_to_imports(content: str) -> str:
    """Add .ts extensions to relative imports that don't have them."""
    
    def replace_import(match):
        quote_char = match.group(1)
        import_path = match.group(2)
        
        # Skip if already has .ts extension
        if import_path.endswith('.ts'):
            return match.group(0)
        
        # Skip if it's a directory import (ends with /)
        if import_path.endswith('/'):
            return match.group(0)
        
        # Skip if it's a package import (doesn't start with ./ or ../)
        if not (import_path.startswith('./') or import_path.startswith('../')):
            return match.group(0)
        
        # Skip if it has another extension
        path_parts = import_path.split('/')
        last_part = path_parts[-1]
        if '.' in last_part:
            return match.group(0)
        
        # Add .ts extension
        return f'from {quote_char}{import_path}.ts{quote_char}'
    
    # Match both single and double quoted imports
    content = re.sub(r"from\s+(['\"])([^'\"]+)(['\"])", replace_import, content)
    
    return content

def process_file(file_path: Path) -> bool:
    """Process a single file and return True if changes were made."""
    try:
        original_content = file_path.read_text()
        modified_content = add_ts_extension_to_imports(original_content)
        
        if original_content != modified_content:
            file_path.write_text(modified_content)
            return True
        return False
    except Exception as e:
        print(f"  ✗ Error processing {file_path}: {e}")
        return False

def main():
    src_dir = Path("./src")
    
    if not src_dir.exists():
        print("Error: src directory not found")
        sys.exit(1)
    
    # Find all .ts files excluding test files
    source_files = [
        f for f in src_dir.rglob("*.ts")
        if not f.name.endswith('_test.ts')
    ]
    
    print(f"Found {len(source_files)} source files to process\n")
    
    modified_count = 0
    for source_file in sorted(source_files):
        if process_file(source_file):
            print(f"✓ Updated: {source_file.relative_to(src_dir)}")
            modified_count += 1
    
    print(f"\n{'='*60}")
    print(f"Processing complete!")
    print(f"Modified {modified_count}/{len(source_files)} files")
    print(f"{'='*60}")

if __name__ == "__main__":
    main()
