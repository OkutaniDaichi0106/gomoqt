#!/usr/bin/env -S deno run --allow-read --allow-write
/**
 * Migration script to convert Vitest test files to Deno test files
 * This script:
 * 1. Renames .test.ts files to _test.ts
 * 2. Converts Vitest imports to Deno standard library imports
 * 3. Converts expect() assertions to assertEquals() style
 * 4. Adds .ts extensions to relative imports
 */

import { walk } from "https://deno.land/std@0.224.0/fs/walk.ts";
import * as path from "https://deno.land/std@0.224.0/path/mod.ts";

const SRC_DIR = "./src";

// Mapping of Vitest functions to Deno equivalents
const IMPORT_REPLACEMENTS: Record<string, string> = {
  "import { describe, it, expect": 'import { describe, it, assertEquals, assertExists',
  "from 'vitest'": 'from "../deps.ts"',
  "import { describe, it, expect, beforeEach, afterEach": 
    'import { describe, it, beforeEach, afterEach, assertEquals, assertExists',
  "import { describe, it, expect, beforeEach, afterEach, vi":
    'import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows',
};

const ASSERTION_REPLACEMENTS: [RegExp, string][] = [
  [/expect\(([^)]+)\)\.toBe\(([^)]+)\)/g, "assertEquals($1, $2)"],
  [/expect\(([^)]+)\)\.toEqual\(([^)]+)\)/g, "assertEquals($1, $2)"],
  [/expect\(([^)]+)\)\.toBeDefined\(\)/g, "assertExists($1)"],
  [/expect\(([^)]+)\)\.toBeUndefined\(\)/g, "assertEquals($1, undefined)"],
  [/expect\(([^)]+)\)\.toBeNull\(\)/g, "assertEquals($1, null)"],
  [/expect\(([^)]+)\)\.toBeTruthy\(\)/g, "assertEquals(!!$1, true)"],
  [/expect\(([^)]+)\)\.toBeFalsy\(\)/g, "assertEquals(!!$1, false)"],
  [/expect\(([^)]+)\)\.toBeInstanceOf\(([^)]+)\)/g, "assertInstanceOf($1, $2)"],
];

async function migrateFile(filePath: string): Promise<void> {
  console.log(`Migrating: ${filePath}`);
  
  let content = await Deno.readTextFile(filePath);
  
  // Replace imports
  for (const [oldImport, newImport] of Object.entries(IMPORT_REPLACEMENTS)) {
    content = content.replaceAll(oldImport, newImport);
  }
  
  // Replace assertions
  for (const [pattern, replacement] of ASSERTION_REPLACEMENTS) {
    content = content.replace(pattern, replacement);
  }
  
  // Add .ts extensions to relative imports
  content = content.replace(
    /from ['"](\.\/.+?)['"];/g,
    (match, importPath) => {
      if (!importPath.endsWith(".ts") && !importPath.includes("/")) {
        return `from "${importPath}.ts";`;
      }
      return match;
    }
  );
  
  // Remove vi.mock statements (will need manual migration)
  content = content.replace(/vi\.mock\([^;]+\);?\n?/g, "// TODO: Migrate mock to Deno\n");
  
  await Deno.writeTextFile(filePath, content);
}

async function main() {
  const testFiles: string[] = [];
  
  // Find all .test.ts files
  for await (const entry of walk(SRC_DIR, { 
    exts: [".ts"],
    match: [/\.test\.ts$/],
  })) {
    if (entry.isFile) {
      testFiles.push(entry.path);
    }
  }
  
  console.log(`Found ${testFiles.length} test files to migrate`);
  
  for (const filePath of testFiles) {
    // Rename .test.ts to _test.ts
    const newPath = filePath.replace(/\.test\.ts$/, "_test.ts");
    
    await migrateFile(filePath);
    await Deno.rename(filePath, newPath);
    
    console.log(`  Renamed: ${path.basename(filePath)} -> ${path.basename(newPath)}`);
  }
  
  console.log("\nMigration complete!");
  console.log("Note: Mock statements marked with TODO need manual migration");
}

if (import.meta.main) {
  main();
}
