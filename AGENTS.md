# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

**siyuan-unlock** is a community fork of [SiYuan Note](https://github.com/siyuan-note/siyuan) that removes licensing restrictions. The repo contains the full SiYuan source merged at root level, an Android subtree under `android/`, and a set of patches applied during CI builds to customize upstream code.

## Repository Remotes & Merge Strategy

| Remote | URL | Purpose |
|--------|-----|---------|
| `origin` | github.com/gt4o4/siyuan-unlock | Fork |
| `upstream` | github.com/appdev/siyuan-unlock | Upstream unlock project |
| `siyuan` | github.com/siyuan-note/siyuan | Official SiYuan |
| `siyuan-android` | github.com/siyuan-note/siyuan-android | Official Android app |

- **siyuan** is merged as a full-repo merge (root level: `kernel/`, `app/`, `scripts/`)
- **siyuan-android** is merged as a git subtree into `android/` using `git merge -s subtree`
- **Tag collision warning**: upstream (appdev) and siyuan share tag names (e.g. `v3.6.0`) pointing to different commits. Delete local tags before fetching siyuan tags.

## Architecture

```
kernel/          Go backend (HTTP server, API, data model, search, sync)
app/             TypeScript/Electron frontend (editor, UI, themes, i18n)
android/         Android app (subtree from siyuan-android)
patches/         Unified diff patches applied to upstream during CI
  siyuan/          3 patches: disable-update, default-config, mock-vip-user
  siyuan-android/  1 patch: debug-build (custom signing config)
  siyuan-ios/      1 patch: build-failed
scripts/         Build scripts (linux-build.sh, darwin-build.sh, win-build.bat)
.github/workflows/  CI pipelines (desktop, android, iOS, docker, cron)
```

**Key relationship**: CI workflows clone upstream repos fresh, apply patches, then build. The merged source in this repo serves as reference; the patches define all fork customizations.

## Build Commands

### Desktop (full build)
```bash
./scripts/linux-build.sh [--target=amd64|arm64|all]
./scripts/darwin-build.sh [--target=amd64|arm64|all]
```

### Frontend only
```bash
cd app && pnpm install --no-frozen-lockfile && pnpm run build
```

### Kernel only
```bash
cd kernel && go build --tags fts5 -ldflags "-s -w" -o ../app/kernel-linux/SiYuan-Kernel .
```

The `fts5` build tag is **required** for SQLite full-text search.

### Android kernel (AAR)
```bash
gomobile bind --tags fts5 -androidapi 21 -target='android/arm64' ./kernel/mobile
```

### Packaging
```bash
cd app
pnpm run dist-linux        # Linux AppImage
pnpm run dist-darwin       # macOS Intel DMG
pnpm run dist-darwin-arm64 # macOS ARM64 DMG
pnpm run dist              # Windows NSIS
```

### Lint
```bash
cd app && pnpm run lint
```

### Docker
```bash
docker build -t siyuan-unlock .
```

## Patch System

Patches use unified diff format and are applied with `git apply` during CI. The siyuan patches use custom path prefixes (`forkSrcPrefix/`, `forkDstPrefix/`) — apply with `-p1`.

**Applying patches locally:**
```bash
# Siyuan patches (already applied in merged repo)
git apply -p1 patches/siyuan/disable-update.patch

# Android patch (needs --directory for subtree)
git apply -p1 --directory=android patches/siyuan-android/debug-build.patch
```

**Verifying patches against upstream:**
```bash
# Check if a patch can apply cleanly to a fresh upstream clone
git apply --check patches/siyuan/disable-update.patch

# Check if a patch is already applied (reverse check)
git apply --check --reverse -p1 patches/siyuan/disable-update.patch
```

After merging a new upstream version, always verify all patches still apply cleanly against the upstream tag.

## What Each Patch Does

- **disable-update.patch**: Stubs out `checkUpdate()`, forces `DownloadInstallPkg=false`, disables startup version check
- **default-config.patch**: Sets defaults (S3 sync provider, zh_CN language, minimize-to-tray on close, hide VIP badge)
- **mock-vip-user.patch**: Replaces `getCloudUser()` with a mock returning a VIP user (subscription never expires)
- **debug-build.patch**: Replaces Android signing config with env-based keystore (`KEYSTORE_PASSWORD`)

## Upstream Sync Workflow

1. `git fetch siyuan && git merge v<version>` — full-repo merge, resolve conflicts preserving fork modifications
2. `git fetch siyuan-android && git merge -s subtree <sha>` — subtree merge into `android/`
3. Verify all patches apply against the new version
4. Update patches if needed (regenerate diffs against the new upstream tag)

## Version Convention

The fork uses version prefix `103.x.y` (e.g., `103.6.0`) in `kernel/util/working.go` and `app/appx/AppxManifest.xml` to distinguish from official builds. The `app/package.json` version tracks upstream (e.g., `3.6.0`).

## Key Technologies

- **Go 1.25+** with SQLite FTS5 for the kernel
- **pnpm 10.x** / Node 20 / Electron 40 / Webpack 5 / TypeScript for the frontend
- **Gradle / Android SDK 36** for Android builds
- **musl libc** cross-compilers for static Linux binaries
