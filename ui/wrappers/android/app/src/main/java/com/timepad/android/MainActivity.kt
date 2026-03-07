package com.timepad.android

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.provider.Settings
import android.view.MotionEvent
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity

private const val CTX = "MainActivity"
private const val TAP_TARGET = 5
private const val TAP_WINDOW_MS = 2_000L

class MainActivity : AppCompatActivity() {

    private lateinit var cfg: Config
    private lateinit var webView: WebView

    private var tapCount = 0
    private var tapFirst = 0L

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        WebView.setWebContentsDebuggingEnabled(BuildConfig.DEBUG)
        cfg = Config(this)

        webView = WebView(this).also { setContentView(it) }
        setupWebView()

        showUrlDialog(
                onCancel = {
                    TPLog.d(CTX, "URL dialog cancelled — loading with current config")
                    loadDashboard()
                }
        )

        checkUsagePermissionAndStartCollector()
    }

    @SuppressLint("SetJavaScriptEnabled", "ClickableViewAccessibility")
    private fun setupWebView() {
        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            databaseEnabled = true
            allowFileAccess = true
        }

        val bridge =
                Bridge(cfg) { newKey ->
                    webView.post {
                        webView.evaluateJavascript(
                                "window.TimePadBridge && " +
                                        "(window.TimePadBridge.getDeviceKey = function() { return ${quote(newKey)}; });",
                                null,
                        )
                    }
                }

        webView.addJavascriptInterface(bridge, "_TimePadNative")

        webView.webViewClient =
                object : WebViewClient() {
                    override fun onPageFinished(view: WebView, url: String) {
                        TPLog.d(CTX, "page loaded: $url")
                        injectBridge(view)
                    }

                    override fun onReceivedError(
                            view: WebView,
                            request: WebResourceRequest,
                            error: WebResourceError,
                    ) {
                        if (!request.isForMainFrame) return
                        TPLog.e(
                                CTX,
                                "WebView error ${error.errorCode}: ${error.description} — ${request.url}"
                        )
                        runOnUiThread { showUrlDialog(onCancel = null) }
                    }
                }

        webView.setOnTouchListener { _, event ->
            if (event.action == MotionEvent.ACTION_DOWN) {
                val now = System.currentTimeMillis()
                if (now - tapFirst > TAP_WINDOW_MS) {
                    tapCount = 0
                    tapFirst = now
                }
                tapCount++
                if (tapCount >= TAP_TARGET) {
                    tapCount = 0
                    showUrlDialog(onCancel = null)
                }
            }
            false
        }
    }

    private fun injectBridge(view: WebView) {
        val deviceKey = cfg.deviceKey
        val js =
                """
(function() {
  window.TimePadBridge = {
    getDeviceKey: function() { return ${quote(deviceKey)}; },
    getPlatform:  function() { return "android"; }
  };
  window.timePadSaveConfig = function(a, r, d) {
    _TimePadNative.saveConfig(a, r, d);
  };
  try {
    var raw = localStorage.getItem('auth-store');
    if (!raw) return;
    var s = JSON.parse(raw).state;
    if (s && s.accessToken && s.refreshToken) {
      _TimePadNative.saveConfig(s.accessToken, s.refreshToken, '');
    }
  } catch(_) {}
})();
        """.trimIndent()
        view.evaluateJavascript(js, null)
    }

    private fun loadDashboard() {
        TPLog.d(CTX, "loading dashboard=${cfg.dashboardURL}")
        webView.loadUrl(cfg.dashboardURL)
    }

    private fun showUrlDialog(onCancel: (() -> Unit)?) {
        TPLog.d(CTX, "showing URL dialog")
        val dp = resources.displayMetrics.density
        val pad = (20 * dp).toInt()
        val gap = (8 * dp).toInt()

        val layout =
                LinearLayout(this).apply {
                    orientation = LinearLayout.VERTICAL
                    setPadding(pad, pad, pad, 0)
                }

        fun label(text: String) =
                TextView(this).apply {
                    this.text = text
                    setPadding(0, gap, 0, 2)
                }

        val serverField =
                EditText(this).apply {
                    hint = BuildConfig.SERVER_URL
                    setText(cfg.serverURL)
                    inputType = android.text.InputType.TYPE_TEXT_VARIATION_URI
                }
        val dashField =
                EditText(this).apply {
                    hint = BuildConfig.DASHBOARD_URL
                    setText(cfg.dashboardURL)
                    inputType = android.text.InputType.TYPE_TEXT_VARIATION_URI
                }

        layout.addView(label("Server URL"))
        layout.addView(serverField)
        layout.addView(label("Dashboard URL"))
        layout.addView(dashField)

        val btnLogs =
                Button(this).apply {
                    text = "View Logs"
                    setOnClickListener { showLogDialog() }
                }
        layout.addView(btnLogs)
        val builder =
                AlertDialog.Builder(this)
                        .setTitle("Connection")
                        .setView(layout)
                        .setPositiveButton("Connect") { _, _ ->
                            val newServer = serverField.text.toString().trimEnd('/')
                            val newDash = dashField.text.toString().trimEnd('/')
                            cfg.setServerURL(newServer)
                            cfg.setDashboardURL(newDash)
                            TPLog.d(CTX, "URLs saved — server=$newServer dashboard=$newDash")
                            loadDashboard()
                        }
                        .setNeutralButton("Reset") { _, _ ->
                            cfg.resetURLs()
                            TPLog.d(CTX, "URLs reset to BuildConfig defaults")
                            loadDashboard()
                        }

        if (onCancel != null) {
            builder.setNegativeButton("Cancel") { _, _ -> onCancel() }
        }

        builder.show()
    }

    private fun showLogDialog() {
        val lines = TPLog.lines()
        val text = if (lines.isEmpty()) "no logs yet" else lines.joinToString("\n")

        val dp = resources.displayMetrics.density
        val pad = (12 * dp).toInt()

        val tv =
                TextView(this).apply {
                    this.text = text
                    setTextIsSelectable(true)
                    typeface = android.graphics.Typeface.MONOSPACE
                    textSize = 11f
                    setPadding(pad, pad, pad, pad)
                }

        val scroll = ScrollView(this).apply { addView(tv) }

        val dialog =
                AlertDialog.Builder(this)
                        .setTitle("Logs (last ${lines.size})")
                        .setView(scroll)
                        .setPositiveButton("Close", null)
                        .setNeutralButton("Copy") { _, _ ->
                            val cm =
                                    getSystemService(CLIPBOARD_SERVICE) as
                                            android.content.ClipboardManager
                            cm.setPrimaryClip(
                                    android.content.ClipData.newPlainText("timepad logs", text)
                            )
                        }
                        .create()

        dialog.show()
        // scroll to bottom after layout
        scroll.post { scroll.fullScroll(ScrollView.FOCUS_DOWN) }
    }

    private fun checkUsagePermissionAndStartCollector() {
        if (!hasUsagePermission()) {
            AlertDialog.Builder(this)
                    .setTitle("Permission required")
                    .setMessage(
                            "Timepad needs \"Usage access\" permission to track which apps you use.\n\n" +
                                    "Please grant it in the next screen."
                    )
                    .setPositiveButton("Open Settings") { _, _ ->
                        startActivity(Intent(Settings.ACTION_USAGE_ACCESS_SETTINGS))
                    }
                    .setNegativeButton("Skip", null)
                    .show()
        } else {
            startCollector()
        }
    }

    private fun hasUsagePermission(): Boolean {
        val usm = getSystemService(USAGE_STATS_SERVICE) as android.app.usage.UsageStatsManager
        val now = System.currentTimeMillis()
        val stats =
                usm.queryUsageStats(
                        android.app.usage.UsageStatsManager.INTERVAL_DAILY,
                        now - 60_000,
                        now,
                )
        return !stats.isNullOrEmpty()
    }

    private fun startCollector() {
        TPLog.d(CTX, "starting collector service")
        startForegroundService(Intent(this, CollectorService::class.java))
    }

    @Deprecated("Deprecated in Java")
    override fun onBackPressed() {
        if (webView.canGoBack()) webView.goBack() else super.onBackPressed()
    }

    private fun quote(s: String): String = "\"${s.replace("\\", "\\\\").replace("\"", "\\\"")}\""
}
