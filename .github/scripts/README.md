# GitHub Scripts

This directory contains automation scripts for the gomoqt repository.

## Requirements

- Deno 2.0+

## Scripts

### translate.ts

Automatically translates GitHub issue and PR bodies into multiple languages using LibreTranslate API.

**Usage:**

```bash
# Run directly
deno run --allow-net --allow-env translate.ts

# Or use the task
deno task translate
```

**Required Environment Variables:**

- `GITHUB_TOKEN` - GitHub personal access token with repo permissions
- `GITHUB_EVENT` - JSON string containing the GitHub event data
- `GITHUB_REPOSITORY` - Repository name in format `owner/repo`

**Supported Languages:**

- English 🇬🇧
- Japanese 🇯🇵
- Chinese (Simplified) 🇨🇳
- Korean 🇰🇷
- French 🇫🇷
- Spanish 🇪🇸
- German 🇩🇪
- Russian 🇷🇺
- Portuguese 🇵🇹
- Arabic 🇸🇦

## Development

All scripts are written in TypeScript and run with Deno. Configuration is in `deno.json`.
