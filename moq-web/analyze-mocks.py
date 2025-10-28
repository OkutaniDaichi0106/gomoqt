#!/usr/bin/env python3
"""
Script to help identify and update vi.fn() mock patterns to createMock()
This script provides a report of mock usage and suggests replacements.
"""

import re
from pathlib import Path

def find_mock_patterns(content: str, file_path: Path) -> list:
    """Find all vi.fn() patterns in the content."""
    patterns = []
    
    # Find vi.fn() calls
    vi_fn_pattern = r'vi\.fn\([^)]*\)'
    for match in re.finditer(vi_fn_pattern, content):
        patterns.append({
            'type': 'vi.fn()',
            'original': match.group(0),
            'suggestion': 'createMock<TYPE>()',
        })
    
    # Find vi.mock() calls
    vi_mock_pattern = r"vi\.mock\([^)]+\)"
    for match in re.finditer(vi_mock_pattern, content):
        patterns.append({
            'type': 'vi.mock()',
            'original': match.group(0),
            'suggestion': '// Manual mock object needed',
        })
    
    # Find .mockReturnValue() calls
    mock_return_pattern = r'\.mockReturnValue\([^)]+\)'
    for match in re.finditer(mock_return_pattern, content):
        patterns.append({
            'type': '.mockReturnValue()',
            'original': match.group(0),
            'suggestion': match.group(0),  # This is already compatible
        })
    
    # Find .mockResolvedValue() calls
    mock_resolved_pattern = r'\.mockResolvedValue\([^)]+\)'
    for match in re.finditer(mock_resolved_pattern, content):
        patterns.append({
            'type': '.mockResolvedValue()',
            'original': match.group(0),
            'suggestion': match.group(0),  # This is already compatible
        })
    
    # Find .mockImplementation() calls
    mock_impl_pattern = r'\.mockImplementation\([^)]+\)'
    for match in re.finditer(mock_impl_pattern, content):
        patterns.append({
            'type': '.mockImplementation()',
            'original': match.group(0),
            'suggestion': match.group(0),  # This is already compatible
        })
    
    return patterns

def analyze_test_files():
    """Analyze all test files for mock patterns."""
    src_dir = Path("./src")
    test_files = list(src_dir.rglob("*_test.ts"))
    
    files_with_mocks = {}
    
    for test_file in test_files:
        content = test_file.read_text()
        
        # Check for vi references
        if 'vi.' in content or 'Mock' in content:
            patterns = find_mock_patterns(content, test_file)
            if patterns:
                files_with_mocks[test_file] = patterns
    
    return files_with_mocks

def print_report(files_with_mocks: dict):
    """Print a detailed report of mock usage."""
    print("="*80)
    print("MOCK PATTERN ANALYSIS REPORT")
    print("="*80)
    print()
    
    if not files_with_mocks:
        print("No mock patterns found! ✓")
        return
    
    print(f"Found mock patterns in {len(files_with_mocks)} files:\n")
    
    for file_path, patterns in files_with_mocks.items():
        print(f"📄 {file_path.relative_to(Path('.'))}")
        print(f"   Found {len(patterns)} mock pattern(s)")
        
        # Group by type
        types = {}
        for pattern in patterns:
            t = pattern['type']
            if t not in types:
                types[t] = 0
            types[t] += 1
        
        for t, count in types.items():
            print(f"   - {t}: {count}")
        print()
    
    print("="*80)
    print("MIGRATION STEPS")
    print("="*80)
    print()
    print("1. Add createMock import to deps.ts (already done ✓)")
    print()
    print("2. Update each test file:")
    print("   - Replace: import { ..., vi } from '../deps.ts'")
    print("   - With:    import { ..., createMock } from '../deps.ts'")
    print()
    print("3. Replace vi.fn() patterns:")
    print("   - Before: const mock = vi.fn().mockReturnValue(42)")
    print("   - After:  const mock = createMock<() => number>().mockReturnValue(42)")
    print()
    print("4. Replace vi.mock() patterns with manual mocks:")
    print("   - Before: vi.mock('./module')")
    print("   - After:  const mockModule = { method: createMock() }")
    print()
    print("5. Remove 'vi' from type imports:")
    print("   - Before: import { ..., type Mock } from '../deps.ts'")
    print("   - After:  import { ..., type MockFunction } from '../deps.ts'")
    print()

def main():
    print("Analyzing test files for mock patterns...\n")
    
    files_with_mocks = analyze_test_files()
    print_report(files_with_mocks)
    
    print("\nFor detailed migration instructions, see DENO_MIGRATION.md")

if __name__ == "__main__":
    main()
