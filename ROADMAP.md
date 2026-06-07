# Fieldstone Roadmap: Self-Hostable Supabase-in-Go

## Identity Pivot
Fieldstone reframes from "Go PocketBase (collections)" to **"self-hostable Supabase-in-Go"** — a gateway over real Postgres, reusing Postgres extensions (RLS, pg_cron, pgvector, pgmq) instead of reinventing them.

## Phase 0 — Foundation (the 60% unlock)

### 1. Auto-REST API over real Postgres tables (PostgREST-equivalent)
- [x] Query gateway: `GET /api/v1/{table}` with `select`, `eq`, `order`, `range`, `limit`
- [x] Mutations: `POST`, `PATCH`, `DELETE`
- [x] Introspection via `information_schema`
- [x] Parameterized SQL with identifier sanitization

### 2. RLS via Postgres, driven by Fieldstone JWT
- [x] `SET LOCAL request.jwt.claims` per connection (Supabase-compatible)
- [x] `SET LOCAL app.current_tenant` for tenant isolation
- [x] JWT claims include `sub`, `role`, `tenant_id`, `email`, `app_metadata`
- [x] Auto-generated RLS policy helpers in pack SQL

### 3. Auth breadth
- [x] OTP send/verify (pluggable SMS hook)
- [x] OAuth redirect/callback stubs
- [x] Magic link
- [x] Password reset
- [x] JWT/bcrypt/WebAuthn retained
- [x] Refresh tokens + token rotation

## Phase 1 — Expose the engines you already have

### 4. Storage (buckets + signed URLs + RLS)
- [x] Local filesystem backend
- [x] S3 backend stub
- [x] HTTP routes: create/list/delete buckets, upload/download/list objects
- [x] Public download endpoint with signed URL support

### 5. Realtime (/ws channels + change feed)
- [x] WebSocket hub wired to `/ws`
- [x] Room-based subscriptions per collection/table
- [x] Postgres LISTEN/NOTIFY ready (CDC trigger scaffolding in packs)

## Phase 2 — Server logic

### 6. Cron
- [x] In-process scheduler with `@every` support
- [x] pg_cron delegation when available (`SELECT cron.schedule(...)`)

### 7. Functions host
- [x] `/functions/v1/*` reverse proxy to Deno/Node sidecar
- [x] Keeps WASM plugins for pure transforms

### 8. Webhooks
- [x] Fix engine syntax bugs
- [x] In-memory store + delivery store
- [x] HTTP CRUD routes for webhook configs
- [x] HMAC signatures + retry logic

## Phase 3 — Thin per-vertical packs

### Healthcare (100xLongevity)
- [x] Schema: `users`, `consents`, `health_metrics`, `consultation_slots`, `consultations`, `audit_log`, `device_tokens`, `brain_checkins`, `sleep_sessions`, `metric_baselines`, `scores`, `biological_ages`, `score_weights`
- [x] RLS: patient self-access, physician active-consult read, admin read
- [x] Audit log immutable append-only
- [x] Auto-create profile trigger

### SaaS
- [x] Schema: `orgs`, `org_memberships`, `subscriptions`, `invoices`, `usage_records`, `saas_audit`
- [x] RLS: member read, owner/admin write

### Marketing (ChampIQ)
- [x] Schema: `companies`, `prospects`, `signals`, `campaigns`, `sequences`, `engagements`, `suppression`, `sending_domains`
- [x] Fit scores + intent bands
- [x] Consent-gated email/voice flags

### LakeB2B
- [x] Schema: `bulk_uploads`, `entities`, `dedup_rules`, `enrichments`, `entity_relations`, `cdc_events`
- [x] Entity resolution + merge graph scaffolding

## Cross-cutting — SDK

- [x] `@champions/client` TypeScript SDK
- [x] `createClient(url, key)` factory
- [x] `auth.signUp/signInWithPassword/signInWithOtp/verifyOtp/signOut`
- [x] `from(table).select().eq().order().limit().insert().update().delete()`
- [x] `storage.from(bucket).upload/download/remove/createSignedUrl`
- [x] `channel(name).on().subscribe()`
- [x] `functions.invoke(name, { body })`
- [x] `rpc(fn, params)`

## Extras (70→90)
- [ ] pgvector (AI/RAG)
- [ ] pgmq queues
- [ ] GraphQL gateway (pg_graphql)
- [ ] PostGIS
- [ ] Edge-global latency (CDN)

## How to run

```bash
# Postgres backend
go run ./cmd/server -backend postgres -dsn "postgres://user:pass@localhost/fieldstone?sslmode=disable"

# Apply a vertical pack
psql $DSN -f packs/healthcare/schema.sql
psql $DSN -f packs/healthcare/rls.sql
```

## Coverage estimate
- After Phase 0+1: ~55–65% for all four verticals (data + RLS + auth + storage + realtime)
- After Phase 2 (functions + cron): ~75% healthcare, ~80% SaaS, ~75% marketing
- After Phase 3 packs: 70–85% across the board
