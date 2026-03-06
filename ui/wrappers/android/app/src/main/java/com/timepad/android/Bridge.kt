package com.timepad.android

import android.webkit.JavascriptInterface

private const val CTX = "Bridge"

/**
 * JavaScript bridge exposed as `_TimePadNative` in the WebView. The injected JS wraps it into the
 * `window.TimePadBridge` / `window.timePadSaveConfig` contract expected by use-native-bridge.ts.
 */
class Bridge(
        private val cfg: Config,
        private val onDeviceKeyUpdated: (String) -> Unit,
) {
    @JavascriptInterface fun getDeviceKey(): String = cfg.deviceKey

    @JavascriptInterface fun getPlatform(): String = "android"

    @JavascriptInterface
    fun saveConfig(accessToken: String, refreshToken: String, deviceKey: String) {
        TPLog.d(
                CTX,
                "saveConfig device_key=${deviceKey.ifEmpty { "(empty)" }} tokens=${accessToken.isNotEmpty()}"
        )
        cfg.setTokens(accessToken, refreshToken)
        if (deviceKey.isNotEmpty()) {
            cfg.setDeviceKey(deviceKey)
            onDeviceKeyUpdated(deviceKey)
        }
        TPLog.d(CTX, "config saved")
    }
}
