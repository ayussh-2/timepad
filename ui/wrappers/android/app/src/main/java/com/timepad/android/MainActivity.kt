package com.timepad.android

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.provider.Settings
import android.util.Log
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity

private const val TAG = "MainActivity"

class MainActivity : AppCompatActivity() {

    private lateinit var cfg: Config
    private lateinit var webView: WebView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        WebView.setWebContentsDebuggingEnabled(BuildConfig.DEBUG)
        cfg = Config(this)

        webView = WebView(this).also { setContentView(it) }
        setupWebView()

        Log.d(TAG, "navigating to ${cfg.dashboardURL}")
        webView.loadUrl(cfg.dashboardURL)

        checkUsagePermissionAndStartCollector()
    }

    @SuppressLint("SetJavaScriptEnabled")
    private fun setupWebView() {
        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            databaseEnabled = true
            allowFileAccess = true
        }

        val bridge = Bridge(cfg) { newKey ->
            // Keep the JS bridge in sync when a device key arrives at runtime
            webView.post {
                webView.evaluateJavascript(
                    "window.TimePadBridge && " +
                        "(window.TimePadBridge.getDeviceKey = function() { return ${quote(newKey)}; });",
                    null,
                )
            }
        }

        // Expose the native object; the page-finished handler wraps it into
        // the TimePadBridge / timePadSaveConfig contract.
        webView.addJavascriptInterface(bridge, "_TimePadNative")

        webView.webViewClient = object : WebViewClient() {
            override fun onPageFinished(view: WebView, url: String) {
                injectBridge(view)
            }
        }
    }

    /**
     * Injects window.TimePadBridge and window.timePadSaveConfig that match the
     * contract in use-native-bridge.ts, delegating to the @JavascriptInterface.
     */
    private fun injectBridge(view: WebView) {
        val deviceKey = cfg.deviceKey
        val js = """
(function() {
  window.TimePadBridge = {
    getDeviceKey: function() { return ${quote(deviceKey)}; },
    getPlatform:  function() { return "android"; }
  };
  window.timePadSaveConfig = function(a, r, d) {
    _TimePadNative.saveConfig(a, r, d);
  };
  // Restore tokens already persisted in localStorage
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
        val stats = usm.queryUsageStats(
            android.app.usage.UsageStatsManager.INTERVAL_DAILY,
            now - 60_000,
            now,
        )
        return !stats.isNullOrEmpty()
    }

    private fun startCollector() {
        startForegroundService(Intent(this, CollectorService::class.java))
    }

    @Deprecated("Deprecated in Java")
    override fun onBackPressed() {
        if (webView.canGoBack()) webView.goBack() else super.onBackPressed()
    }

    /** Wraps a string in a JS string literal. */
    private fun quote(s: String): String = "\"${s.replace("\\", "\\\\").replace("\"", "\\\"")}\""
}
