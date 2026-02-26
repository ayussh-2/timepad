# API Test Scenarios — Copy-Paste Ready

> **Base URL:** `http://localhost:8080/api/v1`  
> **Auth Header:** `Authorization: Bearer <ACCESS_TOKEN>`  
> **Content-Type:** `application/json`

Replace these placeholders before testing:

- `<ACCESS_TOKEN>` — your JWT from login
- `<DEVICE_KEY>` — returned from POST /devices
- `<DEVICE_ID>` — UUID of a device from GET /devices
- `<EVENT_ID>` — UUID of an event from GET /events
- `<CATEGORY_ID>` — UUID of a category from GET /categories

---

## 1. Devices

### 1.1 `POST /devices` — Register Windows device

```json
{
    "name": "Work Laptop",
    "platform": "windows"
}
```

**Expected:** 201 — returns device with `device_key` starting with `windows-`

### 1.2 `POST /devices` — Register Android device

```json
{
    "name": "My Phone",
    "platform": "android"
}
```

**Expected:** 201 — `device_key` starts with `android-`

### 1.3 `POST /devices` — Register Browser device

```json
{
    "name": "Chrome Extension",
    "platform": "browser"
}
```

**Expected:** 201 — `device_key` starts with `browser-`

### 1.4 `POST /devices` — ❌ Invalid platform

```json
{
    "name": "Linux Box",
    "platform": "linux"
}
```

**Expected:** 400 — validation error, platform must be `android`, `windows`, or `browser`

### 1.5 `POST /devices` — ❌ Missing name

```json
{
    "platform": "windows"
}
```

**Expected:** 400 — validation error

### 1.6 `POST /devices` — ❌ Empty body

```json
{}
```

**Expected:** 400 — validation error

### 1.7 `GET /devices`

No body. Just send the GET request with auth header.

**Expected:** 200 — array of all user's devices

### 1.8 `DELETE /devices/<DEVICE_ID>`

No body. Replace `<DEVICE_ID>` with the UUID from GET /devices.

**Expected:** 200 — "Device deleted successfully"

### 1.9 `DELETE /devices/00000000-0000-0000-0000-000000000000` — ❌ Non-existent

**Expected:** 500 — "device not found or unauthorized"

---

## 2. Categories

### 2.1 `POST /categories` — Full creation

```json
{
    "name": "Gaming",
    "color": "#FF4444",
    "icon": "🎮",
    "is_productive": false
}
```

**Expected:** 201 — returns created category

### 2.2 `POST /categories` — Productive category

```json
{
    "name": "Deep Work",
    "color": "#22C55E",
    "icon": "💻",
    "is_productive": true
}
```

**Expected:** 201

### 2.3 `POST /categories` — Minimal (name only)

```json
{
    "name": "Uncategorized Stuff"
}
```

**Expected:** 201 — color defaults to `#6B7280`, `is_productive` is null

### 2.4 `POST /categories` — ❌ Missing name

```json
{
    "color": "#FF0000"
}
```

**Expected:** 400 — validation error

### 2.5 `GET /categories`

No body.

**Expected:** 200 — array of user categories + system categories

### 2.6 `PATCH /categories/<CATEGORY_ID>` — Update name and color

```json
{
    "name": "Serious Gaming",
    "color": "#FF8800"
}
```

**Expected:** 200

### 2.7 `PATCH /categories/<CATEGORY_ID>` — Mark as productive

```json
{
    "is_productive": true
}
```

**Expected:** 200

### 2.8 `PATCH /categories/<CATEGORY_ID>` — Update icon only

```json
{
    "icon": "🔥"
}
```

**Expected:** 200

### 2.9 `DELETE /categories/<CATEGORY_ID>`

No body.

**Expected:** 200 — category deleted, events referencing it get `category_id = null`

### 2.10 `DELETE /categories/00000000-0000-0000-0000-000000000000` — ❌ Non-existent

**Expected:** 500 — "category not found or unauthorized"

---

## 3. Settings

### 3.1 `GET /settings`

No body.

**Expected:** 200

```json
{
    "excluded_apps": [],
    "excluded_urls": [],
    "idle_threshold": 300,
    "tracking_enabled": true,
    "data_retention_days": 365
}
```

### 3.2 `PUT /settings` — Set excluded apps

```json
{
    "excluded_apps": ["Slack", "Discord", "Spotify"]
}
```

**Expected:** 200

### 3.3 `PUT /settings` — Set excluded URLs

```json
{
    "excluded_urls": ["reddit.com", "twitter.com", "youtube.com"]
}
```

**Expected:** 200

### 3.4 `PUT /settings` — Update idle threshold

```json
{
    "idle_threshold": 600
}
```

**Expected:** 200

### 3.5 `PUT /settings` — Disable tracking

```json
{
    "tracking_enabled": false
}
```

**Expected:** 200

### 3.6 `PUT /settings` — Set data retention

```json
{
    "data_retention_days": 90
}
```

**Expected:** 200

### 3.7 `PUT /settings` — Update multiple fields

```json
{
    "excluded_apps": ["Slack"],
    "idle_threshold": 120,
    "data_retention_days": 180
}
```

**Expected:** 200

### 3.8 `PUT /settings` — Clear excluded lists

```json
{
    "excluded_apps": [],
    "excluded_urls": []
}
```

**Expected:** 200 — lists are now empty

---

## 4. Events — Ingestion

### 4.1 `POST /events` — Single event

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Visual Studio Code",
            "window_title": "main.go — timepad",
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:30:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 1 }`

### 4.2 `POST /events` — Multiple events

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Visual Studio Code",
            "window_title": "router.go",
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:45:00Z"
        },
        {
            "app_name": "Google Chrome",
            "window_title": "Stack Overflow",
            "url": "stackoverflow.com",
            "start_time": "2026-02-25T09:45:00Z",
            "end_time": "2026-02-25T10:00:00Z"
        },
        {
            "app_name": "Slack",
            "window_title": "#general",
            "start_time": "2026-02-25T10:00:00Z",
            "end_time": "2026-02-25T10:15:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 3 }`

### 4.3 `POST /events` — With idle event

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "System Idle",
            "start_time": "2026-02-25T12:00:00Z",
            "end_time": "2026-02-25T12:30:00Z",
            "is_idle": true
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 1 }` with `is_idle = true`

### 4.4 `POST /events` — With URL

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Firefox",
            "window_title": "GitHub - timepad",
            "url": "github.com",
            "start_time": "2026-02-25T14:00:00Z",
            "end_time": "2026-02-25T14:30:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 1 }`

### 4.5 `POST /events` — ❌ Zero-duration event (skipped silently)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Notepad",
            "start_time": "2026-02-25T15:00:00Z",
            "end_time": "2026-02-25T15:00:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 0 }`

### 4.6 `POST /events` — ❌ Negative duration (skipped silently)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Notepad",
            "start_time": "2026-02-25T16:00:00Z",
            "end_time": "2026-02-25T15:00:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 0 }`

### 4.7 `POST /events` — ❌ Invalid device key

```json
{
    "device_key": "fake-nonexistent-key",
    "events": [
        {
            "app_name": "VS Code",
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:30:00Z"
        }
    ]
}
```

**Expected:** 500 — "unknown device"

### 4.8 `POST /events` — ❌ Missing device_key

```json
{
    "events": [
        {
            "app_name": "VS Code",
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:30:00Z"
        }
    ]
}
```

**Expected:** 400 — validation error

### 4.9 `POST /events` — ❌ Empty events array

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": []
}
```

**Expected:** 400 — validation error (min=1)

### 4.10 `POST /events` — ❌ Missing app_name

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:30:00Z"
        }
    ]
}
```

**Expected:** 400 — validation error

---

## 5. Events — ExcludedApps/Urls Enforcement

> **Setup:** First run `PUT /settings` with excluded apps/urls, then test ingestion

### 5.1 Setup — Exclude "Slack" and "reddit.com"

`PUT /settings`:

```json
{
    "excluded_apps": ["Slack"],
    "excluded_urls": ["reddit.com"]
}
```

### 5.2 `POST /events` — Excluded app (should be filtered)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Slack",
            "window_title": "#random",
            "start_time": "2026-02-25T11:00:00Z",
            "end_time": "2026-02-25T11:30:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 0 }` (filtered out)

### 5.3 `POST /events` — Excluded app case-insensitive

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "SLACK",
            "start_time": "2026-02-25T11:00:00Z",
            "end_time": "2026-02-25T11:30:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 0 }` (case-insensitive match)

### 5.4 `POST /events` — Excluded URL (should be filtered)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Chrome",
            "url": "reddit.com",
            "start_time": "2026-02-25T11:30:00Z",
            "end_time": "2026-02-25T12:00:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 0 }`

### 5.5 `POST /events` — Non-excluded app (should pass)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "VS Code",
            "start_time": "2026-02-25T11:00:00Z",
            "end_time": "2026-02-25T11:30:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 1 }`

### 5.6 `POST /events` — Mixed batch (1 excluded + 1 valid)

```json
{
    "device_key": "<DEVICE_KEY>",
    "events": [
        {
            "app_name": "Slack",
            "start_time": "2026-02-25T13:00:00Z",
            "end_time": "2026-02-25T13:15:00Z"
        },
        {
            "app_name": "VS Code",
            "start_time": "2026-02-25T13:15:00Z",
            "end_time": "2026-02-25T13:45:00Z"
        }
    ]
}
```

**Expected:** 201 — `{ "inserted": 1 }` (Slack filtered, VS Code saved)

### 5.7 Cleanup — Clear exclusions

`PUT /settings`:

```json
{
    "excluded_apps": [],
    "excluded_urls": []
}
```

---

## 6. Events — Read, Edit, Delete

### 6.1 `GET /events`

No body. Optional query params: `?limit=10&offset=0`

**Expected:** 200 — array of events ordered by `start_time desc`

### 6.2 `GET /events?limit=3`

**Expected:** 200 — at most 3 events

### 6.3 `GET /events?limit=3&offset=3`

**Expected:** 200 — next 3 events (page 2)

### 6.4 `GET /timeline?date=2026-02-25`

No body.

**Expected:** 200 — events for that date with `category` and `device` objects preloaded, ordered `start_time asc`

### 6.5 `GET /timeline` — ❌ Missing date

**Expected:** 400 — "Date parameter is required"

### 6.6 `GET /timeline?date=25-02-2026` — ❌ Wrong format

**Expected:** 500 — "invalid date format"

### 6.7 `GET /timeline?date=1999-01-01` — Empty day

**Expected:** 200 — empty array

### 6.8 `PATCH /events/<EVENT_ID>` — Assign category

```json
{
    "category_id": "<CATEGORY_ID>"
}
```

**Expected:** 200

### 6.9 `PATCH /events/<EVENT_ID>` — Remove category

```json
{
    "category_id": ""
}
```

**Expected:** 200 — `category_id` set to null

### 6.10 `PATCH /events/<EVENT_ID>` — Mark private

```json
{
    "is_private": true
}
```

**Expected:** 200

### 6.11 `PATCH /events/<EVENT_ID>` — Update both

```json
{
    "category_id": "<CATEGORY_ID>",
    "is_private": true
}
```

**Expected:** 200

### 6.12 `PATCH /events/<EVENT_ID>` — ❌ Invalid category UUID

```json
{
    "category_id": "not-a-valid-uuid"
}
```

**Expected:** 500 — "invalid category ID"

### 6.13 `PATCH /events/00000000-0000-0000-0000-000000000000` — ❌ Non-existent

```json
{
    "is_private": true
}
```

**Expected:** 500 — "event not found or unauthorized"

### 6.14 `DELETE /events/<EVENT_ID>`

No body.

**Expected:** 200 — "Event deleted successfully"

### 6.15 `DELETE /events/00000000-0000-0000-0000-000000000000` — ❌ Non-existent

**Expected:** 500 — "event not found or unauthorized"

---

## 7. Summary

### 7.1 `GET /summary/daily?date=2026-02-25`

No body.

**Expected:** 200

```json
{
    "date": "2026-02-25",
    "total_active_secs": 5400,
    "total_idle_secs": 1800,
    "productive_secs": 2700,
    "distraction_secs": 900,
    "top_apps": [
        { "app_name": "VS Code", "total_secs": 2700, "category": {...} }
    ],
    "peak_hour": 9,
    "device_breakdown": [
        { "device_name": "Work Laptop", "platform": "windows", "total_secs": 5400 }
    ]
}
```

### 7.2 `GET /summary/daily?date=1999-01-01` — Empty day

**Expected:** 200 — all zeroes, empty arrays

### 7.3 `GET /summary/daily` — ❌ Missing date

**Expected:** 400 — "Date parameter is required"

### 7.4 `GET /summary/daily?date=invalid` — ❌ Bad format

**Expected:** 500 — "invalid date format"

### 7.5 `GET /summary/weekly?date=2026-02-25`

No body.

**Expected:** 200

```json
{
    "start_date": "2026-02-23",
    "end_date": "2026-03-01",
    "total_active_secs": 28800,
    "total_idle_secs": 3600,
    "productive_secs": 21600,
    "distraction_secs": 3600,
    "top_apps": [...],
    "daily_breakdown": [
        { "date": "2026-02-23", "total_active_secs": 0, ... },
        { "date": "2026-02-24", "total_active_secs": 0, ... },
        { "date": "2026-02-25", "total_active_secs": 5400, ... },
        ...
    ]
}
```

### 7.6 `GET /summary/weekly?date=1999-01-01` — Empty week

**Expected:** 200 — all zeroes, 7 empty daily breakdowns

### 7.7 `GET /summary/weekly?date=invalid` — ❌ Bad format

**Expected:** 500 — "invalid date format"

---

## 8. Reports

### 8.1 `GET /reports?start_date=2026-02-01&end_date=2026-02-28`

No body.

**Expected:** 200

```json
{
    "total_active_secs": 86400,
    "total_idle_secs": 7200,
    "category_usage": {
        "Coding": 43200,
        "Browsing": 21600,
        "Gaming": 7200
    },
    "app_usage": {
        "VS Code": 43200,
        "Chrome": 21600
    },
    "device_usage": {
        "Work Laptop": 64800
    },
    "daily_active_trend": {
        "2026-02-25": 5400,
        "2026-02-26": 3600
    }
}
```

### 8.2 `GET /reports?start_date=2026-02-25` — Only start date

**Expected:** 200 — data from Feb 25 to now

### 8.3 `GET /reports?end_date=2026-02-28` — Only end date

**Expected:** 200 — all data up to Feb 28

### 8.4 `GET /reports` — No date filters

**Expected:** 200 — all events for the user

### 8.5 `GET /reports?start_date=3000-01-01&end_date=3000-12-31` — ❌ Future range

**Expected:** 200 — all zeroes, empty maps

---

## 9. Integration Tests (Suggested Order)

### 9.1 Full Device Lifecycle

1. `POST /devices` → save `device_key` and `device_id`
2. `POST /events` with that `device_key` → verify `inserted: 1`
3. `GET /devices` → verify `last_seen_at` is updated
4. `DELETE /devices/<device_id>` → verify 200
5. `GET /devices` → device is gone

### 9.2 Full Category Lifecycle

1. `POST /categories` with `"name": "Test Cat"` → save `category_id`
2. Ingest an event → save `event_id`
3. `PATCH /events/<event_id>` with `"category_id": "<category_id>"`
4. `GET /timeline?date=2026-02-25` → event has category attached
5. `DELETE /categories/<category_id>`
6. `GET /timeline?date=2026-02-25` → event now has `category_id = null`

### 9.3 Productivity Tracking

1. Create category `"Deep Work"` with `is_productive: true`
2. Create category `"Social Media"` with `is_productive: false`
3. Ingest 2 events, assign one to each category
4. `GET /summary/daily` → verify `productive_secs > 0` and `distraction_secs > 0`

### 9.4 Excluded Apps End-to-End

1. `PUT /settings` with `excluded_apps: ["notepad.exe"]`
2. `POST /events` with `app_name: "notepad.exe"` → `inserted: 0`
3. `POST /events` with `app_name: "code.exe"` → `inserted: 1`
4. `PUT /settings` with `excluded_apps: []` (cleanup)
