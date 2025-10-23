# Hang Media

## Packaging & publish

This package is prepared to be published to npm. Follow these steps:

1. Build

```powershell
npm run build
```

2. Verify `dist` is created and contains `index.js`, `browser.js` and type declarations.

3. Publish (optional):

```powershell
npm publish --access public
```

Notes:
- The package outputs to `dist/`. `package.json` `files` includes `dist` so built artifacts are published.
- Exports provide `./elements` subpath; adjust as needed.
 - This project copies `src/*.ts` into `dist/` after build so published package includes original TypeScript sources for better editor "go to definition" experience.


