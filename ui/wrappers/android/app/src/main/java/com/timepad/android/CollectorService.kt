package com.timepad.android

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.app.usage.UsageStatsManager
import android.content.Context
import android.content.Intent
import android.os.IBinder
import android.util.Log
import java.time.Instant
import java.time.format.DateTimeFormatter
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

private const val TAG = "CollectorService"
private const val CHANNEL_ID = "timepad_collector"
private const val NOTIFICATION_ID = 1
private const val POLL_MS = 30_000L
private const val FLUSH_MS = 60_000L
private const val MAX_BUFFER = 10

class CollectorService : Service() {

    private lateinit var cfg: Config
    private lateinit var apiClient: ApiClient
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val buffer = mutableListOf<EventInput>()
    private var currentApp: String? = null
    private var currentStart: Instant = Instant.now()

    override fun onCreate() {
        super.onCreate()
        cfg = Config(this)
        apiClient = ApiClient(cfg)
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, buildNotification())
        Log.d(TAG, "collector service created")
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        scope.launch { runCollector() }
        return START_STICKY
    }

    private suspend fun runCollector() {
        Log.d(TAG, "collector started — server=${cfg.serverURL}")
        var lastFlush = System.currentTimeMillis()

        while (currentCoroutineContext().isActive) {
            delay(POLL_MS)
            poll()
            if (System.currentTimeMillis() - lastFlush >= FLUSH_MS) {
                flush()
                lastFlush = System.currentTimeMillis()
            }
        }
    }

    private fun poll() {
        val now = Instant.now()
        val fg = getForegroundApp() ?: return

        if (fg == currentApp) return

        // App changed — close previous session
        currentApp?.let { prev ->
            Log.d(TAG, "app change: $prev -> $fg")
            buffer.add(
                    EventInput(
                            appName = prev,
                            windowTitle = prev,
                            startTime = fmt(currentStart),
                            endTime = fmt(now),
                            isIdle = false,
                    )
            )
        }

        currentApp = fg
        currentStart = now

        if (buffer.size >= MAX_BUFFER) {
            Log.d(TAG, "buffer full (${buffer.size}), early flush")
            flush()
        }
    }

    private fun flush() {
        // Close the ongoing session snapshot before sending
        currentApp?.let { app ->
            val now = Instant.now()
            buffer.add(
                    EventInput(
                            appName = app,
                            windowTitle = app,
                            startTime = fmt(currentStart),
                            endTime = fmt(now),
                            isIdle = false,
                    )
            )
            currentStart = now
        }

        if (buffer.isEmpty()) {
            Log.d(TAG, "flush: buffer empty")
            return
        }

        if (cfg.deviceKey.isEmpty() || cfg.accessToken.isEmpty()) {
            Log.w(
                    TAG,
                    "not authenticated (device_key=${cfg.deviceKey.isNotEmpty()}" +
                            " access_token=${cfg.accessToken.isNotEmpty()}), dropping ${buffer.size} events"
            )
            buffer.clear()
            return
        }

        val batch = buffer.toList()
        buffer.clear()
        Log.d(TAG, "sending batch of ${batch.size} events")

        scope.launch {
            apiClient.postEvents(batch)
            Log.d(TAG, "sent ${batch.size} events OK")
        }
    }

    /**
     * Returns the package name of the currently visible app using UsageStatsManager. Requires
     * PACKAGE_USAGE_STATS permission (user must grant via Settings).
     */
    private fun getForegroundApp(): String? {
        val usm = getSystemService(Context.USAGE_STATS_SERVICE) as UsageStatsManager
        val now = System.currentTimeMillis()
        val stats =
                usm.queryUsageStats(
                        UsageStatsManager.INTERVAL_DAILY,
                        now - 10_000L,
                        now,
                )
        if (stats.isNullOrEmpty()) return null
        return stats.maxByOrNull { it.lastTimeUsed }?.packageName
    }

    private fun fmt(instant: Instant): String = DateTimeFormatter.ISO_INSTANT.format(instant)

    private fun buildNotification(): Notification =
            Notification.Builder(this, CHANNEL_ID)
                    .setContentTitle("Timepad")
                    .setContentText("Tracking app usage")
                    .setSmallIcon(android.R.drawable.ic_menu_recent_history)
                    .setOngoing(true)
                    .build()

    private fun createNotificationChannel() {
        val ch =
                NotificationChannel(
                                CHANNEL_ID,
                                "Timepad Collector",
                                NotificationManager.IMPORTANCE_LOW,
                        )
                        .apply { description = "Background app-usage tracking" }
        getSystemService(NotificationManager::class.java).createNotificationChannel(ch)
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
        Log.d(TAG, "collector service destroyed")
    }
}
