## Timepad — Client Implementation

---

### 1. Architecture Mental Model

The React app in core is the **single, shared UI for all platforms**. It does not run "differently" per platform — it is one build that is:

| Surface                     | How it runs                                                                         |
| --------------------------- | ----------------------------------------------------------------------------------- |
| **Browser / Web**           | Served as a normal Vite SPA, hits the Go API directly                               |
| **Android**                 | Loaded inside an Android `WebView` inside `MainActivity`                            |
| **Windows**                 | Loaded inside a `WebView2` window opened from the system tray                       |
| **Browser Extension popup** | Runs its own tiny embedded React build (separate entry in `ui/wrappers/extension/`) |

There are no separate API routes per platform. The Go server treats all clients identically. The only "native" coupling is a tiny **JavaScript bridge** injected by the wrapper so the web app can read `device_key` and `platform` without needing separate credentials.

---

### 2. Full Folder Structure (core)

```
ui/core/
├── app/
│   ├── root.tsx                    ← HTML shell, font loading
│   ├── app.css                     ← Tailwind base + design tokens
│   ├── routes.ts                   ← React Router v7 route config
│   │
│   ├── routes/
│   │   ├── _auth.tsx               ← Unauthenticated layout (centered card)
│   │   ├── _auth.login.tsx
│   │   ├── _auth.register.tsx
│   │   │
│   │   ├── _app.tsx                ← Authenticated layout (sidebar + topbar)
│   │   ├── _app.dashboard.tsx
│   │   ├── _app.timeline.tsx
│   │   ├── _app.reports.tsx
│   │   ├── _app.categories.tsx
│   │   ├── _app.devices.tsx
│   │   └── _app.settings.tsx
│   │
│   ├── api/
│   │   ├── client.ts               ← Axios instance, interceptors, token refresh
│   │   ├── auth.ts
│   │   ├── events.ts
│   │   ├── timeline.ts
│   │   ├── summary.ts
│   │   ├── reports.ts
│   │   ├── categories.ts
│   │   ├── devices.ts
│   │   └── settings.ts
│   │
│   ├── store/
│   │   ├── auth.store.ts           ← Zustand: tokens, user, login/logout
│   │   ├── activity.store.ts       ← Zustand: timeline, summaries, selectedDate
│   │   └── settings.store.ts       ← Zustand: user settings
│   │
│   ├── types/
│   │   └── index.ts                ← All shared TS interfaces
│   │
│   ├── hooks/
│   │   ├── use-native-bridge.ts    ← Reads window.AndroidBridge / window.TimePadBridge
│   │   ├── use-auto-refresh.ts     ← 30-min setInterval re-fetch
│   │   ├── use-timeline.ts
│   │   ├── use-summary.ts
│   │   └── use-reports.ts
│   │
│   └── components/
│       ├── layout/
│       │   ├── sidebar.tsx
│       │   ├── topbar.tsx
│       │   └── mobile-nav.tsx
│       ├── animate-ui/
│       ├── dashboard/
│       ├── timeline/
│       └── ui/                     ← shadcn/ui primitives
```

---

### 3. Feature List & Owning Page

| #   | Feature                                                 | Route                 | Status  |
| --- | ------------------------------------------------------- | --------------------- | ------- |
| 1   | Register / Login / Token refresh                        | `/login`, `/register` | ✅ Done |
| 2   | Delete account                                          | Settings page         | ✅ Done |
| 3   | Dashboard — daily summary card                          | `/dashboard`          | ✅ Done |
| 4   | Dashboard — productivity ring chart                     | `/dashboard`          | ✅ Done |
| 5   | Dashboard — top apps list                               | `/dashboard`          | ✅ Done |
| 6   | Dashboard — device breakdown                            | `/dashboard`          | ✅ Done |
| 7   | Dashboard — peak hour indicator                         | `/dashboard`          | ✅ Done |
| 8   | Dashboard — date picker (navigate days)                 | `/dashboard`          | ✅ Done |
| 9   | Dashboard — manual refresh button                       | `/dashboard`          | ✅ Done |
| 10  | Timeline — horizontal day view                          | `/timeline`           | ✅ Done |
| 11  | Timeline — cursor pagination                            | `/timeline`           | ✅ Done |
| 12  | Timeline — mark event private                           | `/timeline`           | ✅ Done |
| 13  | Timeline — reassign category on event                   | `/timeline`           | ✅ Done |
| 14  | Timeline — filter by device                             | `/timeline`           | ✅ Done |
| 15  | Reports — custom date range                             | `/reports`            | ✅ Done |
| 16  | Reports — bar chart daily trend                         | `/reports`            | ✅ Done |
| 17  | Reports — category doughnut chart                       | `/reports`            | ✅ Done |
| 18  | Reports — app usage table                               | `/reports`            | ✅ Done |
| 19  | Reports — device usage breakdown                        | `/reports`            | ✅ Done |
| 20  | Categories — list system + user categories              | `/categories`         | ✅ Done |
| 21  | Categories — create / edit / delete                     | `/categories`         | ✅ Done |
| 22  | Categories — set productive / distraction / neutral     | `/categories`         | ✅ Done |
| 23  | Categories — rule builder (app name, URL, window title) | `/categories`         | ✅ Done |
| 24  | Devices — list registered devices                       | `/devices`            | ✅ Done |
| 25  | Devices — register new device + show device_key         | `/devices`            | ✅ Done |
| 26  | Devices — delete device (with confirmation)             | `/devices`            | ✅ Done |
| 27  | Settings — excluded apps / URLs list                    | `/settings`           | ✅ Done |
| 28  | Settings — idle threshold slider                        | `/settings`           | ✅ Done |
| 29  | Settings — tracking toggle                              | `/settings`           | ✅ Done |
| 30  | Settings — data retention selector                      | `/settings`           | ✅ Done |
| 31  | Settings — timezone picker                              | `/settings`           | ✅ Done |
| 32  | Native bridge detection (Android/Windows)               | Global hook           | ✅ Done |
| 33  | Auto-refresh every 30 min                               | Layout                | ✅ Done |

---

### 4. Routing Setup (routes.ts)

Using React Router v7 file-based nested routes:

```ts
// app/routes.ts
import {
    type RouteConfig,
    route,
    layout,
    index,
} from "@react-router/dev/routes";

export default [
    // Unauthenticated shell
    layout("routes/_auth.tsx", [
        route("login", "routes/_auth.login.tsx"),
        route("register", "routes/_auth.register.tsx"),
    ]),
    // Authenticated shell — PrivateRoute guard lives in _app.tsx
    layout("routes/_app.tsx", [
        index("routes/_app.dashboard.tsx"), // /
        route("timeline", "routes/_app.timeline.tsx"),
        route("reports", "routes/_app.reports.tsx"),
        route("categories", "routes/_app.categories.tsx"),
        route("devices", "routes/_app.devices.tsx"),
        route("settings", "routes/_app.settings.tsx"),
    ]),
] satisfies RouteConfig;
```

`_app.tsx` checks the auth store. If no token → `redirect("/login")`. This replaces a `PrivateRoute` component.

---

### 5. Native Bridge — How Android & Windows Connect

The React app never calls any Android or Windows API directly. Instead:

**Android side** (Kotlin):

```kotlin
// MainActivity.kt
webView.addJavascriptInterface(AndroidBridge(deviceKey), "TimePadBridge")
```

```kotlin
class AndroidBridge(private val deviceKey: String) {
    @JavascriptInterface fun getDeviceKey(): String = deviceKey
    @JavascriptInterface fun getPlatform(): String = "android"
}
```

**Windows side** (Go + WebView2):

```go
// webview2 window init — inject before page load
w.Init(`window.TimePadBridge = {
    getDeviceKey: () => "` + deviceKey + `",
    getPlatform:  () => "windows"
}`)
```

**React side** — one hook reads both:

```ts
// hooks/use-native-bridge.ts
export function useNativeBridge() {
    const bridge = (window as any).TimePadBridge;
    return {
        isNative: !!bridge,
        deviceKey: bridge?.getDeviceKey?.() ?? null,
        platform: bridge?.getPlatform?.() ?? "web",
    };
}
```

This hook is consumed in the Devices page (pre-fill device key on first open) and in the auth store (attach device key context to session).

---

### 6. API Client (`api/client.ts`)

```
Axios instance
  ├── baseURL = import.meta.env.VITE_API_BASE_URL
  ├── Request interceptor  → attach Authorization: Bearer <accessToken>
  └── Response interceptor → on 401: call POST /auth/refresh
                                       → on success: retry original request
                                       → on failure: authStore.logout() + redirect /login
```

All API modules (`auth.ts`, `timeline.ts`, etc.) are thin wrappers over this single instance that return typed responses using the interfaces from `types/index.ts`.

---

### 7. Zustand Stores

**`auth.store.ts`**

```
state:   { user, accessToken, refreshToken }
actions: login(email, pw) | logout() | refreshToken() | deleteAccount()
persist: localStorage via zustand/middleware/persist
```

**`activity.store.ts`**

```
state:   { timeline[], dailySummary, weeklySummary, selectedDate, isLoading }
actions: fetchTimeline(date) | fetchDailySummary(date) | fetchWeeklySummary(date)
         | setSelectedDate(date) | invalidate()
```

**`settings.store.ts`**

```
state:   { settings: UserSettings | null }
actions: fetchSettings() | updateSettings(partial)
```

---

### 8. TypeScript Types (`types/index.ts`)

All mirror the Go JSON shapes exactly:

```ts
(TimelineEntry,
    DailySummary,
    WeeklySummary,
    ReportData,
    Category,
    CategoryRule,
    Device,
    AppUsage,
    DeviceUsage,
    UserSettings);
```

---

### 9. Page-by-Page UI Components

#### `/login` & `/register`

- `AuthCard` — centered paper card, `#FBF9F5` bg, soft border `#E5E1D8`
- `InputField` — styled form input (focus ring in `#5B7C99`)
- `PrimaryButton` — `#5B7C99` bg, white text
- `FormError` — muted inline error text, no harsh red

---

#### `/dashboard`

- `DateNavigator` — `<` prev day / today / next day `>`, shows `"Today"` / formatted date
- `SummaryCard` — paper card, KG Teacher font for the headline number
    - `ActiveTimeStat` — big formatted duration (e.g. "4h 32m")
    - `ProductivityRing` — Recharts `PieChart` (donut) — productive / distraction / neutral slices in muted tones
    - `PeakHourBadge` — "Most active at 10 AM"
- `TopAppsCard` — ranked list, `AppRow` per item: app icon (placeholder), name, duration bar, category badge
- `DeviceBreakdownCard` — horizontal stacked bar (`BarChart`) per device
- `RefreshButton` — icon button top-right, spins on load, dispatches `invalidate()` + re-fetch
- `WeeklySpark` — 7-bar mini bar chart (Mon–Sun active seconds), built from weekly summary data

---

#### `/timeline`

- `DateNavigator` (reused)
- `DeviceFilterBar` — pill toggles, one per registered device, filters visible lanes
- `TimelineCanvas` — the main component:
    - X-axis: 00:00 → 23:59 (or current time if today)
    - Rows: one row per device (filtered)
    - `EventBar` — rounded rectangle, color from `category.color`, tooltip on hover/tap showing: app name, window title, start–end time, duration, category
    - `NowIndicator` — thin vertical line at current time (if viewing today)
- `EventDetailDrawer` — slides up from bottom (mobile) or right panel (desktop) on `EventBar` click:
    - Shows full metadata
    - `CategorySelect` — dropdown to reassign category (`PATCH /events/:id`)
    - `PrivacyToggle` — toggle `is_private`
    - `DeleteEventButton` — with confirmation
- Infinite scroll / "Load more" button at bottom drives cursor pagination

---

#### `/reports`

- `DateRangePicker` — two `<input type="date">` fields, styled as paper inputs
- `TotalsSummaryRow` — active time, idle time side by side
- `DailyTrendChart` — Recharts `BarChart` — X: date, Y: active seconds; uses `daily_active_trend` map
- `CategoryBreakdownChart` — Recharts `PieChart` (donut); uses `category_usage` map
- `AppUsageTable` — sortable table: App | Duration | % of total
- `DeviceUsageCard` — horizontal bars per device

---

#### `/categories`

- `CategoryList` — two sections: "System" (read-only) and "Your Categories" (editable)
- `CategoryRow` — color swatch, name, productive badge, edit/delete actions
- `CategoryFormSheet` — slide-in panel (shadcn `Sheet`):
    - Name input
    - Color picker (6 preset swatches + hex input)
    - Icon input (text, render as emoji or lucide icon name)
    - `ProductivityToggle` — 3-state: Productive / Distraction / Neutral
    - `RuleBuilder` — add/remove rules:
        - `RuleRow`: type select (`app_name` | `url_domain` | `window_title`) + operator select (`contains` | `equals` | `starts_with`) + value input
- `DeleteCategoryDialog` — shadcn `AlertDialog` warning that all events will be uncategorized

---

#### `/devices`

- `DeviceList` — card grid, one `DeviceCard` per device:
    - Platform icon (Android / Windows / Browser)
    - Name
    - Last seen relative time (Day.js `fromNow()`)
    - Delete button with confirmation
- `RegisterDeviceSheet` — slide-in panel:
    - Name input
    - Platform select (`android` | `windows` | `browser`)
    - On submit → `POST /devices` → show returned `device_key` in a copyable monospace box
- `NativeBridgeBanner` — shown only when `useNativeBridge().isNative === true`: "Running on [platform] — device key pre-configured"

---

#### `/settings`

- `TrackingToggleCard` — big on/off toggle for `tracking_enabled`, reassuring copy
- `ExcludedAppsCard` — tag input (add/remove app names from `excluded_apps[]`)
- `ExcludedUrlsCard` — same pattern for `excluded_urls[]`
- `IdleThresholdCard` — range slider 30s–600s, labelled "Mark idle after X minutes"
- `DataRetentionCard` — select: 30 / 90 / 180 / 365 days / Forever
- `TimezoneCard` — searchable select (list of IANA tz names)
- `DangerZoneCard` — delete account button → `AlertDialog` with "type your email to confirm"
- `VersionBadge` — shows `import.meta.env.VITE_APP_VERSION` at bottom

---

### 10. Shared/Primitive UI Components

All live in `components/ui/` and are built on **shadcn/ui** primitives with the Timepad design tokens applied:

| Component                            | Used on                                                   |
| ------------------------------------ | --------------------------------------------------------- |
| `Button` (primary, secondary, ghost) | Everywhere                                                |
| `Input`                              | Auth, Settings, Category form                             |
| `Sheet` (slide-in panel)             | Category form, Device registration                        |
| `AlertDialog`                        | Delete category, Delete device, Delete account            |
| `Select`                             | Category rule type, Platform select, Timezone             |
| `Badge`                              | Category tag, Productive label                            |
| `Skeleton`                           | Loading states on all data cards                          |
| `Tooltip`                            | Event bars, icon buttons                                  |
| `Separator`                          | Between card sections                                     |
| `Avatar`                             | User menu in topbar                                       |
| `DropdownMenu`                       | User menu                                                 |
| `Toggle`                             | Device filter bar, privacy toggle                         |
| `Slider`                             | Idle threshold                                            |
| `TagInput` (custom)                  | Excluded apps/URLs                                        |
| `ColorSwatch` (custom)               | Category color picker                                     |
| `EmptyState` (custom)                | No events yet, no categories, etc. — uses KG Teacher font |

---

### 11. Layout Components

**`_app.tsx` (authenticated shell)**

- Desktop: fixed left `Sidebar` (nav links) + scrollable `main`
- Mobile: bottom `MobileNav` (4 tabs: Dashboard, Timeline, Reports, Settings)
- `Topbar`: date, user avatar + dropdown (logout, settings)

**`_auth.tsx` (unauthenticated shell)**

- Centered column, paper-white card on `#F7F4EE` background
- KG Teacher logo at top

---

### 12. `useAutoRefresh` Hook (in `_app.tsx` layout)

```ts
// Polls every 30 minutes; also refetches when the tab becomes visible again
useEffect(() => {
  const id = setInterval(() => activityStore.invalidate(), 30 * 60 * 1000)
  document.addEventListener("visibilitychange", handleVisibility)
  return () => { clearInterval(id); document.removeEventListener(...) }
}, [])
```

---

### 13. Additional Dependencies to Add

Beyond what's already in package.json:

| Package                         | Purpose                                |
| ------------------------------- | -------------------------------------- |
| `axios`                         | HTTP client with interceptors          |
| `zustand`                       | State management                       |
| `recharts`                      | Charts (dashboard, reports)            |
| `dayjs`                         | Date formatting, `fromNow()`, timezone |
| `dayjs/plugin/timezone` + `utc` | Timezone-aware date math               |
| `@tanstack/react-virtual`       | Virtual scroll on long timelines       |

---

### 14. Build-time Environment Variables (`.env`)

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_VERSION=1.0.0
```

The Vite dev server proxies `/api` → `localhost:8080` so you avoid CORS during dev (set in `vite.config.ts`).

---

### 15. Implementation Order (Suggested Sprints)

| Sprint | What to build                                                                     |
| ------ | --------------------------------------------------------------------------------- |
| **1**  | `api/client.ts`, `types/index.ts`, auth store, login + register pages             |
| **2**  | Authenticated layout (sidebar, topbar, mobile nav), routing, `PrivateRoute` guard |
| **3**  | Dashboard page — daily summary, productivity ring, top apps, date navigator       |
| **4**  | Timeline page — `TimelineCanvas`, `EventBar`, `EventDetailDrawer`, pagination     |
| **5**  | Reports page — date range picker, all 4 charts                                    |
| **6**  | Categories page — list, form sheet, rule builder                                  |
| **7**  | Devices page — list, register sheet, native bridge banner                         |
| **8**  | Settings page — all setting cards, danger zone                                    |
| **9**  | Polish — loading skeletons, empty states, error boundaries, accessibility         |
| **10** | Native wrapper wiring — Android WebView + bridge, Windows WebView2 + bridge       |

---

### 16. Browser Extension (Separate Build — `ui/wrappers/extension/`)

The extension is **not** the full React app. It is a minimal Vite + React build with:

- A popup with a small login form (email + password, calls `POST /auth/login`, stores token in `chrome.storage.local`)
- A "Today: 2h 14m" summary pulled from `GET /summary/daily`
- A background Service Worker that listens to `chrome.tabs.onActivated` / `onUpdated`, buffers tab sessions ≥5 seconds, and POSTs them to `POST /events` every 60 seconds with the stored JWT + `device_key`

The popup shares the same design tokens but has nothing to do with the core routes.

---

The entire architecture is summarized as: **one React app served at a URL, two native wrappers that open that URL in a WebView and inject a tiny bridge object, and one separate extension build** — the Go server never knows or cares which surface is calling it.
