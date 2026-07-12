# Gostalgia

A high-performance media management and discovery engine built in Go. Optimized for indexing, enriching, and organizing large personal archives.

## Project Evolution

Gostalgia reflects a journey through different architectural paradigms:

*   **2020 (The Beginning):** Originally a hand-crafted Go project created to solve personal media organization needs and mostly cement Golang knowledge from official documentation and the Go blue book. Tagged as `v0.1.0-manual`.
*   **2025 (The C# Era):** Migrated to .NET (C#) to explore modern Enterprise patterns and Entity Framework ([nostalgia](https://github.com/mamcer/nostalgia)).
*   **2026 (The Agentic Renaissance):** Completely re-imagined and rewritten in **Go** for maximum performance and modern observability. This version (`v1.0.0-agentic`) was co-engineered with **Gemini CLI**, leveraging agentic workflows to build a robust, highly parallelized infrastructure.

## Core Features

- **High-Performance Scanning:** Fast file system indexing designed for terabyte-scale storage.
- **Parallel Metadata Enrichment:** Multi-threaded worker pool to extract deep metadata without saturating system I/O.
- **Deep Content Extraction:**
    - **Images:** EXIF data (Date, GPS, Camera model).
    - **Audio:** ID3 tags (Artist, Album, Title, Year).
    - **Archives:** Internal file listing for ZIP archives.
    - **Text:** Full-content indexing for small documents (<100KB).
- **Temporal Heuristics:** Advanced logic to recover lost "date captured" information using filename patterns and directory structures—perfect for fixing metadata lost during file transfers.
- **Observability Ready:** Designed to integrate with Netdata, Prometheus, and Grafana.

## Getting Started

### Prerequisites
- **Go**: 1.22 or higher.
- **Database**: MySQL 8.0+ or SQLite.

### Installation
```bash
go build -o gostalgia-cli ./cmd/gostalgia-cli
```

### Usage

#### 1. Scan a Directory
```bash
./gostalgia-cli scan --source "MyMemories" --tags "Family;Vacation"
```

#### 2. Enrich Metadata
Use parallel workers to process your archive efficiently:
```bash
./gostalgia-cli enrich --workers 8 --batch 500
```

### local api

```bash
docker build -f Dockerfile.api -t gostalgia-api:local .
docker run --rm -d --name gostalgia-api -p 5001:8080 --env-file .env gostalgia-api:local
```

## Architecture

Gostalgia follows **Clean Architecture** principles, maintaining a strict separation between business logic (`internal/domain`), application use cases (`internal/app`), and infrastructure (`internal/infra`).

## Contact

- **GitHub:** [github.com/mamcer](https://github.com/mamcer)
- **Email:** mamcer@proton.com