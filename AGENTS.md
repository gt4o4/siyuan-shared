# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

**siyuan-shared** is a community fork of [SiYuan Note](https://github.com/siyuan-note/siyuan) with an in-process kernel (Go c-shared library via N-API), S3/WebDAV sync, and self-hosted Docker deployment. The repo contains the full SiYuan source merged at root level, an Android subtree under `android/`, an iOS subtree under `ios/`, and a set of patches applied during CI builds.

## Repository Remotes & Merge Strategy

| Remote | URL | Purpose |
|--------|-----|---------|
| `origin` | github.com/gt4o4/siyuan-shared | Fork |
| `upstream` | github.com/appdev/siyuan-unlock | Upstream unlock project |
| `siyuan` | github.com/siyuan-note/siyuan | Official SiYuan |
| `siyuan-android` | github.com/siyuan-note/siyuan-android | Official Android app |
| `siyuan-ios` | github.com/siyuan-note/siyuan-ios | Official iOS app |

- **siyuan** is merged as a full-repo merge (root level: `kernel/`, `app/`, `scripts/`)
- **siyuan-android** is merged as a proper git subtree into `android/` using `git subtree pull --prefix=android siyuan-android main` (no `--squash`)
- **siyuan-ios** is merged as a proper git subtree into `ios/` using `git subtree pull --prefix=ios siyuan-ios main` (no `--squash`)
- **upstream** (appdev) is merged with `git merge upstream/master`, keeping our custom CI workflow
- **Tag collision warning**: upstream (appdev) and siyuan share tag names (e.g. `v3.6.0`) pointing to different commits. Delete local tags before fetching siyuan tags.
- **Release tags**: Use `v<upstream>-gbc<N>` format (e.g. `v3.6.4-gbc1`) for fork releases

## Architecture

```
kernel/              Go backend (HTTP server, API, data model, search, sync)
  server/tunnel/       Cloudflared + Tailscale tunnel support (in-process)
  api/tunnel.go        Tunnel management API endpoints
  conf/tunnel.go       Tunnel configuration structs
app/                 TypeScript/Electron frontend (editor, UI, themes, i18n)
android/             Android app (git subtree from siyuan-android)
ios/                 iOS app (git subtree from siyuan-ios)
third_party/
  cloudflared/         Submodule: gt4o4/cloudflared (patched for in-process embedding)
patches/             Unified diff patches applied to upstream during CI
  siyuan/              4 patches: disable-update, default-config, mock-vip-user, hide-account-entry
  siyuan-android/      1 patch: debug-build (custom signing config)
scripts/             Build scripts (linux-build.sh, darwin-build.sh, win-build.bat)
.github/workflows/   CI pipelines (desktop, android, iOS, docker, cron)
```

### Cloudflared / Tailscale Tunnel Embedding

The fork embeds cloudflared and tailscale tsnet in-process (no external binaries). Key `replace` directives in `kernel/go.mod`:

| Replace | Target | Reason |
|---------|--------|--------|
| `cloudflare/cloudflared` | `../third_party/cloudflared` | Patched fork with `cmd/cloudflared` API for in-process use |
| `quic-go/quic-go` | `chungthuang/quic-go v0.45.1` | Cloudflared needs old `logging/` package + interface API |
| `imroc/req/v3` | `v3.53.0` | Last version compatible with quic-go v0.45 (v3.54+ needs struct API) |
| `quic-go/qpack` | `v0.5.1` | req v3.53 uses `NewDecoder(callback)` + `DecodeFull`, removed in v0.6.0 |
| `urfave/cli/v2` | `ipostelnik/cli/v2` | Cloudflared needs `ApplyInputSource` not in upstream urfave/cli |

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
docker build -t siyuan-shared .
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
- **hide-account-entry.patch**: Removes account/login UI entry points (top bar, sync settings)
- **debug-build.patch**: Replaces Android signing config with env-based keystore (`KEYSTORE_PASSWORD`)

## Upstream Sync Workflow

1. `git fetch siyuan && git merge siyuan/master` — full-repo merge, resolve go.mod conflicts keeping our deps + upstream's newer versions
2. `git fetch upstream && git merge upstream/master` — merge appdev/siyuan-unlock, keep our CI workflow
3. `git subtree pull --prefix=android siyuan-android main` — subtree merge into `android/` (no `--squash`)
4. Verify all patches apply against the new version
5. Update patches if needed (regenerate diffs against the new upstream tag)
6. Tag as `v<upstream>-gbc<N>` and trigger build: `gh workflow run desktop-release.yml --ref <tag> -f version=<tag> -f packageManager=pnpm@latest`

## Version Convention

The fork uses version prefix `103.x.y` (e.g., `103.6.0`) in `kernel/util/working.go` and `app/appx/AppxManifest.xml` to distinguish from official builds. The `app/package.json` version tracks upstream (e.g., `3.6.0`).

## Key Technologies

- **Go 1.26+** with SQLite FTS5 for the kernel
- **pnpm** / Node / Electron / Webpack 5 / TypeScript for the frontend
- **Gradle / Android SDK 36** for Android builds
- **goreleaser-cross** container for cross-compiling kernel (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)

## Submodules

| Submodule | Path | URL |
|-----------|------|-----|
| cloudflared | `third_party/cloudflared` | github.com/gt4o4/cloudflared |

After cloning, run `git submodule update --init --recursive` to fetch submodules.

---

# Upstream SiYuan guide (siyuan-note/siyuan)

> Upstream's AGENTS.md follows verbatim, kept in sync via siyuan merges. Where it conflicts with the fork guide above, the fork guide wins.

SiYuan repository guide. Module path `github.com/siyuan-note/siyuan`, license AGPL-3.0.

---

## 1. Non-negotiable constraints

### Do not hand-edit

- `app/stage/protyle/js/lute/lute.min.js` (built from upstream `88250/lute`)
- `app/stage/build/**`, `app/src/types/dist/**`
- `app/changelogs/**` (generated by separate tooling)
- `app/kernel/SiYuan-Kernel*`, `*.syso`, `kernel/kernel.aar`
- `app/pandoc/*`

### Verification and prohibited operations

1. **Frontend verification:** Do not use `npx webpack` or `pnpm dev` to verify changes; after changes, run `cd app && pnpm run lint` to check code style
2. **Frontend build:** Do NOT run `pnpm build` — the developer runs `pnpm dev` manually, and `pnpm build` will conflict with it, producing broken bundles
3. **Kernel development:** After modifying Go code, run `gofmt`, but do not compile the kernel binary or restart a running kernel; the developer handles both manually
4. **Git:** **NEVER** run `git commit` / `git push` unless explicitly asked — no exceptions

---

## 2. Project-specific rules

1. **i18n:**
   - New keys go at the **top** of each `langs/*.json` object; add to every language file (reference `en.json`)
   - Indent `langs/*.json` with tabs, using one tab per nesting level; do not use spaces for indentation
   - Exception: inside the `_kernel` object, append new entries at the **end** using the next incremental numeric key
   - Each language must be properly translated — do NOT copy the same text across all language files
   - Use three ASCII periods (`...`) for ellipses in all localized strings; do not use Unicode ellipsis characters (`…` or `……`)
   - Setting description tip strings must not end with a period or equivalent sentence-ending mark (for example `.`, `。`, or `।`)
   - Domains: `ld246.com` only in `zh-CN.json`; use `liuyun.io` in all other languages
   - In `zh-TW` localization and the Traditional Chinese user guide, translate SiYuan's content-model term Block as `區塊`; never abbreviate it as `塊`
   - Use `區塊` consistently in compound terms, for example `內容區塊`, `子區塊`, `父區塊`, `嵌入區塊`, `程式碼區塊`, `區塊 ID`, `區塊標`, and `區塊級`
   - Translate Block Reference as `區塊引用` and Blockquote as `引述區塊`; do not confuse them or reverse the word order
   - When counting content blocks, use `個區塊` rather than using `塊` as a classifier or abbreviation
   - Do not mechanically replace `塊` in ordinary words with `區塊`; preserve non-content-block terms such as `分塊` for data chunks and `覈取方塊` for Checkbox
   - Keep block terminology consistent between the Traditional Chinese interface and user guide
   - After modifying i18n files, run `python scripts/check-lang-keys.py` to verify key completeness across all language files
2. **Windows scripting:** Prefer Node.js / Python; avoid PowerShell unless necessary
3. **Icons:** Do not hand-write SVG; use existing icons from `app/appearance/icons/litheness/icon.js` when possible
4. **User guide:** When editing the user guide, follow `docs/SY-FORMAT.md`
   - When a feature adds or changes shortcuts, update the shortcut documentation in the user guide in the same change; if the appropriate section is unclear, ask the user where it should be placed
   - Represent in-app UI navigation paths as segmented `kbd` text marks: use one `NodeTextMark` with `TextMarkType: "kbd"` per navigation level, and place a plain `NodeText` containing ` - ` between adjacent levels
   - When a `kbd` path is embedded in prose, use exactly one ASCII space outside the path on each side when adjacent ordinary text exists; do not add an outer space at the start or end of a block
   - Omit the left outer space when the first `kbd` immediately follows full-width punctuation (for example, `，` or `、`); apply this rule to every language, including Chinese and Japanese, but do not apply it to half-width punctuation
   - Omit the right outer space when `kbd` is immediately followed by punctuation, whether full-width or half-width; keep the internal ` - ` separators of segmented UI paths unchanged
5. **Git:**
   - When explicitly asked to commit, follow the style of recent commits (gitmoji prefix + subject, in English)
   - Append the full issue/PR URL to the end of the commit title (e.g. `https://github.com/siyuan-note/siyuan/issues/<NNN>`, not the `#NNN` short form — it is clickable) only when a related issue exists; never put the URL in the commit body, and do not fabricate one
6. **GitHub:** Prefer the GitHub CLI (`gh`) for all GitHub operations, including reading issues, comments, pull requests, commits, statuses, and metadata. If `gh` is unavailable or does not support the operation, fall back to the GitHub API or web interface
   - On Windows, when creating or updating GitHub text that contains non-ASCII characters, write the request payload to a UTF-8 JSON file and call the GitHub API with `gh api --input <file>`; do not pipe the text through PowerShell because its encoding may corrupt the content. Verify the published content and remove the temporary file afterward
7. **Issue titles:** Whenever the user asks to generate an issue title, provide it in English regardless of the wording of the request, and do not start it with `Fix`
   - If the issue is labeled `Bug`, objectively describe the problem or symptom instead of writing from a bug-fix perspective
   - If the issue is labeled `Enhancement`:
     - For improvements to existing functionality, write the title from an improvement perspective and prefer `Improve ...`
     - For capabilities that did not previously exist, write the title from a support perspective and prefer `Support ...`
   - If no applicable label is available, infer the perspective from the issue content
8. **LD246:** When accessing `ld246.com`, set the HTTP `User-Agent` header to `SiYuan-Coding-Agent`
9. **Configurable entries:**
   - Treat the `data-id` of a configurable desktop menu item and the `data-type` of a configurable dock entry as persisted configuration identifiers. Do not rename or reuse them unless the same change migrates existing visibility and order configuration
   - When adding, removing, renaming, or moving a configurable desktop menu item or dock entry, or changing its `data-id` / `data-type`, update `app/src/config/entryVisibility/catalog.ts` in the same change, including its type, hierarchy, label, Simple profile default, and default position, and update the related tests
   - Give every configurable desktop menu separator a stable `data-id` and register it in the catalog as a `separator`. Keep the catalog order aligned with the actual menu declaration order because it defines the built-in order and where new entries are merged into existing custom profiles
   - Keep parent and child paths aligned with the actual menu hierarchy. Dock entries support visibility only and must not be included in sorting
   - Cover catalog consistency, separator placement, order migration, and plugin-slot preservation in the related tests. Configured menus must not produce leading, trailing, or consecutive separators
   - The menu `ignore` option controls conditional rendering and must not be used to opt an entry out of visibility or order configuration

---

## 3. Coding conventions

1. **Comments:** Wrap code comments at 120 characters
2. **Comments:** Describe what the code does, not what it replaced — don't reference the old implementation in comments
3. **Comments:** Write comments in Chinese
4. **Punctuation:** Use language-appropriate punctuation (e.g. Chinese punctuation ，。：；！？「」 for Chinese, not ASCII); do not hard-code it in code — put it in the i18n language files so each locale renders its own. Applies to comments, user guide, `.md` docs, etc.
5. **UI paths:** In all contexts, including code comments, UI text, i18n, user guides, documentation, issue/PR content, and responses, separate navigation levels with a hyphen surrounded by spaces (for example, `设置 - 快捷键 - 通用`); do not use arrow symbols such as `→`
6. **Markdown:** Do not hand-wrap; keep each line (paragraphs, table rows, list items, etc.) on a single line
7. **TypeScript/JavaScript:** Semicolons required, use double quotes, indent with spaces
8. **CSS:** Do not use the `:has()` selector because of its performance impact
9. **Go:** Format with `gofmt` after editing

---

## 4. Required toolchain

| Tool | Version | Source of truth |
|---|---|---|
| Go | see `go` directive | `kernel/go.mod` |
| Node (+ pnpm) | see CI matrix | `.github/workflows/cd.yml`, `app/package.json` (`packageManager` field) |

---

## 5. Repository layout

**Architecture:** Go kernel (`kernel/`) + TypeScript frontend (`app/`), plus a separate `export` bundle (global `Protyle`, entry `src/protyle/method.ts`) for rendering rich content in exported HTML / PDF preview. Read versions from `kernel/go.mod`, `app/package.json`, `kernel/util/working.go`.

Top level (repo root):

| Path | Contents |
|---|---|
| `kernel/` | Go backend — server, data engine, API, all domain logic |
| `app/` | TypeScript frontend (Electron/web), built by webpack into `app/stage/build/` |
| `app/appearance/` | Themes, icons, **i18n** (`appearance/langs/*.json`) |
| `app/stage/` | Build output served by the kernel |
| `app/changelogs/` | Per-version changelog markdown |
| `.github/` | `CONTRIBUTING.md` (+zh-CN), `SECURITY.md`, `CODE_OF_CONDUCT.md`, `PULL_REQUEST_TEMPLATE.md`, issue templates, `workflows/` |
| `scripts/` | Release packaging: `win-build.bat`, `darwin-build.sh`, `linux-build.sh`, `parse-changelog.py`, `check-lang-keys.py` |

### Major `kernel/` packages (under `kernel/`)

| Package | Responsibility |
|---|---|
| `main.go` (`//go:build !mobile`) | Desktop entry point → `cli/cmd` |
| `cli/cmd/` | Cobra CLI subcommands (`serve`, `notebook`, `block`, `search`, `sql`, `export`, `repo`, `sync`, …) |
| `model/` | **Core domain** (~70 files): blocks/trees, transactions, indexing, search, attribute views, export, history, sync, flashcards, AI, CalDAV/CardDAV, auth |
| `treenode/` | In-memory tree over the Lute AST + `blocktree.db` (`BlockTree{ID,RootID,ParentID,BoxID,Path,HPath,Type,...}`) |
| `av/` | **Attribute View** (database) engine: values, filters, sorts, layouts (table/kanban/gallery) |
| `sql/` | **Embedded SQLite** (`siyuan.db`, `history.db`, `asset_content.db`) + FTS5; async index queues |
| `search/` | FTS tokenizer helpers, CJK conversion (`hanconv.go`) |
| `bazaar/` | Marketplace: plugins/widgets/themes/icons/templates |
| `filesys/` | Read/write `.sy` files on disk (via `filelock`) |
| `server/` | Gin server bootstrap (`serve.go`): middleware, TLS/cmux, WebDAV/CalDAV/CardDAV, WebSocket, MCP |
| `api/` | HTTP route registration (`router.go::ServeAPI`, ~400 endpoints) + per-area handlers |
| `conf/` | Configuration structs |
| `util/` | Cross-cutting: `working.go` (workspace, `Boot()`), `lute.go`, `i18n.go`, `websocket.go` (melody push), `result.go` (API envelope) |
| `plugin/` | Plugin subsystem (kernel side) |
| `mcp/` | MCP (Model Context Protocol) server |
| `agent/` | AI agent runtime |
| `mobile/`, `harmony/` | `//go:build mobile` gomobile bindings for Android/iOS/HarmonyOS |

### Frontend (`app/src/`) highlights

| Dir | Purpose |
|---|---|
| `index.ts` | Main `App` class — boots SPA, opens main WebSocket, handles WS push events |
| `window/` | Detached Electron window variant |
| `protyle/` | **The block editor** — `wysiwyg/`, `toolbar/`, `gutter/`, `breadcrumb/`, `hint/`, `scroll/`, `undo/`, `preview/`, `render/` (incl. `render/av/`) |
| `editor/`, `layout/`, `menus/`, `dialog/`, `config/`, `mobile/`, `ai/`, `sync/`, `history/`, `search/`, `card/` | Feature modules |
| `util/fetch.ts` | `fetchGet`/`fetchPost` — all kernel calls |
| `layout/Model.ts` | WebSocket client all UI binds to |
| `constants.ts` | Global constants (version, IDs, storage keys) |

Four webpack configs each emit a separate bundle to `app/stage/build/{app,desktop,mobile,export}/`. The kernel's `serveAppearance` picks which bundle to serve based on User-Agent. The `export` bundle is different from the other three: it is not an app UI — it is a client-side library (global `Protyle`, entry `src/protyle/method.ts`) exposing renderers for code highlighting, math (KaTeX), and diagrams (Mermaid/flowchart/graphviz/…). It is loaded by the HTML pages assembled during export (`app/src/protyle/export/index.ts`) — the desktop PDF preview window and standalone exported HTML files — so rich content renders outside the editor.

---

## 6. Related repositories (navigation)

SiYuan spans several repos. This repo (`siyuan`) holds the kernel + Electron/web frontend; the others are separate projects with their own tooling.

| Repo | Role / what to know |
|---|---|
| `siyuan` | **This repo** — kernel + Electron/web/tablet UI |
| `siyuan-android` / `siyuan-ios` / `siyuan-harmony` | Native apps wrapping the gomobile kernel; build steps differ per platform — see each project's README |
| `siyuan-chrome` | Browser extension (web clipper); talks to the running kernel over HTTP only |
| `siyuan-testing` | Playwright end-to-end tests for a running SiYuan instance; test data belongs in the `SiYuan Testing` notebook — see that repository's `AGENTS.md` |
| `petal` | SiYuan Plugin API declaration (the plugin system is named "petal"); consumed by plugins, not a kernel Go dependency |
| `lute` | Markdown/Kramdown AST engine — the editor + `.sy` format; also the source of the bundled `lute.min.js` (a GopherJS build served to the frontend). **Lives under `$GOPATH/src/github.com/88250/lute`, not as a sibling repo** |
| `dejavu` | Data repo / sync engine (encrypted snapshots) |
| `riff` | Spaced-repetition (SRS) flashcard scheduler |
| `gulu` | General Go utility library (`gulu.Ret`, `gulu.JSON`, …) |
| `eventbus` | In-process event bus |
| `filelock` | Cross-platform file locking (`.sy` read/write) |
| `httpclient` | HTTP client wrapper (cloud / sync / bazaar calls) |
| `logging` | Leveled logging used throughout the kernel |
| `go-sqlite3` / `pdfcpu` | Maintainer's forks, pulled in via permanent `replace` in `kernel/go.mod` (keep those) |
| `epub` / `clipboard` / `go-humanize` / `vitess-sqlparser` / `dataparser` / `encryption` | Smaller Go libraries (export / clipboard / formatting / SQL parse / data parse / crypto) |

All Go libraries above are dependencies in `kernel/go.mod`. GitHub org: `siyuan-note/*` for the `siyuan-` apps and most libs; `88250/*` for lute, gulu, and the forks (go-sqlite3 / pdfcpu).

### Cross-repo notes

- **Editing any Go dependency (Lute / dejavu / gulu / eventbus / riff / filelock / httpclient / logging / go-sqlite3 / pdfcpu / epub / …):** these are imported by the kernel as Go modules (`kernel/go.mod`). To test a local change, add a temporary `replace` in `kernel/go.mod` pointing at your local checkout — but **never commit that `replace`**; it breaks builds for everyone else.
- **Rebuilding `lute.min.js`:** it's the JS build of the Go `lute` project — generated upstream and checked into `app/stage/protyle/js/lute/`. Don't edit it here; change `lute`, rebuild, and copy the artifact in.
- **Mobile apps (`siyuan-android` / `siyuan-ios` / `siyuan-harmony`):** each is a separate native app that wraps the kernel built from this repo. For how to build, vendor the kernel binding, and wire everything up, **read each project's own README** — the toolchains and steps differ per platform and aren't documented here.
- **`siyuan-chrome`:** independent TypeScript project; it only interacts with a running SiYuan instance through the public HTTP API documented in `docs/API.md`.

---

## 7. Response style

1. **Language:** Match the user's language; do not mix languages mid-sentence (keep proper nouns / identifiers in their original form)
