# Timepad Architectural Review & Improvements

This document outlines key technical debts and systemic weaknesses identified during the implementation of the observer telemetry services, accompanied by suggested remediation paths.

## 1. Synchronous Telemetry Ingestion Flow

**Status: ✅ Resolved**

**The Problem**:
Previously, `EventsService.IngestEvents` synchronously committed events to PostgreSQL before responding, coupling client latency to DB I/O.

**Implementation**:

- `POST /events` validates the device and filters events, then enqueues the payload onto a Redis list (`timepad:ingest_queue`) and returns **HTTP 202 Accepted** immediately.
- A `StartIngestWorker(ctx)` goroutine (started at server boot) runs a `BRPOP` loop, pops jobs, and calls `processEvents` which does `CreateInBatches(100)`, auto-categorization, and `LastSeenAt` update.
- **Graceful degradation**: if Redis is unreachable at startup, `rdb` is `nil`, the worker is a no-op, and the server falls back to synchronous inserts (HTTP 201) transparently.

## 2. Lack of Pagination on Timeline Retrieval

**Status: ✅ Resolved**

**The Problem**:
Previously, `GetTimeline` fetched an entire day's events in one unbound query — a potential OOM risk for power users.

**Implementation**:

- `GetTimeline` now accepts `cursor` (opaque base64-encoded RFC3339Nano timestamp) and `limit` (default 100, max 500) query params.
- Internally adds `WHERE start_time > cursorTime` and fetches `limit + 1` rows to detect the next page.
- Response shape: `{ "events": [...], "next_cursor": "<base64>" }` — `next_cursor` is omitted on the last page.

## 3. String-based Error Handlers

**Status: ✅ Resolved**

**The Problem**:
Previously, raw `err.Error()` strings (including potential SQL driver messages) were sent directly to clients.

**Implementation**:

- `utils.AppError` is a typed struct carrying an HTTP status code and a safe user-facing message.
- Constructor helpers: `NewNotFoundError`, `NewBadRequestError`, `NewConflictError`.
- `utils.HandleError(c, fallbackMsg, err)` uses `errors.As` to detect `*AppError` and maps it to the correct HTTP status; any non-`AppError` falls back to a generic **500** message — the raw error is only logged server-side.

## 4. Hardcoded Timezone Overrides

**Status: ✅ Resolved**

**The Problem**:
Previously, all summary and timeline date parsing used `time.Parse` (UTC), causing shifted day boundaries for non-UTC users.

**Implementation**:

- Each service (`SummaryService`, `ReportsService`, `EventsService`) has a `userLocation(userID) *time.Location` helper that fetches `User.Timezone` from the DB and calls `time.LoadLocation`; falls back to `time.UTC` on any error.
- All date parsing switched to `time.ParseInLocation("2006-01-02", date, loc)` and day boundaries constructed via `time.Date(y, m, d, 0, 0, 0, 0, loc)` instead of `Truncate(24h)`.
- Peak-hour bucketing in summary services uses `e.StartTime.In(loc).Hour()` rather than bare `.Hour()`.
