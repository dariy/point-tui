# Point TUI — Implementation Plan

## Context

The [Point engine](https://github.com/dariy/point/) is a self-hosted microblog written in 
Go (Echo v4 + SQLite). It cleanly separates a JSON HTTP API (under `/api/`) from 
a vanilla-JS SPA frontend. Because the API is self-contained and JSON-only, it can be driven by
any client.

This plan designs a **terminal UI reader** for Point: a single Go binary that browses a Point
instance's public content (posts, tags, timeline) and **renders photos as ANSI half-block art**
(the universal fallback `yazi` uses — no Kitty/Sixel/iTerm2 graphics protocol required). Default
target instance: `https://darii.net`.

**Scope (confirmed with user):** read-only reader. Anonymous browsing of public content; optional
password login to reveal private posts/media. No authoring, media management, or admin.

**Decisions (confirmed with user):**
- **Stack:** Go + Bubble Tea (matches the engine's language; single static binary).
- **Layout:** `yazi`-style 3-pane miller columns (Tags → Posts → Preview), vim keys.
- **Photos:** ANSI half-block rendering.

## API surface used (verified against `/point/api/cmd/api/main.go`)

All read endpoints below allow anonymous access (`OptionalAuthMiddleware`). Authenticating only
expands visibility (drafts, `is_public=0` media).

| Purpose | Method & path | Notes |
|---|---|---|
| Public blog config | `GET /api/settings/public` | title, subtitle, author, posts_per_page, theme |
| Home feed | `GET /api/pages/home?page&per_page` | compound: posts + settings + tags |
| List posts | `GET /api/posts?page&per_page&tag&search` | paginated; `{posts, total, pages}` |
| Post by slug / id | `GET /api/posts/slug/:slug`, `GET /api/posts/:id` | full body |
| Prev/next | `GET /api/posts/:id/navigation?tag` | `{previous, next}` |
| Tag tree | `GET /api/tags?include_empty` | hierarchical: each has `parents[]`, `children[]` |
| Posts in a tag | `GET /api/tags/slug/:slug/posts?page&per_page` | paginated |
| Timeline | `GET /api/timeline?context=<slug>` | year/count pills |
| Media file | `GET /:year/:month/:filename` | original bytes |
| Thumbnail | `GET /:year/:month/:filename?thumb` | smaller bytes — preferred for ANSI |
| Login (optional) | `POST /api/auth/login` | see auth note below |

**Post JSON shape** (from `postToResponse`, `api/internal/api/mappers.go:88`): `id, title, slug,
type, content` (markdown), `excerpt, status, published_at, created_at, media_url, meta_description,
tags[]` where each tag is `{name, slug}`. `media_url` is a normalized simplified path such as
`/2024/05/photo.jpg`; append `?thumb` for the thumbnail. Inline images live in `content` as markdown
`![alt](/YYYY/MM/file.jpg)` (regex in `mappers.go:19`).

**Auth note (optional login, verified `api/internal/api/auth.go:34-92`):** `POST /api/auth/login`
with body `{"username": <user>, "name": <sha256_hex_of_password>, "remember_me": <bool>}`. The
password is **SHA-256-hashed client-side to 64 hex chars** before sending (server compares hashes).
On success the server sets an HttpOnly `session` cookie — persist it in a `cookiejar` and send on
every request.

## Architecture

Single Go module, standard Bubble Tea (Elm) architecture: a root model holds three pane sub-models;
all network/image work happens in `tea.Cmd`s that emit `tea.Msg`s, keeping the update loop
non-blocking.

```
point-tui/
  go.mod
  Makefile
  PLAN.md
  cmd/point-tui/main.go        # flags (-l/--login), bootstrap, tea.NewProgram
  internal/
    config/config.go           # base URL, optional creds; ~/.config/point-tui/config.toml + env
    api/
      client.go                # http.Client + cookiejar, baseURL, doJSON, ctx timeouts, errors
      auth.go                  # Login: sha256(password) hex -> POST /api/auth/login
      models.go                # Post, Tag, Settings, TimelinePill, Paginated[T]
      posts.go                 # ListPosts, GetPostBySlug/ID, GetNavigation
      tags.go                  # ListTags -> tree; PostsByTag
      pages.go                 # GetHomePage; GetPublicSettings
      timeline.go              # GetTimeline
      media.go                 # FetchImage(path, thumb bool) []byte
    image/
      ansi.go                  # bytes -> ANSI half-block string sized to (cols,rows)
      cache.go                 # keyed by (path,cols,rows); LRU in-mem + raw-bytes disk cache
    ui/
      app.go                   # root model: layout, focus, resize, async wiring
      keys.go                  # vim keymap + bubbles/help
      styles.go                # lipgloss theme (borders, active-pane highlight)
      tags_pane.go             # left: tag tree + "All / Feed", "Timeline"
      posts_pane.go            # middle: post list (title + date), pagination
      preview_pane.go          # right: glamour markdown viewport + ANSI photo
      statusbar.go             # bottom: context, loading, errors, help hint
      messages.go              # tea.Msg types (postsLoaded, postLoaded, imageRendered, errMsg…)
```

**Dependencies (all reuse, nothing hand-rolled that a library covers):**
- `github.com/charmbracelet/bubbletea` — runtime
- `github.com/charmbracelet/bubbles` — `list`, `viewport`, `spinner`, `textinput`, `help`, `key`
- `github.com/charmbracelet/lipgloss` — layout/styling, `JoinHorizontal` for the 3 columns
- `github.com/charmbracelet/glamour` — render post markdown `content` to styled ANSI
- `github.com/eliukblau/pixterm/pkg/ansimage` — **ANSI half-block image renderer** (upper-half-block
  `▀` with fg=top pixel / bg=bottom pixel, optional dithering). This is exactly the protocol-free
  rendering `yazi` falls back to. `ansimage.NewScaledFromReader(r, rows*2, cols, bg, scaleMode,
  ditherMode)` then `.Render()` → ANSI string. (Fallback if undesired: `nfnt/resize` +
  `golang.org/x/image` and a ~30-line half-block encoder.)
- stdlib `net/http`, `net/http/cookiejar`, `crypto/sha256`, `image/jpeg`, `image/png`.

## UI design — miller columns

```
┌ Tags ────────┬ Posts ───────────────┬ Preview ───────────────┐
│ > All / Feed │ > Sunset over harbor │ ▀▀▀▀▀▀▀▀▀ (ANSI photo) │
│   Timeline   │   City lights        │ ▀▀▀▀▀▀▀▀▀               │
│   Photos   > │   Notes on autumn    │                        │
│   Travel     │   …                  │ # Sunset over harbor   │
│   Notes      │                      │ 2026-05-20 · #photos   │
│              │                      │ Body rendered via      │
│              │                      │ glamour…               │
└──────────────┴──────────────────────┴────────────────────────┘
 h/l ←→ pane · j/k move · enter open · / search · g/G top/bottom · r reload · q quit
```

- **Focus model:** exactly one pane focused (highlighted border). `h`/`l` (and `←`/`→`) move focus;
  `Tab` cycles. Selecting in Tags reloads Posts; selecting in Posts loads Preview.
- **Tags pane:** built from `GET /api/tags` hierarchy (`parents`/`children`). Two synthetic top
  entries — **All / Feed** (`/api/pages/home`) and **Timeline** (`/api/timeline`). Child tags shown
  indented; expand/collapse with `enter`/`l`.
- **Posts pane:** `bubbles/list` of `{title, published_at}`; lazy-loads next page on scroll past the
  end using `page`/`per_page` from public settings (`{posts,total,pages}`).
- **Preview pane:** `bubbles/viewport` (scrollable with `j/k`/`PgUp`/`PgDn`). Top = ANSI photo
  (chosen via T10), below = glamour-rendered `content`. Re-renders the photo on terminal resize so it
  fits the pane width.
- **Search:** `/` opens a `textinput`; submits to `GET /api/posts?search=` and shows results in the
  Posts pane.
- **Async:** every fetch/render is a `tea.Cmd`; spinner shows in the relevant pane until its
  `…Loaded`/`imageRendered` msg arrives. Errors surface in the status bar, never panic.

## ANSI photo rendering (the `yazi`-style core)

1. Determine the photo URL for the selected post (T10): prefer `media_url`, else first markdown image
   in `content`. Request the `?thumb` variant first (fast), original on demand.
2. `api.FetchImage` downloads bytes (cookie-aware for private media), cached on disk by path.
3. `image/ansi.go` decodes and renders to a half-block ANSI string sized to the preview pane
   (`cols`, `rows*2` source pixels — each cell = 2 vertical pixels). Account for terminal cell aspect
   (~2:1) so photos aren't vertically squashed.
4. Rendered string cached by `(path, cols, rows)`; invalidated on resize.

## Task breakdown (bd-ready — not created here)

Each task: **title**, intent, key files, dependencies, acceptance criteria.

### Milestone 1 — Scaffolding
- **T1 · Project scaffold & config.** Init module, add deps, create layout, `main.go` skeleton,
  config loader (positional URL default `https://darii.net`, `-l/--login`; env `POINT_TUI_BASE_URL`;
  optional `~/.config/point-tui/config.toml`). *Deps: none.* **AC:** `go build ./...` succeeds;
  `point-tui --help` prints flags; base URL resolves from flag/env/config with correct precedence.

### Milestone 2 — API client
- **T2 · HTTP client core.** `api/client.go`: `http.Client` with `cookiejar`, base URL join,
  `doJSON` (ctx + timeout, status→typed error, JSON decode). *Deps: T1.* **AC:** unit test with
  `httptest` covers 200 decode, 401/404/500 → typed errors, context cancel.
- **T3 · Domain models.** `api/models.go`: `Post, Tag (with Parents/Children), Settings,
  TimelinePill, Paginated[T]` matching engine JSON (`mappers.go`, `posts.go`). *Deps: T1.* **AC:**
  round-trip unmarshal of captured sample responses with no lost fields.
- **T4 · Posts endpoints.** `ListPosts(page,perPage,tag,search)`, `GetPostBySlug`, `GetPostByID`,
  `GetNavigation`. *Deps: T2,T3.* **AC:** httptest-backed tests for each; pagination fields parsed.
- **T5 · Tags endpoints + tree.** `ListTags` → build parent/child tree; `PostsByTag`. *Deps: T2,T3.*
  **AC:** given hierarchical fixture, tree has correct roots/children, no cycles/dupes.
- **T6 · Pages, settings, timeline.** `GetHomePage`, `GetPublicSettings`, `GetTimeline`. *Deps:
  T2,T3.* **AC:** home feed posts + posts_per_page parsed; timeline pills parsed.
- **T7 · Optional login.** `api/auth.go`: `sha256` hex of password → `POST /api/auth/login` with
  `{username, name, remember_me}`; persist `session` cookie in jar. *Deps: T2.* **AC:** sends
  64-char hex hash; cookie captured and replayed on a follow-up request (verified via httptest).
- **T8 · Media fetch + cache.** `FetchImage(path, thumb)`; disk cache of raw bytes. *Deps: T2.*
  **AC:** correct URL incl. `?thumb`; cache hit avoids second network call.

### Milestone 3 — ANSI image rendering
- **T9 · ANSI half-block renderer.** `image/ansi.go` (wrap `pixterm/ansimage`) + `image/cache.go`.
  Size to pane, correct aspect, JPEG/PNG. *Deps: T1.* **AC:** sample JPEG → non-empty ANSI using
  `▀` + SGR color; output width ≤ requested cols; cached by `(path,cols,rows)`.
- **T10 · Photo source selection.** Pick best image URL for a post (`media_url` → first markdown
  image). *Deps: T3.* **AC:** unit tests across thumbnail-set, inline-image-only, and no-image posts.

### Milestone 4 — UI core
- **T11 · Root model & miller layout.** `ui/app.go`, `ui/messages.go`: three panes via
  `lipgloss.JoinHorizontal`, focus state, `WindowSizeMsg` resize → recompute pane widths & trigger
  photo re-render. *Deps: T1.* **AC:** renders 3 bordered columns; resize reflows without artifacts;
  focused pane visibly highlighted.
- **T12 · Keybindings & help.** `ui/keys.go`: `h/j/k/l`, arrows, `Tab`, `enter`, `/`, `g/G`, `r`,
  `q`/`Ctrl-C`; `bubbles/help` bar. *Deps: T11.* **AC:** every binding triggers its action; help bar
  lists them; `q` exits cleanly restoring the terminal.
- **T13 · Styles/theme.** `ui/styles.go`. *Deps: T11.* **AC:** consistent borders/active highlight;
  degrades on non-truecolor terminals.

### Milestone 5 — Panes
- **T14 · Tags pane.** `ui/tags_pane.go`: synthetic All/Feed + Timeline, hierarchical tags,
  expand/collapse; selection emits a load-posts command. *Deps: T5,T6,T11,T12.* **AC:** tree renders
  with indentation; selecting a tag repopulates Posts; Timeline entry loads timeline context.
- **T15 · Posts pane.** `ui/posts_pane.go`: `bubbles/list` (title + date), lazy pagination, loading
  spinner, empty state. *Deps: T4,T6,T11,T12.* **AC:** lists a tag's/feed's posts; scrolling past end
  loads next page; selecting loads Preview.
- **T16 · Preview pane.** `ui/preview_pane.go`: `viewport` with ANSI photo on top + glamour markdown
  body; scrolls; re-renders photo on resize. *Deps: T4,T8,T9,T10,T11.* **AC:** opening a photo post
  shows a recognizable ANSI image then readable body text; scroll works; resize re-fits the photo.
- **T17 · Search & status bar.** `ui/statusbar.go` + `/` search via `textinput` → `search=` query.
  *Deps: T4,T11,T12.* **AC:** query returns matching posts; status bar shows context/loading/errors.

### Milestone 6 — Polish & ship
- **T18 · Async, errors, empty/loading states.** Audit all fetch/render paths go through `tea.Cmd`;
  uniform error → status bar; spinners everywhere. *Deps: M2–M5.* **AC:** unreachable host shows a
  clear error, not a crash; no blocking UI hangs.
- **T19 · CLI/config finalize.** `-l/--login` triggers password prompt (hidden input) → T7; finalize
  precedence & docs. *Deps: T7,T1.* **AC:** anonymous run works against `https://darii.net`;
  `-l/--login` reveals private content when creds valid.
- **T20 · Build, README, manual verification.** `Makefile`/optional goreleaser; README with usage;
  manual run against the live instance. *Deps: all.* **AC:** `make build` → single binary; README
  documents flags & keys; live browse + photo render confirmed.
- **T21 · Tests.** Unit tests for API client (httptest), tree builder, photo-source selection, ANSI
  renderer; smoke test of root model update. *Deps: M2–M5.* **AC:** `go test ./...` green; client and
  image packages have meaningful coverage.

## Verification (end-to-end)

1. **Build:** `cd point-tui && go build ./... && go test ./...`.
2. **Anonymous browse:** `./point-tui https://darii.net` — confirm Tags pane lists tags +
   All/Feed + Timeline; selecting a tag fills Posts; opening a post shows glamour body.
3. **ANSI photo:** open a photo-bearing post; confirm a recognizable half-block image renders in the
   Preview pane and re-fits when the terminal is resized. Cross-check the same post in a browser at
   `https://darii.net` to confirm fidelity.
4. **Fallback target:** repeat with `./point-tui https://point.darii.net`.
5. **Optional login:** `./point-tui https://darii.net -l/--login` — confirm SHA-256 hashing,
   `session` cookie persistence, and that private content becomes visible.
6. **Resilience:** point at an unreachable host → clear status-bar error, no panic.

## Notes / non-goals

- No writes of any kind (no create/edit/publish/upload/settings/admin). The full authoring API exists
  and could seed a future "compose" milestone, but is out of scope here.
- ANSI half-block is the only photo path (per requirement). Kitty/Sixel/iTerm2 protocols are
  explicitly **not** implemented; the `pixterm/ansimage` approach matches `yazi`'s universal fallback.
- Offline snapshot endpoints (`/api/system/offline/*`) are admin-only (auth required) and not used.
