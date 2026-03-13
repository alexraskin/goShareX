# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoShareX is a ShareX file uploader service built with TinyGo, compiled to WebAssembly, and deployed on Cloudflare Workers. It uses Cloudflare R2 for object storage and Cloudflare Cache API for caching served files.

## Build & Development Commands

```bash
# Build (compiles Go to WASM via TinyGo + generates worker assets)
npm run build

# Local development server
npm run dev        # or: npm run start

# Deploy to Cloudflare Workers
npm run deploy
```

The build process runs two steps: `workers-assets-gen` (generates worker.mjs entry point) then `tinygo build` targeting WASM. Output goes to `./build/`.

**Prerequisites**: TinyGo compiler, Node.js, wrangler CLI. Local dev requires a `.dev.vars` file with `SHAREX_AUTH_KEY` (see `.dev.vars.example`).

There are no tests in this project.

## Architecture

**Entry point**: `main.go` — creates the Server struct with auth key and R2 bucket name from Cloudflare bindings, then serves via `syumai/workers`.

**HTTP handlers** in `server/`:
- `server.go` — routing (sync.Once-initialized ServeMux), auth middleware, error helpers, R2 bucket access
- `upload_handler.go` — POST `/upload` — generates 6-char random IDs, maps MIME types to extensions, stores in R2
- `delete_handler.go` — GET `/delete` — removes from R2, async cache purge via WaitUntil
- `key_handler.go` — GET `/{key}` — serves files with Cloudflare Cache API layer (7-day TTL)
- `config_handler.go` — GET `/config` — generates ShareX .sxcu config file
- `stats_handler.go` — GET `/stats` — reports TinyGo version, memory usage, R2 object count

All endpoints except `GET /{key}` require `authKey` query parameter authentication.

## Key Dependencies

- **syumai/workers** — Go bindings for Cloudflare Workers runtime (HTTP serving, R2, Cache API, env vars)
- **wrangler** — Cloudflare Workers CLI for dev/deploy
- **TinyGo** — compiles Go to WASM (not standard Go compiler)

## Cloudflare Bindings

Configured in `wrangler.jsonc`:
- `SHAREX_AUTH_KEY` — environment variable (set as secret for production)
- `IMAGE_BUCKET` — R2 bucket binding for file storage
