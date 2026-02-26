# Timepad Service Logic & Data Flow

This document details the core algorithms, logic, and data flow mechanisms governing the observer services in the Timepad backend.

---

## 1. System Data Flow

The flow of telemetry data operates in a unidirectional pipeline from client devices up to aggregated reporting views on the frontend.

1.  **Collection Layer**: Background trackers (Windows/Android/Browser Extensions) natively poll the active window or application every few seconds.
2.  **Ingestion API (`POST /api/v1/events`)**:
    - Trackers flush batches of `EventInput` objects to the Go server via `EventsService.IngestEvents`.
    - Each batch is validated against a unique `DeviceKey` to identify the source platform.
    - Invalid durations (<= 0) are discarded; valid events are batch-inserted via `CreateInBatches(100)` into the PostgreSQL `ActivityEvent` table.
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

**Endpoint**: `GET /api/v1/timeline?date=YYYY-MM-DD`  
**Purpose**: Build a chronologically sorted array of a user's day, enriched with category and device context.

- **Date Normalization**: Input `YYYY-MM-DD` is parsed and truncated to `00:00:00` (start of day) and `24:00:00` (end of day).
- **Relational Joining**: The GORM query uses `.Preload("Device").Preload("Category")` to hydrate foreign keys into full nested structs.
- **Sorting**: Enforces `.Order("start_time asc")` so the UI can map events sequentially left-to-right.

**Status**: ✅ Implemented

---

### 2.2 Daily Summary Aggregation (`GetDailySummary`)

**Endpoint**: `GET /api/v1/summary/daily?date=YYYY-MM-DD`  
**Purpose**: Reduce thousands of daily event strings down to digestible metrics.

- Retrieves all events for the requested 24-hour window.
- Iterates with several `O(n)` maps:
    - `appUsageMap[appName] += duration`
    - `deviceUsageMap[deviceId] += duration`
    - `hourUsageMap[hour] += duration`
- **Idle vs Active**: `IsIdle = true` → `TotalIdleSecs`; otherwise → `TotalActiveSecs`.
- **Productive vs Distraction**: If an event's `CategoryID` is set and `Category.IsProductive != nil`:
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
2.  **Range Query**: Fetches all events spanning the calculated 7-day window.
3.  **Bucket Sorting**: `dayIndex = int(event.StartTime.Sub(monday).Hours() / 24)` (clamped to 0–6).
4.  Metrics are accumulated **both** globally into the root `WeeklySummary` and locally into `DailyBreakdown[dayIndex]`, including `ProductiveSecs` and `DistractionSecs`.

**Status**: ✅ Implemented

---

### 2.4 Dynamic Date Reporting (`GetReports`)

**Endpoint**: `GET /api/v1/reports?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`  
**Purpose**: Generate extensive datasets supporting custom UI charts.

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

| Area                                            | Status             | Notes                                                                                                      |
| ----------------------------------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------- |
| **WebSocket real-time push**                    | 🔴 Not implemented | Interim: 30-min poll + manual refresh. Deferred to future sprint.                                          |
| **Device registration** (`POST /devices`)       | 🔴 Not implemented | Only `GET /devices` exists. Trackers have no way to self-register a new `DeviceKey` via the API.           |
| **Device deletion** (`DELETE /devices/:id`)     | 🔴 Not implemented | No way for users to remove a stale/lost device.                                                            |
| **Category creation/deletion**                  | 🔴 Not implemented | Only `GET` + `PATCH` exist. Users cannot create custom categories or delete them via API.                  |
| **`ExcludedApps` / `ExcludedUrls` enforcement** | 🔴 Not implemented | `UserSetting` stores these lists but `IngestEvents` does not filter events against them at ingestion time. |
| **Data retention purge job**                    | 🔴 Not implemented | `DataRetentionDays` is stored in settings but no background job deletes old events automatically.          |
| **`LastSeenAt` update on device**               | 🟡 Partial         | The field exists on `Device` but `IngestEvents` does not update it on each ingest.                         |
