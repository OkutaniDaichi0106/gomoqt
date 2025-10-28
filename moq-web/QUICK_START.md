# Quick Start Guide - Completing the Deno Migration

This guide helps you complete the final steps of the Deno migration.

## Prerequisites

Ensure Deno is installed:
```bash
deno --version
```

If not installed, install Deno:
```bash
curl -fsSL https://deno.land/install.sh | sh
```

## Step 1: Update Mock Patterns (Required)

8 test files use Vitest's `vi.fn()` which needs to be replaced with `createMock()`.

### Run the analyzer to see what needs updating:
```bash
cd moq-web
python3 analyze-mocks.py
```

### Manual updates needed:

For each file listed by the analyzer, make these changes:

**1. Update imports:**
```typescript
// Before
import { describe, it, expect, beforeEach, afterEach, vi, type Mock } from "../deps.ts";

// After
import { describe, it, beforeEach, afterEach, assertEquals, assertExists, createMock, type MockFunction } from "../deps.ts";
```

**2. Replace vi.fn() calls:**
```typescript
// Before
const mockFn = vi.fn();

// After
const mockFn = createMock<() => void>();

// With return value
// Before
const mockFn = vi.fn().mockReturnValue(42);

// After
const mockFn = createMock<() => number>().mockReturnValue(42);

// With async resolution
// Before
const mockFn = vi.fn().mockResolvedValue('result');

// After
const mockFn = createMock<() => Promise<string>>().mockResolvedValue('result');
```

**3. Remove vi.mock() lines:**
```typescript
// Before
vi.mock('./module');

// After
// Remove this line - mocks are now created manually as objects
```

### Files to update (in order of complexity):

1. ✅ **src/client_test.ts** - Simplest (6 patterns)
2. ✅ **src/session_test.ts** - Simple (15 patterns)
3. ✅ **src/group_stream_test.ts** - Medium (26 patterns)
4. ✅ **src/track_mux_test.ts** - Medium (26 patterns)
5. ✅ **src/track_test.ts** - Medium (30 patterns)
6. ✅ **src/session_stream_test.ts** - Medium (44 patterns)
7. ✅ **src/subscribe_stream_test.ts** - Complex (56 patterns)
8. ✅ **src/announce_stream_test.ts** - Most complex (142 patterns)

## Step 2: Test the Migration

Once mocks are updated, test with Deno:

```bash
cd moq-web

# 1. Type check
deno check mod.ts

# 2. Run linter
deno lint

# 3. Check formatting
deno fmt --check

# 4. Run all tests
deno task test

# 5. Generate coverage
deno task coverage
```

### If you encounter errors:

**Import errors:**
- Ensure all relative imports have `.ts` extensions
- Check that deps.ts is properly configured

**Type errors:**
- Verify mock function signatures match expected types
- Add explicit type parameters to createMock<TYPE>()

**Network errors:**
- Ensure you have internet access for downloading dependencies
- First run may be slower as Deno downloads and caches dependencies

## Step 3: Clean Up Node.js Files

After confirming all tests pass:

```bash
cd moq-web

# See what will be removed
./list-node-files.sh

# Remove Node.js files (CAUTION: Make sure tests pass first!)
rm -rf package.json package-lock.json pnpm-lock.yaml \
       .npmrc .npmignore node_modules/ \
       vitest.config.ts vitest.setup.ts \
       tsconfig.json tsconfig.test.json tsconfig.browser.json \
       eslint.config.js tslint.json api-extractor.json .pnp.cjs
```

**Note:** Keep `package.json` if you still need to publish to npm.

## Step 4: Update CI/CD

1. **Enable Deno CI:**
   - The workflow is already created at `.github/workflows/moq-web-deno-ci.yml`
   - It will run automatically on push/PR to main branch

2. **Optional - Disable Node.js CI:**
   - Edit `.github/workflows/moq-web-ci.yml`
   - Either disable it or update it to use the old Node.js workflow only for npm publishing

## Step 5: Verify Everything Works

Final checklist:

- [ ] All tests pass: `deno task test`
- [ ] No lint errors: `deno lint`
- [ ] Properly formatted: `deno fmt --check`
- [ ] Types check: `deno check mod.ts`
- [ ] Coverage generated: `deno task coverage`
- [ ] CI workflow runs successfully
- [ ] Documentation is up to date

## Common Issues & Solutions

### Issue: "Module not found"
**Solution:** Add `.ts` extension to the import path

### Issue: "Type 'MockFunction<...>' is not assignable"
**Solution:** Add explicit type parameter: `createMock<YourType>()`

### Issue: "Cannot find name 'vi'"
**Solution:** Replace `vi.fn()` with `createMock()` and update imports

### Issue: SSL certificate error
**Solution:** You're likely behind a corporate proxy. Set `DENO_TLS_CA_STORE=system` or use `--unsafely-ignore-certificate-errors` (not recommended for production)

## Development Workflow

### Running tests during development:
```bash
# Watch mode - reruns tests on file changes
deno task test:watch

# Run specific test file
deno test --allow-all src/info_test.ts

# Run with coverage
deno task coverage
```

### Formatting code:
```bash
# Check formatting
deno fmt --check

# Auto-format all files
deno fmt
```

### Linting:
```bash
# Run linter
deno lint

# Auto-fix some issues
deno lint --fix
```

## Need Help?

1. **See detailed migration guide:** `DENO_MIGRATION.md`
2. **Check migration summary:** `MIGRATION_SUMMARY.md`
3. **Analyze mock usage:** `python3 analyze-mocks.py`
4. **List Node files:** `./list-node-files.sh`

## Success Criteria

You've successfully completed the migration when:

✅ `deno task test` shows all tests passing
✅ `deno lint` shows no errors
✅ `deno fmt --check` shows no formatting issues
✅ `deno check mod.ts` shows no type errors
✅ Coverage report is generated successfully
✅ CI workflow runs without errors
✅ Node.js files are removed (optional)

## Next Steps After Migration

1. Consider publishing to JSR (JavaScript Registry) for Deno users
2. Update any documentation that references npm commands
3. Inform contributors about the new Deno workflow
4. Update CONTRIBUTING.md with Deno development instructions

---

**Estimated time to complete:** 2-4 hours (depending on mock complexity)

**Questions?** Check the detailed documentation in `DENO_MIGRATION.md` or the `MIGRATION_SUMMARY.md`
