# Timepad Architectural Review & Improvements

This document outlines key technical debts and systemic weaknesses identified during the implementation of the observer telemetry services, accompanied by suggested remediation paths.

## 1. Synchronous Telemetry Ingestion Flow

**The Problem**:
Currently, `EventsService.IngestEvents` receives bulk event blocks and synchronously commits them to PostgreSQL before responding to the client device. This couples the frontend response latency directly to the database I/O performance. As devices scale or DB load increases, clients could face timeouts or battery drain.

**The Solution**:
Introduce an **Asynchronous Message Queue** (like Asynq, which is already referenced in `TECHNICAL_DOC.md` but not leveraged in the ingestion path).

1. The `/events` controller instantly acknowledges receipt (HTTP 202 Accepted) after dropping the payload onto a Redis queue.
2. Background workers pop items off the queue and batch-insert them into PostgreSQL at safe, controlled limits without blocking web threads.

## 2. Lack of Pagination on Timeline Retrieval

**The Problem**:
`GetTimeline` natively fetches an entire day's worth of `ActivityEvent` records. Power users tracking window focus changes might generate 10,000+ tiny events per day. Fetching, serializing, and transmitting this array as a single JSON block will cause Out-Of-Memory (OOM) errors on the Go server and severely crash the browser renderer trying to parse the massive payload.

**The Solution**:
Implement **Cursor-based Pagination**:

1. Change the API to return a limited set (e.g., `LIMIT 100`).
2. Include a `next_cursor` (an encrypted string holding the `id` or `timestamp` of the last record) to securely and quickly fetch the next batch using `WHERE start_time > ?`.

## 3. String-based Error Handlers

**The Problem**:
Most API controllers invoke utility responders (`utils.InternalServerError`) by explicitly parsing Go errors into raw strings (`err.Error()`). This creates a vulnerability where database query leaks or sensitive driver messages are dumped natively to the client interface.

**The Solution**:
Implement **Domain Errors**:

1. Wrap all `models` and `services` layer errors logically using custom typed structs (e.g., `ErrRecordNotFound()`, `ErrUnauthorized()`).
2. Have `utils.go` detect the type of error and translate internal flags to standard UX-friendly messages like "An unexpected system error occurred" while only logging the raw SQL dumps securely on the backend terminal.

## 4. Hardcoded Timezone Overrides

**The Problem**:
The `GetDailySummary` and `GetWeeklySummary` aggregation mechanisms force parsing inputs assuming the server's local timezone (truncating to 24-hour UTC blocks via `time.Parse`). The database itself runs universally in UTC. If a user in Tokyo requests their `2026-02-26` summary, they will receive data fundamentally shifted by 9 hours, polluting their dashboard.

**The Solution**:
Utilize the User's `Timezone` preference native to `UserSetting`:

1. Controllers must parse incoming dates mapping onto the specific location `time.LoadLocation(user.Timezone)`.
2. Push date-shifting calculations away from the Go application and directly into PostgreSQL via `AT TIME ZONE` SQL queries to accurately slice day boundaries dynamically per user context.
