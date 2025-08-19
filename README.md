# Lumium

Self-hosted photo warehouse - Go, Postgres, ClickHouse, NATS.

Keep your photos fast, organized, and private without shoveling them into someone else's cloud.

## Elevator pitch

Lumium ingests your photos, deduplicates exact and near-duplicates, and auto-builds albums from time + place. You get RAW/JPEG pairing, burst stacks, and "pick the best shot" workflows.

It's private by default (multi-tenant out of the box), runs on commodity hardware, and can push encrypted backups off-site if you want.

Own your photos. Keep the best. Reclaim space.

### Features

Lumium is under active development. Our roadmap is:

* Ingests millions of images without choking
* Finds duplicates and near-duplicates
* Groups trips and bursts automatically
* Lets you keep RAWs + JPEGs linked
* Multi-tenant (families, studios, clubs)
* Optional end-to-end encryption + encrypted cloud backup

## Stack

* `lm_api` - Go HTTP API
* `lm_pgsql` - Postgres for metadata
* `lm_clickhouse` - ClickHouse for analytics + search
* `lm_minio` - S3-compatible object storage
* `lm_nats` - Event bus for workers
* `lm_exifd` - EXIF extraction service
* `lm_phashd` - Perceptual hash service (near-dup detection)
* `lm_albumer` - Trip/stack/album builder
* `lm_replicator` - Encrypted off-site S3 replication
* `lm_scanner` - Folder/S3 watcher - kicks off ingest
* `lm_adminer` - Web DB client (Postgres)
* `lm_goconvey` - Test UI

## HTTP endpoints (dev defaults)

* API: 
* Adminer (Postgres): 
* ClickHouse UI: 
* MinIO console: 

Each service also exposes `/health`, `/ready` & `/whoami`.
