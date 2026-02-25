# Cross-Device Time Tracker — Technical Documentation

**Version:** 1.0.0  
**Last Updated:** February 2026  
**Status:** Draft

---

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Technology Stack](#2-technology-stack)
3. [Repository Structure](#3-repository-structure)
4. [Central Go Server](#4-central-go-server)
5. [Web App (Vue 3)](#5-web-app-vue-3)
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

**Presentation Layer** — A single Vue 3 web application that handles all UI: dashboards, timelines, reports, and settings. It is served as a standalone web app and also embedded via WebView inside the Android and Windows native wrappers.

```
┌─────────────────────────────────────────────────────────┐
│                      USER DEVICES                       │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Android    │  │   Windows    │  │   Browser    │  │
│  │  Native App  │  │  Tray App    │  │  Extension   │  │
│  │  (Kotlin)    │  │    (Go)      │  │    (JS)      │  │
│  │              │  │              │  │              │  │
│  │  WebView ──► │  │  WebView ──► │  │  Opens Web ► │  │
│  │  Vue App     │  │  Vue App     │  │  App in Tab  │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │           │
└─────────┼─────────────────┼─────────────────┼───────────┘
          │   Activity      │   Events        │
          │   POST /events  │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────┐
│                  CENTRAL GO SERVER                      │
│                                                         │
│   ┌──────────┐  ┌──────────┐  ┌──────────────────────┐ │
│   │   API    │  │ Business │  │   Background Jobs    │ │
│   │  Router  │  │  Logic   │  │  (Categorizer, Sync, │ │
│   │  (Gin)   │  │  Layer   │  │   Aggregator)        │ │
│   └──────────┘  └──────────┘  └──────────────────────┘ │
│                                                         │
│   ┌──────────────────────┐   ┌──────────────────────┐  │
│   │     PostgreSQL       │   │       Redis          │  │
│   │  (Primary Storage)   │   │  (Cache & Sessions)  │  │
│   └──────────────────────┘   └──────────────────────┘  │
└─────────────────────────────────────────────────────────┘
          ▲
          │  REST API (JSON)
          ▼
┌─────────────────────────────────────────────────────────┐
│                  VUE 3 WEB APP                          │
│          (Served standalone + embedded in WebView)      │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Technology Stack

### Central Server

| Component        | Choice            | Reason                                    |
| ---------------- | ----------------- | ----------------------------------------- |
| Language         | Go 1.22+          | Performant, low memory, great concurrency |
| HTTP Framework   | Gin               | Fast routing, middleware support          |
| ORM              | GORM              | Clean Go ORM, good migration support      |
| Database         | PostgreSQL 16     | Reliable, JSONB support, strong typing    |
| Cache / Sessions | Redis 7           | Fast session store, pub/sub for sync      |
| Auth             | JWT (RS256)       | Stateless, works across devices           |
| Task Queue       | Asynq             | Go-native, Redis-backed background jobs   |
| WebSocket        | Gorilla WebSocket | Real-time sync push                       |

### Web App (Dashboard UI)

| Component            | Choice                 | Reason                                   |
| -------------------- | ---------------------- | ---------------------------------------- |
| Framework            | Vue 3                  | Composition API, reactive, lightweight   |
| Build Tool           | Vite                   | Fast HMR, excellent Vue support          |
| State Management     | Pinia                  | Official Vue store, simple and typed     |
| Routing              | Vue Router 4           | Official, well-integrated                |
| HTTP Client          | Axios                  | Interceptors, easy auth header injection |
| Charts               | Chart.js + vue-chartjs | Flexible, well-documented                |
| UI Component Library | shadcn-vue             | Accessible, unstyled, customizable       |
| CSS                  | Tailwind CSS           | Utility-first, easy to customize         |
| Date Handling        | Day.js                 | Lightweight Moment.js alternative        |
| TypeScript           | Yes                    | Type safety across the entire app        |

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

```
time-tracker/
├── server/                         # Central Go server
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/           # Route handlers
│   │   │   ├── middleware/         # Auth, logging, rate-limit
│   │   │   └── router.go
│   │   ├── models/                 # GORM models
│   │   ├── services/               # Business logic
│   │   ├── jobs/                   # Asynq background tasks
│   │   └── sync/                   # WebSocket sync hub
│   ├── migrations/                 # SQL migration files
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── webapp/                         # Vue 3 dashboard
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   │   ├── timeline/
│   │   │   ├── charts/
│   │   │   └── ui/
│   │   ├── views/
│   │   │   ├── DashboardView.vue
│   │   │   ├── TimelineView.vue
│   │   │   ├── ReportsView.vue
│   │   │   └── SettingsView.vue
│   │   ├── stores/                 # Pinia stores
│   │   │   ├── auth.ts
│   │   │   ├── activity.ts
│   │   │   └── settings.ts
│   │   ├── api/                    # Axios client + typed API calls
│   │   │   ├── client.ts
│   │   │   ├── activity.ts
│   │   │   └── auth.ts
│   │   ├── types/                  # Global TypeScript types
│   │   ├── composables/            # Reusable Vue composables
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── App.vue
│   │   └── main.ts
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── clients/
│   ├── android/                    # Kotlin Android app
│   │   ├── app/
│   │   │   └── src/main/
│   │   │       ├── java/com/timetracker/
│   │   │       │   ├── services/
│   │   │       │   │   ├── ActivityCollectorService.kt
│   │   │       │   │   └── SyncService.kt
│   │   │       │   ├── webview/
│   │   │       │   │   └── MainActivity.kt
│   │   │       │   └── api/
│   │   │       └── AndroidManifest.xml
│   │   └── build.gradle
│   │
│   ├── windows/                    # Go Windows tray app
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── collector/
│   │   │   └── windows_collector.go
│   │   ├── tray/
│   │   │   └── tray.go
│   │   ├── webview/
│   │   │   └── webview.go
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── extension/                  # Browser extension
│       ├── src/
│       │   ├── background/
│       │   │   └── service-worker.ts
│       │   ├── content/
│       │   │   └── tracker.ts
│       │   └── popup/
│       │       └── Popup.vue
│       ├── manifest.json
│       ├── vite.config.ts
│       └── package.json
│
└── docker/
    ├── docker-compose.yml
    └── docker-compose.prod.yml
```

---

## 4. Central Go Server

### 4.1 Server Entry Point

```go
// server/cmd/server/main.go
package main

import (
    "log"
    "time-tracker/internal/api"
    "time-tracker/internal/config"
    "time-tracker/internal/jobs"
    "time-tracker/internal/sync"
)

func main() {
    cfg := config.Load()

    db := config.ConnectDB(cfg)
    redis := config.ConnectRedis(cfg)

    syncHub := sync.NewHub(redis)
    go syncHub.Run()

    jobServer := jobs.NewServer(redis)
    go jobServer.Run()

    router := api.NewRouter(cfg, db, redis, syncHub)
    log.Fatal(router.Run(cfg.ServerAddr))
}
```

### 4.2 Router Setup

```go
// server/internal/api/router.go
func NewRouter(cfg *config.Config, db *gorm.DB, rdb *redis.Client, hub *sync.Hub) *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery())
    r.Use(middleware.CORS())
    r.Use(middleware.RateLimit(rdb))

    v1 := r.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        auth.POST("/register", handlers.Register(db))
        auth.POST("/login", handlers.Login(db, cfg))
        auth.POST("/refresh", handlers.Refresh(cfg))

        // Protected routes
        protected := v1.Group("/")
        protected.Use(middleware.Auth(cfg))
        {
            protected.POST("/events", handlers.IngestEvents(db, hub))
            protected.GET("/events", handlers.GetEvents(db))
            protected.GET("/timeline", handlers.GetTimeline(db))
            protected.GET("/summary/daily", handlers.GetDailySummary(db))
            protected.GET("/summary/weekly", handlers.GetWeeklySummary(db))
            protected.GET("/reports", handlers.GetReports(db))
            protected.GET("/categories", handlers.GetCategories(db))
            protected.PATCH("/events/:id", handlers.EditEvent(db))
            protected.DELETE("/events/:id", handlers.DeleteEvent(db))
            protected.GET("/ws", handlers.WebSocketHandler(hub))
            protected.GET("/devices", handlers.GetDevices(db))
            protected.GET("/settings", handlers.GetSettings(db))
            protected.PUT("/settings", handlers.UpdateSettings(db))
        }
    }

    return r
}
```

### 4.3 Models

```go
// server/internal/models/activity_event.go
type ActivityEvent struct {
    ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
    DeviceID     uuid.UUID  `gorm:"type:uuid;not null;index"`
    AppName      string     `gorm:"not null"`
    WindowTitle  string
    URL          string
    CategoryID   *uuid.UUID `gorm:"type:uuid"`
    StartTime    time.Time  `gorm:"not null;index"`
    EndTime      time.Time  `gorm:"not null"`
    DurationSecs int        `gorm:"not null"`
    IsIdle       bool       `gorm:"default:false"`
    IsPrivate    bool       `gorm:"default:false"`
    RawMeta      datatypes.JSON
    CreatedAt    time.Time
}

// server/internal/models/device.go
type Device struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
    Name       string    `gorm:"not null"`
    Platform   string    `gorm:"not null"` // "android" | "windows" | "browser"
    DeviceKey  string    `gorm:"uniqueIndex;not null"`
    LastSeenAt time.Time
    CreatedAt  time.Time
}

// server/internal/models/category.go
type Category struct {
    ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID   *uuid.UUID
    Name     string `gorm:"not null"`
    Color    string
    Icon     string
    IsSystem bool `gorm:"default:false"`
    Rules    datatypes.JSON // matching rules: app name, URL pattern, etc.
}
```

### 4.4 Event Ingestion Handler

```go
// server/internal/api/handlers/events.go
type IngestPayload struct {
    DeviceKey string          `json:"device_key" binding:"required"`
    Events    []EventInput    `json:"events" binding:"required,min=1"`
}

type EventInput struct {
    AppName     string    `json:"app_name" binding:"required"`
    WindowTitle string    `json:"window_title"`
    URL         string    `json:"url"`
    StartTime   time.Time `json:"start_time" binding:"required"`
    EndTime     time.Time `json:"end_time" binding:"required"`
    IsIdle      bool      `json:"is_idle"`
}

func IngestEvents(db *gorm.DB, hub *sync.Hub) gin.HandlerFunc {
    return func(c *gin.Context) {
        var payload IngestPayload
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        userID := middleware.GetUserID(c)

        // Resolve device
        var device models.Device
        result := db.Where("device_key = ? AND user_id = ?", payload.DeviceKey, userID).
            First(&device)
        if result.Error != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown device"})
            return
        }

        // Batch insert events
        events := make([]models.ActivityEvent, 0, len(payload.Events))
        for _, e := range payload.Events {
            duration := int(e.EndTime.Sub(e.StartTime).Seconds())
            if duration <= 0 {
                continue
            }
            events = append(events, models.ActivityEvent{
                UserID:       userID,
                DeviceID:     device.ID,
                AppName:      e.AppName,
                WindowTitle:  e.WindowTitle,
                URL:          e.URL,
                StartTime:    e.StartTime,
                EndTime:      e.EndTime,
                DurationSecs: duration,
                IsIdle:       e.IsIdle,
            })
        }

        if err := db.CreateInBatches(events, 100).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save events"})
            return
        }

        // Notify WebSocket clients of new data
        hub.BroadcastToUser(userID, sync.Message{Type: "events_updated"})

        // Queue categorization job
        // jobs.EnqueueCategorize(userID)

        c.JSON(http.StatusCreated, gin.H{"inserted": len(events)})
    }
}
```

### 4.5 Background Jobs (Asynq)

```go
// server/internal/jobs/categorizer.go
const TaskCategorize = "task:categorize"

func HandleCategorize(db *gorm.DB) asynq.HandlerFunc {
    return func(ctx context.Context, t *asynq.Task) error {
        var payload struct{ UserID string }
        json.Unmarshal(t.Payload(), &payload)

        // Fetch uncategorized events
        var events []models.ActivityEvent
        db.Where("user_id = ? AND category_id IS NULL", payload.UserID).
            Order("start_time desc").
            Limit(500).
            Find(&events)

        // Load user + system category rules
        var categories []models.Category
        db.Where("user_id = ? OR is_system = true", payload.UserID).Find(&categories)

        for i, e := range events {
            for _, cat := range categories {
                if matchesCategory(e, cat) {
                    events[i].CategoryID = &cat.ID
                    break
                }
            }
        }

        // Batch update
        for _, e := range events {
            if e.CategoryID != nil {
                db.Model(&e).Update("category_id", e.CategoryID)
            }
        }

        return nil
    }
}
```

---

## 5. Web App (Vue 3)

### 5.1 Vite Config

```typescript
// webapp/vite.config.ts
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        port: 5173,
        proxy: {
            "/api": {
                target: "http://localhost:8080",
                changeOrigin: true,
            },
            "/ws": {
                target: "ws://localhost:8080",
                ws: true,
            },
        },
    },
});
```

### 5.2 Axios API Client

```typescript
// webapp/src/api/client.ts
import axios from "axios";
import { useAuthStore } from "@/stores/auth";
import router from "@/router";

const apiClient = axios.create({
    baseURL: "/api/v1",
    timeout: 10000,
});

// Attach JWT to every request
apiClient.interceptors.request.use((config) => {
    const auth = useAuthStore();
    if (auth.token) {
        config.headers.Authorization = `Bearer ${auth.token}`;
    }
    return config;
});

// Handle token expiry
apiClient.interceptors.response.use(
    (res) => res,
    async (err) => {
        if (err.response?.status === 401) {
            const auth = useAuthStore();
            const refreshed = await auth.refreshToken();
            if (refreshed) {
                return apiClient(err.config);
            }
            auth.logout();
            router.push("/login");
        }
        return Promise.reject(err);
    },
);

export default apiClient;
```

### 5.3 Pinia Auth Store

```typescript
// webapp/src/stores/auth.ts
import { defineStore } from "pinia";
import { ref, computed } from "vue";
import apiClient from "@/api/client";

export const useAuthStore = defineStore("auth", () => {
    const token = ref<string | null>(localStorage.getItem("token"));
    const refreshToken = ref<string | null>(
        localStorage.getItem("refresh_token"),
    );
    const user = ref<User | null>(null);

    const isAuthenticated = computed(() => !!token.value);

    async function login(email: string, password: string) {
        const { data } = await apiClient.post("/auth/login", {
            email,
            password,
        });
        token.value = data.access_token;
        refreshToken.value = data.refresh_token;
        user.value = data.user;
        localStorage.setItem("token", data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);
    }

    async function refresh() {
        try {
            const { data } = await apiClient.post("/auth/refresh", {
                refresh_token: refreshToken.value,
            });
            token.value = data.access_token;
            localStorage.setItem("token", data.access_token);
            return true;
        } catch {
            return false;
        }
    }

    function logout() {
        token.value = null;
        refreshToken.value = null;
        user.value = null;
        localStorage.removeItem("token");
        localStorage.removeItem("refresh_token");
    }

    return { token, user, isAuthenticated, login, refresh, logout };
});
```

### 5.4 Pinia Activity Store

```typescript
// webapp/src/stores/activity.ts
import { defineStore } from "pinia";
import { ref } from "vue";
import { getTimeline, getDailySummary } from "@/api/activity";
import type { TimelineEntry, DailySummary } from "@/types";

export const useActivityStore = defineStore("activity", () => {
    const timeline = ref<TimelineEntry[]>([]);
    const summary = ref<DailySummary | null>(null);
    const loading = ref(false);
    const selectedDate = ref(new Date());

    async function fetchTimeline(date: Date) {
        loading.value = true;
        try {
            const { data } = await getTimeline(date);
            timeline.value = data;
        } finally {
            loading.value = false;
        }
    }

    async function fetchSummary(date: Date) {
        const { data } = await getDailySummary(date);
        summary.value = data;
    }

    return {
        timeline,
        summary,
        loading,
        selectedDate,
        fetchTimeline,
        fetchSummary,
    };
});
```

### 5.5 WebSocket Composable

```typescript
// webapp/src/composables/useRealtimeSync.ts
import { onMounted, onUnmounted } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useActivityStore } from "@/stores/activity";

export function useRealtimeSync() {
    let ws: WebSocket | null = null;
    const auth = useAuthStore();
    const activity = useActivityStore();

    function connect() {
        const protocol = location.protocol === "https:" ? "wss" : "ws";
        ws = new WebSocket(
            `${protocol}://${location.host}/api/v1/ws?token=${auth.token}`,
        );

        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (msg.type === "events_updated") {
                activity.fetchTimeline(activity.selectedDate);
                activity.fetchSummary(activity.selectedDate);
            }
        };

        ws.onclose = () => {
            // Reconnect after 3s
            setTimeout(connect, 3000);
        };
    }

    onMounted(connect);
    onUnmounted(() => ws?.close());
}
```

### 5.6 Router

```typescript
// webapp/src/router/index.ts
import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const routes = [
    {
        path: "/login",
        component: () => import("@/views/LoginView.vue"),
        meta: { public: true },
    },
    { path: "/", redirect: "/dashboard" },
    {
        path: "/dashboard",
        component: () => import("@/views/DashboardView.vue"),
    },
    { path: "/timeline", component: () => import("@/views/TimelineView.vue") },
    { path: "/reports", component: () => import("@/views/ReportsView.vue") },
    { path: "/settings", component: () => import("@/views/SettingsView.vue") },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach((to) => {
    const auth = useAuthStore();
    if (!to.meta.public && !auth.isAuthenticated) {
        return "/login";
    }
});

export default router;
```

### 5.7 Key TypeScript Types

```typescript
// webapp/src/types/index.ts
export interface TimelineEntry {
    id: string;
    appName: string;
    windowTitle: string;
    url: string;
    category: Category;
    device: Device;
    startTime: string; // ISO 8601
    endTime: string;
    durationSecs: number;
    isIdle: boolean;
}

export interface DailySummary {
    date: string;
    totalActiveSecs: number;
    totalIdleSecs: number;
    productiveSecs: number;
    distractionSecs: number;
    topApps: AppUsage[];
    peakHour: number;
    deviceBreakdown: DeviceUsage[];
}

export interface Category {
    id: string;
    name: string;
    color: string;
    icon: string;
}

export interface Device {
    id: string;
    name: string;
    platform: "android" | "windows" | "browser";
}

export interface AppUsage {
    appName: string;
    category: Category;
    totalSecs: number;
}
```

---

## 6. Client Collectors

### 6.1 Android Collector (Kotlin)

The Android collector runs as a persistent Foreground Service to survive background restrictions. It uses `UsageStatsManager` to query app usage in polling intervals.

```kotlin
// clients/android/app/src/main/java/com/timetracker/services/ActivityCollectorService.kt
class ActivityCollectorService : Service() {

    private val pollingIntervalMs = 30_000L  // 30 seconds
    private val handler = Handler(Looper.getMainLooper())
    private val apiClient = RetrofitClient.create()

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        startForeground(NOTIF_ID, buildNotification())
        schedulePolling()
        return START_STICKY
    }

    private fun schedulePolling() {
        handler.postDelayed({
            collectAndSend()
            schedulePolling()
        }, pollingIntervalMs)
    }

    private fun collectAndSend() {
        val usm = getSystemService(USAGE_STATS_SERVICE) as UsageStatsManager
        val now = System.currentTimeMillis()
        val since = now - pollingIntervalMs

        val stats = usm.queryUsageStats(
            UsageStatsManager.INTERVAL_BEST, since, now
        )

        val events = stats
            .filter { it.totalTimeInForeground > 0 }
            .map { stat ->
                EventInput(
                    appName = getAppLabel(stat.packageName),
                    startTime = Instant.ofEpochMilli(stat.lastTimeUsed - stat.totalTimeInForeground),
                    endTime = Instant.ofEpochMilli(stat.lastTimeUsed),
                    isIdle = false,
                )
            }

        if (events.isNotEmpty()) {
            apiClient.ingestEvents(IngestPayload(
                deviceKey = DeviceRegistry.getKey(this),
                events = events
            )).enqueue(/* handle errors */)
        }
    }
}
```

**Required Android Manifest Permissions:**

```xml
<uses-permission android:name="android.permission.PACKAGE_USAGE_STATS"
    tools:ignore="ProtectedPermissions"/>
<uses-permission android:name="android.permission.FOREGROUND_SERVICE"/>
<uses-permission android:name="android.permission.POST_NOTIFICATIONS"/>
<uses-permission android:name="android.permission.INTERNET"/>
```

**WebView for Dashboard:**

```kotlin
// clients/android/app/src/main/java/com/timetracker/webview/MainActivity.kt
class MainActivity : AppCompatActivity() {
    private lateinit var webView: WebView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        webView = findViewById(R.id.webview)
        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
        }
        webView.addJavascriptInterface(AndroidBridge(this), "AndroidBridge")
        webView.loadUrl(BuildConfig.WEBAPP_URL)  // e.g. "https://app.timetracker.io"
    }
}

// Bridge for passing device token from native to web app
class AndroidBridge(private val context: Context) {
    @JavascriptInterface
    fun getDeviceKey(): String = DeviceRegistry.getKey(context)

    @JavascriptInterface
    fun getPlatform(): String = "android"
}
```

---

### 6.2 Windows Collector (Go)

```go
// clients/windows/collector/windows_collector.go
package collector

import (
    "golang.org/x/sys/windows"
    "time"
    "unsafe"
)

var (
    user32               = windows.NewLazyDLL("user32.dll")
    procGetForeground    = user32.NewProc("GetForegroundWindow")
    procGetWindowText    = user32.NewProc("GetWindowTextW")
)

type ActiveWindow struct {
    Title     string
    ProcessID uint32
}

func GetActiveWindow() (ActiveWindow, error) {
    hwnd, _, _ := procGetForeground.Call()
    if hwnd == 0 {
        return ActiveWindow{}, nil
    }

    var pid uint32
    windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid)

    buf := make([]uint16, 256)
    procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)

    return ActiveWindow{
        Title:     windows.UTF16ToString(buf),
        ProcessID: pid,
    }, nil
}

func GetProcessName(pid uint32) (string, error) {
    handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
    if err != nil {
        return "", err
    }
    defer windows.CloseHandle(handle)

    var buf [windows.MAX_PATH]uint16
    size := uint32(len(buf))
    err = windows.QueryFullProcessImageName(handle, 0, &buf[0], &size)
    if err != nil {
        return "", err
    }
    fullPath := windows.UTF16ToString(buf[:size])
    // Extract just the exe name
    parts := strings.Split(fullPath, `\`)
    return parts[len(parts)-1], nil
}
```

```go
// clients/windows/cmd/main.go — polling loop
func runCollector(apiClient *api.Client, deviceKey string) {
    ticker := time.NewTicker(30 * time.Second)
    var currentApp string
    var sessionStart time.Time
    pending := []api.EventInput{}

    for range ticker.C {
        win, _ := collector.GetActiveWindow()
        appName, _ := collector.GetProcessName(win.ProcessID)
        idle := idledetect.IsIdle(300) // 5 min idle threshold

        now := time.Now()

        if appName != currentApp && currentApp != "" {
            pending = append(pending, api.EventInput{
                AppName:     currentApp,
                WindowTitle: win.Title,
                StartTime:   sessionStart,
                EndTime:     now,
                IsIdle:      idle,
            })
            if len(pending) >= 10 {
                apiClient.IngestEvents(deviceKey, pending)
                pending = pending[:0]
            }
        }

        currentApp = appName
        sessionStart = now
    }
}
```

---

### 6.3 Browser Extension

```typescript
// clients/extension/src/background/service-worker.ts
const API_BASE = "https://api.timetracker.io/api/v1";
let currentTab: { url: string; title: string; startTime: number } | null = null;
const eventBuffer: EventInput[] = [];

chrome.tabs.onActivated.addListener(async ({ tabId }) => {
    const tab = await chrome.tabs.get(tabId);
    flushCurrentTab();
    currentTab = {
        url: tab.url ?? "",
        title: tab.title ?? "",
        startTime: Date.now(),
    };
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === "complete" && tab.active) {
        flushCurrentTab();
        currentTab = {
            url: tab.url ?? "",
            title: tab.title ?? "",
            startTime: Date.now(),
        };
    }
});

function flushCurrentTab() {
    if (!currentTab) return;
    const duration = (Date.now() - currentTab.startTime) / 1000;
    if (duration < 5) return; // Ignore sub-5s visits

    eventBuffer.push({
        app_name: new URL(currentTab.url).hostname,
        window_title: currentTab.title,
        url: currentTab.url,
        start_time: new Date(currentTab.startTime).toISOString(),
        end_time: new Date().toISOString(),
        is_idle: false,
    });
}

// Flush buffer every 60s
setInterval(async () => {
    if (eventBuffer.length === 0) return;
    const token = (await chrome.storage.local.get("token")).token;
    const deviceKey = (await chrome.storage.local.get("device_key")).device_key;

    await fetch(`${API_BASE}/events`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
            device_key: deviceKey,
            events: [...eventBuffer],
        }),
    });
    eventBuffer.length = 0;
}, 60_000);
```

---

## 7. Database Design

### 7.1 Schema

```sql
-- Users
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name  TEXT,
    timezone      TEXT NOT NULL DEFAULT 'UTC',
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Devices
CREATE TABLE devices (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    platform     TEXT NOT NULL CHECK (platform IN ('android', 'windows', 'browser')),
    device_key   TEXT UNIQUE NOT NULL,
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Categories
CREATE TABLE categories (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = system category
    name      TEXT NOT NULL,
    color     TEXT NOT NULL DEFAULT '#6B7280',
    icon      TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    rules     JSONB DEFAULT '[]'
);

-- Activity Events (primary data table)
CREATE TABLE activity_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id      UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    app_name       TEXT NOT NULL,
    window_title   TEXT,
    url            TEXT,
    category_id    UUID REFERENCES categories(id),
    start_time     TIMESTAMPTZ NOT NULL,
    end_time       TIMESTAMPTZ NOT NULL,
    duration_secs  INTEGER NOT NULL,
    is_idle        BOOLEAN DEFAULT FALSE,
    is_private     BOOLEAN DEFAULT FALSE,
    raw_meta       JSONB,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_events_user_start ON activity_events(user_id, start_time DESC);
CREATE INDEX idx_events_device ON activity_events(device_id);
CREATE INDEX idx_events_category ON activity_events(category_id);
CREATE INDEX idx_events_app_name ON activity_events(user_id, app_name);

-- User settings
CREATE TABLE user_settings (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    excluded_apps    TEXT[] DEFAULT '{}',
    excluded_urls    TEXT[] DEFAULT '{}',
    idle_threshold   INTEGER DEFAULT 300,
    tracking_enabled BOOLEAN DEFAULT TRUE,
    data_retention_days INTEGER DEFAULT 365,
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
```

### 7.2 Category Rules Schema (JSONB)

```json
[
    { "type": "app_name", "op": "contains", "value": "code" },
    { "type": "url_domain", "op": "equals", "value": "github.com" },
    { "type": "window_title", "op": "startsWith", "value": "YouTube" }
]
```

---

## 8. API Reference

### Base URL

```
https://api.timetracker.io/api/v1
```

### Authentication

All protected endpoints require:

```
Authorization: Bearer <access_token>
```

---

### Auth Endpoints

#### `POST /auth/register`

```json
// Request
{ "email": "user@example.com", "password": "s3cure!", "display_name": "Alice" }

// Response 201
{ "user": { "id": "uuid", "email": "...", "display_name": "Alice" } }
```

#### `POST /auth/login`

```json
// Request
{ "email": "user@example.com", "password": "s3cure!" }

// Response 200
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 3600,
  "user": { "id": "uuid", "email": "...", "display_name": "Alice" }
}
```

#### `POST /auth/refresh`

```json
// Request
{ "refresh_token": "eyJ..." }

// Response 200
{ "access_token": "eyJ...", "expires_in": 3600 }
```

---

### Events Endpoints

#### `POST /events` — Ingest activity events (called by collectors)

```json
// Request
{
  "device_key": "android-xxxx-uuid",
  "events": [
    {
      "app_name": "Visual Studio Code",
      "window_title": "router.go — time-tracker",
      "url": "",
      "start_time": "2026-02-25T09:00:00Z",
      "end_time": "2026-02-25T09:30:00Z",
      "is_idle": false
    }
  ]
}

// Response 201
{ "inserted": 1 }
```

#### `GET /timeline?date=2026-02-25&device_id=optional`

```json
// Response 200
{
    "date": "2026-02-25",
    "entries": [
        {
            "id": "uuid",
            "app_name": "Visual Studio Code",
            "window_title": "router.go",
            "category": { "id": "uuid", "name": "Coding", "color": "#3B82F6" },
            "device": {
                "id": "uuid",
                "name": "Work Laptop",
                "platform": "windows"
            },
            "start_time": "2026-02-25T09:00:00Z",
            "end_time": "2026-02-25T09:30:00Z",
            "duration_secs": 1800,
            "is_idle": false
        }
    ]
}
```

#### `GET /summary/daily?date=2026-02-25`

```json
// Response 200
{
  "date": "2026-02-25",
  "total_active_secs": 28800,
  "total_idle_secs": 3600,
  "productive_secs": 21600,
  "distraction_secs": 7200,
  "peak_hour": 10,
  "top_apps": [
    { "app_name": "VS Code", "category": {...}, "total_secs": 9000 }
  ],
  "category_breakdown": [
    { "category": {...}, "total_secs": 21600, "percentage": 75.0 }
  ],
  "device_breakdown": [
    { "device": {...}, "total_secs": 18000 }
  ]
}
```

#### `PATCH /events/:id` — Edit a single event

```json
// Request
{
  "category_id": "uuid",
  "window_title": "Updated title",
  "is_private": true
}

// Response 200
{ "event": { ...updated event... } }
```

#### `DELETE /events/:id`

```
Response 204 No Content
```

---

### Devices Endpoints

#### `GET /devices`

```json
// Response 200
{
    "devices": [
        {
            "id": "uuid",
            "name": "Work Laptop",
            "platform": "windows",
            "device_key": "windows-xxxx",
            "last_seen_at": "2026-02-25T14:30:00Z"
        }
    ]
}
```

#### `POST /devices/register`

```json
// Request
{ "name": "My Phone", "platform": "android" }

// Response 201
{ "device_key": "android-generated-uuid", "device_id": "uuid" }
```

---

### Settings Endpoints

#### `GET /settings`

```json
// Response 200
{
    "excluded_apps": ["Slack", "Discord"],
    "excluded_urls": ["mail.google.com"],
    "idle_threshold": 300,
    "tracking_enabled": true,
    "data_retention_days": 365
}
```

#### `PUT /settings`

```json
// Request — send only fields to update
{
    "idle_threshold": 600,
    "excluded_apps": ["Slack"]
}
```

---

### WebSocket

#### `GET /ws?token=<access_token>`

Upgrade to WebSocket. Server pushes messages:

```json
{ "type": "events_updated" }
{ "type": "sync_complete", "device_id": "uuid" }
{ "type": "category_updated" }
```

---

## 9. Authentication & Security

### JWT Strategy

- **Access Token:** RS256, 1-hour expiry. Signed with a private key kept only on the server.
- **Refresh Token:** Stored in the database with a 30-day expiry. Can be revoked server-side.
- **Device Key:** A UUID assigned at device registration. Included in event POST requests alongside the JWT to bind events to a specific device.

### WebView Security

The Android and Windows WebView embeds the web app from a trusted origin (`https://app.timetracker.io`). Native bridges (`AndroidBridge`, Windows equivalent) only expose the minimum needed: the device key and platform identifier. No raw API credentials are passed through the bridge.

### Data Privacy

- Events marked `is_private: true` are stored encrypted at rest (AES-256) and never appear in exports.
- The `excluded_apps` and `excluded_urls` settings are enforced server-side: matching events are rejected at ingestion time.
- Users can delete all data via `DELETE /account` which hard-deletes all records.

### Rate Limiting

All endpoints are rate-limited via Redis using a token bucket algorithm. The `/events` ingestion endpoint allows 60 requests/minute per device. Auth endpoints are limited to 10 requests/minute per IP.

---

## 10. Data Flow

### Event Ingestion Flow

```
[Collector detects activity]
        │
        ▼
[Buffer events locally (30s–60s window)]
        │
        ▼
[POST /api/v1/events with device_key + JWT]
        │
        ▼
[Server validates JWT + device_key]
        │
        ▼
[Deduplicate against recent events]
        │
        ▼
[Batch INSERT into activity_events table]
        │
        ├──► [Enqueue Asynq categorization job]
        │
        └──► [WebSocket broadcast: "events_updated" to user's sessions]
                        │
                        ▼
              [Vue app refetches timeline/summary]
```

### Timeline Query Flow

```
[User opens Timeline view]
        │
        ▼
[GET /timeline?date=2026-02-25]
        │
        ▼
[Server queries activity_events WHERE user_id + date range]
        │
        ▼
[JOIN categories, devices]
        │
        ▼
[Order by start_time ASC]
        │
        ▼
[Return unified timeline across all devices]
        │
        ▼
[Vue renders merged chronological view]
```

---

## 11. Cross-Device Sync

All three collectors report to the same central server independently. There is no device-to-device communication. The server merges events from all devices into one timeline, tagged with `device_id` so the UI can filter or color-code by device.

**Sync timing:**

- Android: polls every 30 seconds
- Windows: polls every 30 seconds
- Browser Extension: flushes buffer every 60 seconds

On first install, each client registers a device via `POST /devices/register` and persists the returned `device_key` locally. This key is used in every subsequent event POST.

**Conflict handling:** Overlapping time ranges across different devices are allowed and stored as-is — the timeline view renders them as parallel lanes. Overlapping events on the _same_ device are deduplicated at ingestion using a 5-second overlap threshold.

---

## 12. Deployment

### Docker Compose (Development)

```yaml
# docker/docker-compose.yml
version: "3.9"

services:
    postgres:
        image: postgres:16-alpine
        environment:
            POSTGRES_DB: timetracker
            POSTGRES_USER: tt_user
            POSTGRES_PASSWORD: dev_password
        ports:
            - "5432:5432"
        volumes:
            - pgdata:/var/lib/postgresql/data

    redis:
        image: redis:7-alpine
        ports:
            - "6379:6379"

    server:
        build:
            context: ../server
            dockerfile: Dockerfile
        ports:
            - "8080:8080"
        environment:
            - DATABASE_URL=postgres://tt_user:dev_password@postgres:5432/timetracker
            - REDIS_URL=redis://redis:6379
            - JWT_PRIVATE_KEY_FILE=/secrets/private.pem
            - JWT_PUBLIC_KEY_FILE=/secrets/public.pem
        depends_on:
            - postgres
            - redis
        volumes:
            - ./secrets:/secrets:ro

    webapp:
        build:
            context: ../webapp
            dockerfile: Dockerfile
        ports:
            - "5173:80"
        depends_on:
            - server

volumes:
    pgdata:
```

### Server Dockerfile

```dockerfile
# server/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
EXPOSE 8080
CMD ["/server"]
```

### Web App Dockerfile

```dockerfile
# webapp/Dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### Nginx Config (for Vue SPA routing)

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://server:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 13. Environment Variables

### Server

| Variable               | Description                   | Example                             |
| ---------------------- | ----------------------------- | ----------------------------------- |
| `DATABASE_URL`         | PostgreSQL connection string  | `postgres://user:pass@host:5432/db` |
| `REDIS_URL`            | Redis connection string       | `redis://localhost:6379`            |
| `JWT_PRIVATE_KEY_FILE` | Path to RS256 private key PEM | `/secrets/private.pem`              |
| `JWT_PUBLIC_KEY_FILE`  | Path to RS256 public key PEM  | `/secrets/public.pem`               |
| `JWT_ACCESS_EXPIRY`    | Access token TTL (seconds)    | `3600`                              |
| `JWT_REFRESH_EXPIRY`   | Refresh token TTL (seconds)   | `2592000`                           |
| `SERVER_ADDR`          | Bind address                  | `:8080`                             |
| `ENV`                  | Environment                   | `development` / `production`        |
| `RATE_LIMIT_RPM`       | Requests per minute per IP    | `60`                                |

### Web App (Build-time via Vite)

| Variable            | Description           | Example                      |
| ------------------- | --------------------- | ---------------------------- |
| `VITE_API_BASE_URL` | API base URL          | `https://api.timetracker.io` |
| `VITE_WS_URL`       | WebSocket URL         | `wss://api.timetracker.io`   |
| `VITE_APP_VERSION`  | Displayed in settings | `1.0.0`                      |

### Android Client

| Variable       | Description            |
| -------------- | ---------------------- |
| `WEBAPP_URL`   | URL loaded in WebView  |
| `API_BASE_URL` | Collector API endpoint |

---

## 14. Development Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker + Docker Compose
- Android Studio (for Android client)
- Go + WebView2 SDK (for Windows client)

### 1. Clone and start infrastructure

```bash
git clone https://github.com/yourorg/time-tracker
cd time-tracker

# Generate JWT keys
mkdir -p docker/secrets
openssl genrsa -out docker/secrets/private.pem 2048
openssl rsa -in docker/secrets/private.pem -pubout -out docker/secrets/public.pem

# Start Postgres and Redis
cd docker && docker compose up postgres redis -d
```

### 2. Run the Go server

```bash
cd server
cp .env.example .env   # fill in values
go mod download

# Run migrations
go run ./cmd/migrate

# Start server
go run ./cmd/server
# Server available at http://localhost:8080
```

### 3. Run the Vue web app

```bash
cd webapp
npm install
cp .env.example .env.local  # set VITE_API_BASE_URL=http://localhost:8080
npm run dev
# App available at http://localhost:5173
```

### 4. Run the browser extension (dev mode)

```bash
cd clients/extension
npm install
npm run dev   # Outputs to dist/

# Chrome: go to chrome://extensions → Load Unpacked → select dist/
```

### 5. Run Windows collector (optional)

```bash
cd clients/windows
go run ./cmd/main.go
```

### Useful Scripts

```bash
# Run all server tests
cd server && go test ./...

# Run Vue unit tests
cd webapp && npm run test:unit

# Lint Vue app
cd webapp && npm run lint

# Build production Vue bundle
cd webapp && npm run build

# Generate Go API mocks
cd server && go generate ./...
```

---

_This document covers the full technical implementation of the Cross-Device Time Tracker. As the system evolves, update the API Reference and Data Flow sections first — they are the primary contract between all components._
