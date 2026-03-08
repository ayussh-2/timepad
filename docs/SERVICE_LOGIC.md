# Timepad Service Logic & Data Flow

This document details the core algorithms, logic, and data flow mechanisms governing the observer services in the Timepad backend.

---

## 1. System Data Flow

The flow of telemetry data operates in a unidirectional pipeline from client devices up to aggregated reporting views on the frontend.

1.  **Collection Layer**: Background trackers (Windows/Android/Browser Extensions) natively poll the active window or application every few seconds.
2.  **Ingestion API (`POST /api/v1/events`)**:
    - Trackers flush batches of `EventInput` objects to the Go server via `EventsService.IngestEvents`.
    - Each batch is validated against a unique `DeviceKey` to identify the source platform.
    - Events matching the user's `ExcludedApps` or `ExcludedUrls` lists are silently dropped.
    - Invalid durations (<= 0) are discarded; valid events are enqueued on Redis (HTTP **202 Accepted**).
    - A background `IngestWorker` goroutine pops jobs from the queue and batch-inserts them via `CreateInBatches(100)` into PostgreSQL, then runs auto-categorization and updates `Device.LastSeenAt`.
    - **Fallback**: if Redis is unavailable at startup, the server writes synchronously (HTTP 201) and the ingest worker is a no-op.
3.  **Real-time Data Freshness (Interim Strategy)**:
    - Full WebSocket push is deferred. In the interim, the frontend will **auto-refresh summary data every 30 minutes**.
    - A **manual refresh button** is also present in the UI to allow on-demand data reload.
    - _Future_: Upon successful DB insertion, `IngestEvents` will broadcast an `events_updated` signal to all connected WebSocket clients for that `UserID`, triggering an immediate frontend refetch.
4.  **Aggregation Layer (`SummaryService`, `ReportsService`)**:
    - The web application queries these endpoints on load and after each refresh.
    - Raw `ActivityEvent` rows are parsed and dynamically grouped in-memory based on requested date ranges.

---

## 2. Core Service Algorithms

### 2.1 Timeline Generation (`GetTimeline`)

**Endpoint**: `GET /api/v1/timeline?date=YYYY-MM-DD&cursor=<opaque>&limit=100&app_name=<string>`  
**Purpose**: Build a chronologically sorted, paginated array of a user's day, enriched with app, category, and device context.

- **Date Normalization**: Input `YYYY-MM-DD` is parsed using the user's `Timezone` preference (`time.LoadLocation`) so day boundaries are correct for any timezone.
- **Privacy Filter**: Events with `is_private = true` are excluded from all results.
- **App Name Filter**: Optional `app_name` query parameter narrows results to a specific application.
- **Cursor Pagination**: Returns up to `limit` events (default 100, max 500). The cursor is a `base64(RFC3339Nano timestamp)` of the last returned event's `start_time`. Pass `cursor` in the next request to get the following page. Response includes `next_cursor` when more pages exist, omitted on the last page.
- **Relational Joining**: The GORM query uses `.Preload("Device").Preload("App.Category")` to hydrate foreign keys into full nested structs.
- **Sorting**: Enforces `.Order("start_time asc")` so the UI can map events sequentially left-to-right.

**Status**: ✅ Implemented

---

### 2.2 Daily Summary Aggregation (`GetDailySummary`)

**Endpoint**: `GET /api/v1/summary/daily?date=YYYY-MM-DD`  
**Purpose**: Reduce thousands of daily event strings down to digestible metrics.

- **Timezone-aware day boundaries**: `date` is parsed with `time.LoadLocation(user.Timezone)` so the 24-hour window is always correct for the requesting user's local time.
- **Privacy Filter**: Events with `is_private = true` are excluded.
- Retrieves all non-private events for the requested 24-hour window.
- Iterates with several `O(n)` maps:
    - `appUsageMap[appName] += duration`
    - `deviceUsageMap[deviceId] += duration`
    - `hourUsageMap[hour] += duration` (hour computed in user's local timezone)
- **Idle vs Active**: `IsIdle = true` → `TotalIdleSecs`; otherwise → `TotalActiveSecs`.
- **Productive vs Distraction**: If an event has an associated `App` with a `Category` and `Category.IsProductive != nil`:
    - `*IsProductive == true` → `ProductiveSecs += duration`
    - `*IsProductive == false` → `DistractionSecs += duration`
    - `IsProductive == nil` → uncategorized, not counted in either.
- **Peak Hour Detection**: Scans `hourUsageMap` to find the single 60-minute block with the highest active time.

**Status**: ✅ Implemented

---

### 2.3 Weekly Summary Aggregation (`GetWeeklySummary`)

**Endpoint**: `GET /api/v1/summary/weekly?date=YYYY-MM-DD`  
**Purpose**: Compile an entire week (Mon–Sun) into 7 `DailySummary` objects alongside weekly totals.

1.  **Anchor Calculation**: `offset = int(time.Monday - parsedDate.Weekday())`. If offset is positive (i.e., input is Sunday), shift backward by 6 days.
2.  **Timezone-aware boundaries**: Week start and end are constructed with `time.Date(..., user.Location)` so the Monday 00:00:00 boundary is in the user's local timezone, not UTC.
3.  **Privacy Filter**: Events with `is_private = true` are excluded.
4.  **Range Query**: Fetches all non-private events spanning the calculated 7-day window.
5.  **Bucket Sorting**: `dayIndex = int(event.StartTime.Sub(monday).Hours() / 24)` (clamped to 0–6).
6.  Metrics are accumulated **both** globally into the root `WeeklySummary` and locally into `DailyBreakdown[dayIndex]`, including `ProductiveSecs` and `DistractionSecs`.

**Status**: ✅ Implemented

---

### 2.4 Dynamic Date Reporting (`GetReports`)

**Endpoint**: `GET /api/v1/reports?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`  
**Purpose**: Generate extensive datasets supporting custom UI charts.

- **Timezone-aware date parsing**: `start_date` / `end_date` are parsed with `time.LoadLocation(user.Timezone)` so the boundaries align with the user's local midnight.
- **Privacy Filter**: Events with `is_private = true` are excluded.
- Binds `start_date` / `end_date` from HTTP query parameters into `ReportParams`.
- Constructs a dynamic parameterized GORM query with conditional `.Where("start_time >= ?")` / `.Where("start_time < ?")` clauses.
- Produces:
    - `CategoryUsage map[string]int` — time per category name (falls back to `"Uncategorized"`).
    - `AppUsage map[string]int` — time per app name.
    - `DeviceUsage map[string]int` — time per device name.
    - `DailyActiveTrend map[string]int` — active seconds keyed by `YYYY-MM-DD` for chart plotting.

**Status**: ✅ Implemented

---

### 2.5 Productivity Classification via Category

**Endpoint**: `PATCH /api/v1/categories/:id`  
**Purpose**: Allow users to tag their categories as productive or distracting, which feeds the summary calculations above.

- `Category.IsProductive *bool`:
    - `true` — category counts toward `ProductiveSecs`.
    - `false` — category counts toward `DistractionSecs`.
    - `null` (default) — category is uncategorized; excluded from both scores.
- `UpdateCategoryParams` supports partial updates to `name`, `color`, `icon`, and `is_productive`.
- Only the owning user can patch their categories (`user_id = ?` guard). System-wide categories (`IsSystem = true`) cannot be patched.

**Status**: ✅ Implemented

---

## 3. Database Strategy

- **Batch Inserts**: `IngestEvents` uses `CreateInBatches(events, 100)` to reduce round-trips during client sync flushes. ✅
- **Indexes**: `ActivityEvent` relies on B-Tree indexes on `(user_id, start_time)` and `(user_id, app_name)` to ensure aggregations over thousands of events stay under ~5ms. ✅
- **Cascading Deletes**: Categories, Devices, and UserSettings use `OnDelete:CASCADE` to preserve referential integrity when users or parent rows are removed. ✅
- **Nullable `IsProductive`**: Stored as `boolean DEFAULT NULL` — distinguishes "unclassified" from explicitly productive or distracting. ✅

---

## 4. Remaining Gaps

| Area                                            | Status             | Notes                                                                                                                                                                                                                                                                                        |
| ----------------------------------------------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **WebSocket real-time push**                    | 🔴 Not implemented | Interim: 30-min poll + manual refresh. Deferred to future sprint.                                                                                                                                                                                                                            |
| **Device registration** (`POST /devices`)       | ✅ Implemented     | `POST /devices` accepts `name` + `platform`, generates a unique `device_key`, and returns the new device record.                                                                                                                                                                             |
| **Device deletion** (`DELETE /devices/:id`)     | ✅ Implemented     | Users can remove their own devices; cascades to all associated events.                                                                                                                                                                                                                       |
| **Category creation/deletion**                  | ✅ Implemented     | `POST /categories` creates user-owned categories; `DELETE /categories/:id` nullifies event references then removes the row.                                                                                                                                                                  |
| **`ExcludedApps` / `ExcludedUrls` enforcement** | ✅ Implemented     | `IngestEvents` loads user settings and drops matching events before insert (case-insensitive O(1) map lookup).                                                                                                                                                                               |
| **Data retention purge job**                    | ✅ Implemented     | `PurgeService.PurgeExpiredEvents` deletes events older than `DataRetentionDays` per user. Runnable via `go run ./cmd/purge` or a cron/scheduler.                                                                                                                                             |
| **`LastSeenAt` update on device**               | ✅ Implemented     | `processEvents` calls `UPDATE devices SET last_seen_at = now() WHERE id = ?` after each successful batch insert (both sync and async paths).                                                                                                                                                 |
| **Auto-categorization**                         | 🔴 Removed         | The `autoCategorize` rule-matching approach was replaced. Categories are now assigned directly to `App` records via `PATCH /apps/:id/category` or `PATCH /apps/:id/classify`. The `rules` JSONB field remains on the Category model for reference but is no longer evaluated at ingest time. |
| **Timezone-aware summaries**                    | ✅ Implemented     | `GetDailySummary`, `GetWeeklySummary`, `GetReports`, and `GetTimeline` all parse dates using `time.LoadLocation(user.Timezone)`.                                                                                                                                                             |
| **`is_private` enforcement**                    | ✅ Implemented     | All read queries (timeline, summary, reports) include `AND is_private = false` so private events are invisible to the UI.                                                                                                                                                                    |
| **`DELETE /auth/account`**                      | ✅ Implemented     | Permanently deletes the user row; ON DELETE CASCADE removes all devices, events, settings, and categories.                                                                                                                                                                                   |
| **Rate limiting**                               | ✅ Implemented     | `middleware.RateLimit(cfg.RateLimitRPM)` (per-IP fixed-window) is applied globally. Controlled via `RATE_LIMIT_RPM` env var.                                                                                                                                                                 |
| **Async event ingestion**                       | ✅ Implemented     | Payloads are enqueued to Redis (HTTP 202); `StartIngestWorker` goroutine processes them. Falls back to sync insert if Redis is unavailable.                                                                                                                                                  |
| **Cursor-based timeline pagination**            | ✅ Implemented     | `GET /timeline` accepts `cursor` + `limit` params. Returns `next_cursor` (base64 timestamp) when more pages exist.                                                                                                                                                                           |
