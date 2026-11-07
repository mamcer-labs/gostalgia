# Gostalgia

A media archive scanner and API in Go: it walks a directory tree,
deduplicates files by SHA-256 into a single managed copy, extracts
metadata (EXIF, ID3, ZIP listings, small text files) and serves it over
a REST API.

> **Status: retired as a data system (November 2026).** Gostalgia did
> its real job: it built the deduplicated copy of a ~1.2 TB family
> archive. It no longer scans or copies anything. The archive is now
> served by mature tools (see [What replaced it](#what-replaced-it)), and
> gostalgia stays deployed only as the **pilot app** of a homelab
> platform, with sample data. No new features are planned.

## What it does

- **Scan** (`gostalgia-cli scan`): walks `NOSTALGIA_SCAN_PATH` with a
  worker pool, hashes every file with SHA-256 and **copies** each unique
  file into `NOSTALGIA_HOME_PATH`. Every location where a file was seen
  is kept as a node (`nfile` + `nfilenode`), so one copy can have many
  original paths.
- **Enrich** (`gostalgia-cli enrich`): parallel metadata extraction:
  EXIF for images, ID3 for audio, file listings for ZIP archives, and
  the content of text files under 100 KB. Thumbnails through ImageMagick.
- **API** (`gostalgia-api`): Gin REST API over the catalog (directories,
  files, tags, scans, search by name), with Swagger docs in `docs/`.
- **Web** (`web/`): a Svelte UI, never deployed (see below).

## What it was used for

Between 2025 and 2026, gostalgia (and its C# predecessor, nostalgia)
scanned decades of CDs, DVDs and disk backups into one deduplicated tree
of about 250,000 files. That tree is still the canonical copy of the
archive.

Two lessons from that run, both fixed outside gostalgia:

- **The copy did not preserve modification times.** Every file ended up
  dated on the day it was scanned. The original dates were recovered
  later by matching identical files against the original backups.
- **Copying was the wrong model.** It doubled the storage and left the
  originals to clean up afterwards. Indexing in place, with a read-only
  mount, would have avoided both.

## What is unfinished

Honest list, from a final code review:

- **No in-place indexing:** a scan always copies; there is no
  catalog-only mode.
- **Scans are not incremental:** every scan re-hashes every file.
  Re-scanning the same tree duplicates directory rows.
- **One bad file aborts a scan** (e.g. a permission error), and a scan
  is marked completed before its data is persisted. There are no real
  transactions.
- **`--tags` is accepted but never saved.**
- **The enricher opens relative paths**, so outside `NOSTALGIA_HOME_PATH`
  every open fails silently.
- **Search is `LIKE` on file names only.** The extracted text, EXIF and
  ID3 data are stored but never searchable, despite the original goal.
- **No authentication.** CORS combines `*` with credentials, and a
  `/v1/debug/fail` endpoint is exposed.
- **The web UI has the API URL hardcoded** to `http://localhost:8080`.
- **`Dockerfile.cli` is broken:** it builds `./cmd/nostalgia-cli`, but
  the directory is `cmd/gostalgia-cli`.
- Configuration still uses the old `NOSTALGIA_*` prefix.

## What it does well

The point of the 2026 rewrite was to learn Go on a real system, and
that part worked:

- Layered architecture (`internal/domain`, `internal/app`,
  `internal/infra`) with generic repositories and a filesystem interface
  that can be mocked in tests.
- Concurrency with worker pools for scanning and enrichment.
- Tests with `-race`, including integration tests against a real MySQL
  through testcontainers (`go test ./...` passes).
- Observability: OpenTelemetry tracing, RED metrics, structured logs
  correlated by `trace_id`/`span_id`, and a latency SLO with multi-window
  burn-rate alerts ([`docs/SLO.md`](docs/SLO.md)).
- Embedded goose migrations, graceful shutdown, CI on a self-hosted
  runner, and deployment to K3s through the platform's `pmf deploy`,
  with secrets from Vault.

## What replaced it

| Need | Tool |
|---|---|
| Music | [Navidrome](https://www.navidrome.org/) |
| Photos, videos, phone backup | [Immich](https://immich.app/) |
| Search across documents and archives | [sist2](https://github.com/simon987/sist2) |
| One-off deduplication | [fclones](https://github.com/pkolaczk/fclones) |

Building a lower-quality clone of those in hobby time would have kept the
archive disorganized for years. The deduplication gostalgia did well is
a one-time task, and it is done.

## Project history

- **2020:** first hand-written Go version, to organize personal media and
  learn Go from the official docs. Tag `v0.1.0-manual`.
- **2025:** rewritten in .NET to explore enterprise patterns and Entity
  Framework ([nostalgia](https://github.com/mamcer/nostalgia)).
- **2026:** rewritten in Go again, co-engineered with Gemini CLI, focused
  on concurrency and observability. Tag `v1.0.0-agentic`.
- **November 2026:** retired as a data system; kept as a platform pilot.

## Running it

Requires Go 1.25 and MySQL 8 (SQLite is used only in tests). Copy
`.env.example` to `.env` and set the connection string and paths.

```bash
go test ./...
go build -o gostalgia-cli ./cmd/gostalgia-cli
./gostalgia-cli scan --source "MyMemories"
./gostalgia-cli enrich --workers 8 --batch 500
```

API with Docker:

```bash
docker build -f Dockerfile.api -t gostalgia-api:local .
docker run --rm -d --name gostalgia-api -p 5001:8080 --env-file .env gostalgia-api:local
```

## Contact

- **GitHub:** [github.com/mamcer](https://github.com/mamcer)
- **Email:** mamcer@proton.com
