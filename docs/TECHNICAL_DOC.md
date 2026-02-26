# Cross-Device Time Tracker — Technical Documentation

**Version:** 1.1.0  
**Last Updated:** February 2026  
**Status:** Draft

---

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Technology Stack](#2-technology-stack)
3. [Repository Structure](#3-repository-structure)
4. [Central Go Server](#4-central-go-server)
5. [Web App (React)](#5-web-app-react)
6. [Client Collectors](#6-client-collectors)
7. [Database Design](#7-database-design)
8. [API Reference](#8-api-reference)
9. [Authentication & Security](#9-authentication--security)
10. [Data Flow](#10-data-flow)
11. [Cross-Device Sync](#11-cross-device-sync)
12. [Deployment](#12-deployment)
13. [Environment Variables](#13-environment-variables)
14. [Development Setup](#14-development-setup)

---

## 1. System Architecture Overview

The application is divided into two distinct layers:

**Collector Layer** — Thin, platform-native components that run in the background, detect user activity, and ship raw events to the central server. Each platform has its own native implementation.

**Presentation Layer** — A single React web application that handles all UI: dashboards, timelines, reports, and settings. It is served as a standalone web app and also embedded via WebView inside the Android and Windows native wrappers.

The architecture is a classic hub-and-spoke model. Each collector independently sends data directly to the central Go server using a REST API over HTTPS. The server is the single source of truth — all storage, aggregation, and categorization happen there. The React web app reads from the same API and presents unified views across all devices.

---

## 2. Technology Stack

### Central Server

| Component      | Choice        | Reason                                    |
| -------------- | ------------- | ----------------------------------------- |
| Language       | Go 1.24+      | Performant, low memory, great concurrency |
| HTTP Framework | Gin           | Fast routing, middleware support          |
| ORM            | GORM          | Clean Go ORM, good migration support      |
| Database       | PostgreSQL 16 | Reliable, JSONB support, strong typing    |
| Auth           | JWT (RS256)   | Stateless, works across devices           |

### Web App (Dashboard UI)

| Component            | Choice          | Reason                                       |
| -------------------- | --------------- | -------------------------------------------- |
| Framework            | React 19        | Component model, hooks, wide ecosystem       |
| Build Tool           | Vite            | Fast HMR, excellent React support            |
| State Management     | Zustand         | Lightweight, hook-based, minimal boilerplate |
| Routing              | React Router v6 | Industry standard, file-based routing        |
| HTTP Client          | Axios           | Interceptors, easy auth header injection     |
| Charts               | Recharts        | React-native chart library, composable       |
| UI Component Library | shadcn/ui       | Accessible, unstyled, customizable           |
| CSS                  | Tailwind CSS    | Utility-first, easy to customize             |
| Date Handling        | Day.js          | Lightweight Moment.js alternative            |
| TypeScript           | Yes             | Type safety across the entire app            |

### Android Collector

| Component          | Choice                                   |
| ------------------ | ---------------------------------------- |
| Language           | Kotlin                                   |
| Min SDK            | 26 (Android 8)                           |
| Background Service | Foreground Service                       |
| Activity Detection | UsageStatsManager + AccessibilityService |
| WebView            | Android WebView (Chromium)               |
| HTTP               | Retrofit 2 + OkHttp                      |

### Windows Collector

| Component       | Choice                                   |
| --------------- | ---------------------------------------- |
| Language        | Go                                       |
| Window Tracking | Win32 API via `golang.org/x/sys/windows` |
| UI              | Systray (`getlantern/systray`)           |
| WebView         | WebView2 (`jchv/go-webview2`)            |
| HTTP            | Standard `net/http`                      |

### Browser Extension

| Component  | Choice                         |
| ---------- | ------------------------------ |
| Target     | Chrome, Edge, Firefox          |
| Manifest   | V3 (Chrome/Edge), V2 (Firefox) |
| Background | Service Worker (MV3)           |
| Language   | TypeScript                     |
| Build      | Vite + CRXJS                   |

---

## 3. Repository Structure

The monorepo is organized into two top-level areas: `server/` for the Go backend and `ui/` (planned) for all frontend and client code.

**Server (`server/`)** — ✅ Exists. Contains the main Go application entry point under `cmd/`, with all business logic organized inside `internal/` using the following sub-packages: `controllers/` for HTTP handlers, `services/` for business logic, `models/` for GORM models, `routes/` for route registration, `middleware/` for auth and CORS, and `utils/` for shared helpers. Migrations live in the `cmd/migrate` directory. A `cmd/seed` directory also exists.

**UI Core (`ui/core/`)** — 🔴 Not yet created. Will be the React 19 web application — the single, shared dashboard UI. Planned as a standard Vite + React project with key directories: `components/`, `pages/`, `store/` (Zustand), `api/` (Axios wrappers), `types/`, and `hooks/`.

**UI Wrappers (`ui/wrappers/`)** — 🔴 Not yet created. Will contain the platform-specific native shells that embed the `ui/core` React app inside a WebView and provide a background activity collector. Planned sub-directories: `android/` (Kotlin), `windows/` (Go tray app), `extension/` (browser extension).

**Infrastructure (`docker/`)** — 🔴 Not yet created. Will contain Docker Compose files for local dev and production.

---

## 4. Central Go Server

### 4.1 Server Entry Point

The main entry point in `cmd/server/main.go` loads configuration, establishes the database connection, constructs the Gin router via `routes.SetupRouter`, and starts listening. The server uses `air` for hot-reload in development.

### 4.2 Router Setup

All routes are registered in `routes/routes.go`. Routes are split into two groups:

- **Public routes** (`/api/v1/auth/*`, `/api/v1/health`): No authentication required.
- **Protected routes** (all others): Guarded by the JWT `middleware.Auth` middleware which validates the bearer token on every request and injects the `userID` string into the Gin context.

Currently registered protected endpoints:

| Method | Path              | Handler            |
| ------ | ----------------- | ------------------ |
| POST   | `/events`         | `IngestEvents`     |
| GET    | `/events`         | `GetEvents`        |
| PATCH  | `/events/:id`     | `EditEvent`        |
| DELETE | `/events/:id`     | `DeleteEvent`      |
| GET    | `/timeline`       | `GetTimeline`      |
| GET    | `/summary/daily`  | `GetDailySummary`  |
| GET    | `/summary/weekly` | `GetWeeklySummary` |
| GET    | `/reports`        | `GetReports`       |
| GET    | `/categories`     | `GetCategories`    |
| PATCH  | `/categories/:id` | `UpdateCategory`   |
| GET    | `/devices`        | `GetDevices`       |
| GET    | `/settings`       | `GetSettings`      |
| PUT    | `/settings`       | `UpdateSettings`   |

### 4.3 Models

All models live in `internal/models/models.go` and are managed by GORM AutoMigrate on startup. See **Section 7** for the full schema.

### 4.4 Event Ingestion

`EventsService.IngestEvents` validates the `DeviceKey` against the calling user's devices, filters out events with zero or negative duration, and bulk-inserts valid events using `db.CreateInBatches(events, 100)`.

### 4.5 Background Jobs

There are currently no scheduled background jobs running. The following are identified as future work:

- **Auto-categorization job** — scan events with `category_id IS NULL` and apply user-defined category rules.
- **Data retention purge** — delete events older than the user's `data_retention_days` setting.

---

## 5. Web App (React) — 🔴 Planned

> **Note:** The React web app has not been built yet. This section describes the planned architecture.

### 5.1 Project Setup

The frontend will be a Vite-powered React 19 application written in TypeScript, living in `ui/core/`. It will proxy `/api` requests to the Go server at `localhost:8080` during development.

### 5.2 API Client

A shared Axios instance is created with the API base URL pre-configured. Request interceptors attach the JWT bearer token from the auth store on every outgoing request. Response interceptors handle 401 errors by attempting a token refresh via the `/auth/refresh` endpoint; on failure, the user is logged out and redirected to the login page.

### 5.3 State Management (Zustand)

Three main stores:

- **Auth store** — holds access and refresh tokens, the current user object, and exposes `login`, `logout`, and `refreshToken` actions. Tokens are persisted to `localStorage`.
- **Activity store** — holds the current `timeline` array, `dailySummary`, `weeklySummary`, `selectedDate`, and `isLoading` state. Exposes fetch actions that call the Axios API wrappers.
- **Settings store** — holds user settings and exposes a `updateSettings` action.

### 5.4 Data Freshness Strategy

WebSocket support is deferred. In the interim, the web app uses two mechanisms to keep data fresh:

1. **Auto-refresh every 30 minutes** — A `useEffect` hook in the main layout sets up a `setInterval` that re-calls the active summary and timeline fetch actions.
2. **Manual refresh button** — A refresh icon button in the dashboard header dispatches the same fetch actions on click.

### 5.5 Routing

Routes are defined using React Router v6. All routes except `/login` and `/register` are wrapped in a `PrivateRoute` component that checks the auth store and redirects unauthenticated users to `/login`.

| Path         | Page Component          |
| ------------ | ----------------------- |
| `/login`     | `LoginPage`             |
| `/register`  | `RegisterPage`          |
| `/`          | Redirect → `/dashboard` |
| `/dashboard` | `DashboardPage`         |
| `/timeline`  | `TimelinePage`          |
| `/reports`   | `ReportsPage`           |
| `/settings`  | `SettingsPage`          |

### 5.6 Key TypeScript Types

The `src/types/index.ts` file defines all shared interfaces: `TimelineEntry`, `DailySummary`, `WeeklySummary`, `ReportData`, `Category`, `Device`, `AppUsage`, `DeviceUsage`, and `UserSettings`. These mirror the JSON shapes returned by the Go API.

**Notable fields on `DailySummary`:**

| Field               | Type            | Description                                      |
| ------------------- | --------------- | ------------------------------------------------ |
| `total_active_secs` | `number`        | Total non-idle time                              |
| `total_idle_secs`   | `number`        | Total idle time                                  |
| `productive_secs`   | `number`        | Time in categories marked `is_productive: true`  |
| `distraction_secs`  | `number`        | Time in categories marked `is_productive: false` |
| `peak_hour`         | `number`        | Hour of day (0-23) with highest activity         |
| `top_apps`          | `AppUsage[]`    | Per-app breakdown with category                  |
| `device_breakdown`  | `DeviceUsage[]` | Per-device breakdown                             |

---

## 6. Client Collectors — 🔴 Planned

> **Note:** None of the client collectors have been built yet. This section describes the planned architecture for each platform.

### 6.1 Android Collector (Kotlin)

The Android collector runs as a persistent Foreground Service to survive background restrictions. It uses `UsageStatsManager` to query app usage in 30-second polling intervals. On each poll, it maps usage stats into `EventInput` objects and POSTs them to the server.

**Required Android Manifest Permissions:**

- `android.permission.PACKAGE_USAGE_STATS` — to read app usage stats
- `android.permission.FOREGROUND_SERVICE` — to run persistently
- `android.permission.POST_NOTIFICATIONS` — for the persistent notification
- `android.permission.INTERNET` — for API calls

**WebView for Dashboard:** `MainActivity` loads the React web app URL in an Android WebView. A native bridge (`AndroidBridge`) exposes two methods via `@JavascriptInterface`: `getDeviceKey()` and `getPlatform()`, so the web app can read the device identity without needing separate credentials.

---

### 6.2 Windows Collector (Go)

The Windows collector is a Go application that runs as a system tray icon. It uses the Win32 API via `golang.org/x/sys/windows` to poll the foreground window every 30 seconds. It detects the active process name and window title, tracks session starts and ends, and buffers events locally. When the buffer reaches 10 events or a flush interval elapses, it POSTs the batch to the server.

Idle detection is done by reading the system's last-input timestamp via `GetLastInputInfo` and comparing against the user's configured `idle_threshold` seconds.

A WebView2 window can be launched from the tray to display the React web app dashboard inline.

---

### 6.3 Browser Extension

The extension targets Chrome, Edge (MV3), and Firefox (MV2). A background Service Worker listens to `chrome.tabs.onActivated` and `chrome.tabs.onUpdated` events to track tab switches and navigations. When the active tab changes, the previous tab's session duration is calculated and pushed into a local buffer (visits under 5 seconds are ignored). The buffer is flushed to the server every 60 seconds via a `fetch` call.

The extension stores the JWT token and `device_key` in `chrome.storage.local`. The popup UI provides a simple login form and displays today's tracked time.

---

## 7. Database Design

### 7.1 Schema

All tables are managed by GORM AutoMigrate. The following tables exist:

**`users`** — Stores user accounts. Fields: `id` (UUID PK), `email` (unique), `password_hash`, `display_name`, `timezone` (default `UTC`), `created_at`, `updated_at`.

**`devices`** — Stores registered devices per user. Fields: `id` (UUID PK), `user_id` (FK → users, CASCADE), `name`, `platform` (`android` | `windows` | `browser`, enforced by CHECK constraint), `device_key` (unique), `last_seen_at`, `created_at`.

**`categories`** — Stores categorization labels, either system-wide (`is_system = true`, `user_id = NULL`) or user-specific. Fields: `id` (UUID PK), `user_id` (nullable FK → users, CASCADE), `name`, `color` (default `#6B7280`), `icon`, `is_system`, `is_productive` (nullable boolean — `true` = productive, `false` = distraction, `NULL` = uncategorized), `rules` (JSONB array of matching rules).

**`activity_events`** — Primary data table. Fields: `id` (UUID PK), `user_id` (FK → users, CASCADE), `device_id` (FK → devices, CASCADE), `app_name`, `window_title`, `url`, `category_id` (nullable FK → categories), `start_time`, `end_time`, `duration_secs`, `is_idle` (default `false`), `is_private` (default `false`), `raw_meta` (JSONB), `created_at`.

**`user_settings`** — One-to-one with users. Fields: `user_id` (UUID PK, FK → users, CASCADE), `excluded_apps` (text[]), `excluded_urls` (text[]), `idle_threshold` (default 300s), `tracking_enabled` (default `true`), `data_retention_days` (default 365), `updated_at`.

### 7.2 Indexes

| Index                   | Columns                      | Purpose                                        |
| ----------------------- | ---------------------------- | ---------------------------------------------- |
| `idx_events_user_start` | `(user_id, start_time DESC)` | Primary query pattern for timeline and summary |
| `idx_events_device`     | `(device_id)`                | Device-filtered queries                        |
| `idx_events_category`   | `(category_id)`              | Category breakdown queries                     |
| `idx_events_app_name`   | `(user_id, app_name)`        | App usage aggregation                          |

### 7.3 Category Rules Schema

The `rules` JSONB column on categories stores an array of matching rule objects. Each rule has a `type` (`app_name`, `url_domain`, or `window_title`), an `op` (`contains`, `equals`, `startsWith`), and a `value` string. The auto-categorization job (future) evaluates these rules against incoming events.

---

## 8. API Reference

### Base URL

Development: `http://localhost:8080/api/v1`

### Authentication

All protected endpoints require an `Authorization: Bearer <access_token>` header.

---

### Auth Endpoints

#### `POST /auth/register`

**Request body:** `email`, `password`, `display_name`  
**Response 201:** Returns the new user object with `access_token` and `refresh_token`.

#### `POST /auth/login`

**Request body:** `email`, `password`  
**Response 200:** Returns `access_token`, `refresh_token`, `expires_in`, and the user object.

#### `POST /auth/refresh`

**Request body:** `refresh_token`  
**Response 200:** Returns a new `access_token` and `refresh_token`.

---

### Events Endpoints

#### `POST /events` — Ingest activity events (called by collectors)

**Request body:** `device_key` (string), `events` (array of event objects with `app_name`, `window_title`, `url`, `start_time`, `end_time`, `is_idle`). At least one event is required.  
**Response 201:** Returns `{ "inserted": N }` where N is the count of valid events saved.

#### `GET /events?limit=50&offset=0`

Returns a paginated list of raw activity events for the authenticated user, ordered by `start_time` descending.

#### `GET /timeline?date=YYYY-MM-DD`

Returns all events for the specified date, enriched with full `category` and `device` structs, ordered by `start_time` ascending.

#### `PATCH /events/:id`

**Request body (all optional):** `category_id` (string UUID or empty string to unset), `is_private` (boolean).  
**Response 200:** Success confirmation.

#### `DELETE /events/:id`

**Response 200:** Success confirmation. Only the owning user can delete their events.

---

### Summary Endpoints

#### `GET /summary/daily?date=YYYY-MM-DD`

Returns a `DailySummary` object for the specified date. See Section 5.6 for the full field list.

#### `GET /summary/weekly?date=YYYY-MM-DD`

Returns a `WeeklySummary` containing global weekly totals plus a `daily_breakdown` array of 7 `DailySummary` objects (Mon–Sun of the week containing the provided date).

---

### Reports Endpoint

#### `GET /reports?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`

Both date parameters are optional. Returns a `ReportData` object with: `total_active_secs`, `total_idle_secs`, `category_usage` (map of category name → seconds), `app_usage` (map of app name → seconds), `device_usage` (map of device name → seconds), `daily_active_trend` (map of date string → active seconds).

---

### Categories Endpoints

#### `GET /categories`

Returns all categories visible to the user — their own user-specific categories plus all system categories (`is_system = true`).

#### `PATCH /categories/:id`

**Request body (all optional):** `name`, `color`, `icon`, `is_productive` (boolean). Only the owning user's categories can be patched; system categories are protected.  
**Response 200:** Success confirmation.

---

### Devices Endpoint

#### `GET /devices`

Returns all devices registered to the authenticated user.

> **Note:** `POST /devices` (device registration) and `DELETE /devices/:id` are not yet implemented.

---

### Settings Endpoints

#### `GET /settings`

Returns the user's current settings: `excluded_apps`, `excluded_urls`, `idle_threshold`, `tracking_enabled`, `data_retention_days`.

#### `PUT /settings`

**Request body (all optional):** Any subset of the settings fields. Only provided fields are updated.

---

### WebSocket (Deferred)

`GET /ws?token=<access_token>` — WebSocket upgrade endpoint. **Not yet implemented.** Planned to push `events_updated`, `sync_complete`, and `category_updated` messages. In the interim, the React app uses a 30-minute auto-refresh and a manual refresh button.

---

## 9. Authentication & Security

### JWT Strategy

- **Access Token:** RS256, 1-hour expiry. Signed with a private key kept only on the server.
- **Refresh Token:** 30-day expiry. Can be invalidated server-side by rotating the token.
- **Device Key:** A UUID assigned at device registration. Included in every event POST request alongside the JWT to bind events to a specific device.

### WebView Security

The Android and Windows WebView embeds the React web app from a trusted origin. Native bridges only expose the minimum needed: the `device_key` and `platform` identifier. No raw API credentials are passed through the bridge.

### Data Privacy

- Events marked `is_private: true` are stored with the flag but **not** currently encrypted or filtered from reports. Privacy enforcement is planned.
- The `excluded_apps` and `excluded_urls` settings are stored in `user_settings` but **not** enforced at ingestion time — all events are accepted regardless.
- `DELETE /account` for full data deletion is not yet implemented.

### Rate Limiting

Rate limiting is planned but not yet implemented. The `RATE_LIMIT_RPM` config value is loaded from the environment but not applied to any middleware. The design calls for a token bucket approach with 60 requests/minute per device on `/events`, and 10 requests/minute per IP on auth endpoints.

---

## 10. Data Flow

### Event Ingestion Flow

1. Collector detects a change in active application or tab.
2. Collector buffers events locally for 30–60 seconds.
3. Collector POSTs the batch to `POST /api/v1/events` with `device_key` + JWT.
4. Server validates JWT and resolves `device_key` to a known device.
5. Server filters invalid events (duration ≤ 0).
6. Server batch-inserts valid events into `activity_events`.
7. _(Future)_ Server broadcasts `events_updated` signal via WebSocket.
8. React app refetches timeline/summary on next refresh cycle (auto or manual).

### Timeline Query Flow

1. User opens the Timeline page.
2. React app calls `GET /timeline?date=YYYY-MM-DD`.
3. Server queries `activity_events` for the user + date range.
4. GORM preloads `Category` and `Device` associations.
5. Events are ordered by `start_time ASC`.
6. Unified timeline returned — merges all devices into one chronological view.
7. React renders events in a horizontal bar timeline, color-coded by category.

### Summary Aggregation Flow

1. React app calls `GET /summary/daily?date=YYYY-MM-DD`.
2. Server fetches all events for the day.
3. Server reduces in a single O(n) pass: accumulates `appUsageMap`, `deviceUsageMap`, `hourUsageMap`, total active/idle seconds, and productive/distraction seconds (based on `Category.IsProductive`).
4. Peak hour is determined by scanning `hourUsageMap` for the max value.
5. Aggregated `DailySummary` is returned and rendered on the dashboard.

---

## 11. Cross-Device Sync

All three collectors report to the same central server independently. There is no device-to-device communication. The server merges events from all devices into one timeline, tagged with `device_id` so the UI can filter or color-code by device.

**Sync timing:**

- Android: polls every 30 seconds
- Windows: polls every 30 seconds
- Browser Extension: flushes buffer every 60 seconds

On first install, each client must be registered and assigned a `device_key`. Currently, device registration (`POST /devices`) is not exposed via API and must be done manually with a direct DB insert. This is a known gap — see the remaining gaps section in `SERVICE_LOGIC.md`.

**Conflict handling:** Overlapping time ranges across different devices are allowed and stored as-is — the timeline view will render them as parallel lanes. There is **no** deduplication logic currently implemented at ingestion time.

---

## 12. Deployment — 🔴 Planned

> **Note:** No Docker or deployment infrastructure has been created yet. This section describes the planned setup.

### Docker Compose (Development)

A `docker/docker-compose.yml` will start PostgreSQL 16 and the Go server. The React app will run separately via `npm run dev` during development.

**Planned services:**

- `postgres` — PostgreSQL 16 Alpine, exposed on port 5432, with a named volume for data persistence.
- `server` — Built from a Go Dockerfile. Reads `DATABASE_URL`, JWT key files, and other config from environment variables.

### Server Dockerfile

Planned as a multi-stage build: `golang:1.24-alpine` compiles the binary, then the final image is `alpine:3.19` with just the binary and CA certificates.

### Web App Deployment

The React app will be built with `npm run build` (producing a `dist/` folder) and served as static files behind an Nginx reverse proxy with `try_files $uri /index.html` for client-side routing.

---

## 13. Environment Variables

### Server (✅ Exists in `server/.env.example`)

| Variable               | Description                   | Example                             |
| ---------------------- | ----------------------------- | ----------------------------------- |
| `SERVER_ADDR`          | Bind address                  | `:8080`                             |
| `ENV`                  | Environment                   | `development` / `production`        |
| `DATABASE_URL`         | PostgreSQL connection string  | `postgres://user:pass@host:5432/db` |
| `REDIS_URL`            | Redis connection string       | `redis://localhost:6379`            |
| `JWT_PRIVATE_KEY_FILE` | Path to RS256 private key PEM | `./secrets/private.pem`             |
| `JWT_PUBLIC_KEY_FILE`  | Path to RS256 public key PEM  | `./secrets/public.pem`              |
| `JWT_ACCESS_EXPIRY`    | Access token TTL (seconds)    | `3600`                              |
| `JWT_REFRESH_EXPIRY`   | Refresh token TTL (seconds)   | `2592000`                           |
| `RATE_LIMIT_RPM`       | Requests per minute per IP    | `60`                                |

### Web App — 🔴 Planned (Build-time via Vite)

| Variable            | Description           | Example                        |
| ------------------- | --------------------- | ------------------------------ |
| `VITE_API_BASE_URL` | API base URL          | `http://localhost:8080/api/v1` |
| `VITE_APP_VERSION`  | Displayed in settings | `1.0.0`                        |

### Android Wrapper — 🔴 Planned

| Variable       | Description            |
| -------------- | ---------------------- |
| `WEBAPP_URL`   | URL loaded in WebView  |
| `API_BASE_URL` | Collector API endpoint |

---

## 14. Development Setup

### Prerequisites

- Go 1.24+
- Node.js 20+ (for future React app and extension)
- PostgreSQL 16 (running locally or in Docker)

### 1. Start PostgreSQL

Ensure PostgreSQL is running and accessible. The server reads its connection string from the `DATABASE_URL` environment variable in `server/.env`.

### 2. Run the Go server

Copy `server/.env.example` to `server/.env` and fill in the database URL and JWT key paths. Run migrations with `make migrate` (or `go run ./cmd/migrate`). Start the server with `make run` (or `air` for hot-reload). Server runs at `http://localhost:8080`.

### 3. Run the React web app — 🔴 Not yet available

Once `ui/core/` is created, install dependencies with `npm install` and start the dev server with `npm run dev`.

### 4. Run the browser extension — 🔴 Not yet available

Once `ui/wrappers/extension/` is created, install dependencies and run `npm run dev`.

### 5. Run Windows collector — 🔴 Not yet available

Once `ui/wrappers/windows/` is created, run the Go collector directly.

### Useful Commands (currently available)

| Command                         | Description                         |
| ------------------------------- | ----------------------------------- |
| `make run` (in `server/`)       | Start Go server with air hot-reload |
| `make migrate` (in `server/`)   | Run GORM auto-migrations            |
| `go build ./...` (in `server/`) | Compile check                       |

---

_This document covers the full technical implementation of the Cross-Device Time Tracker. As the system evolves, update the API Reference and Data Flow sections first — they are the primary contract between all components._
