package com.timepad.android

import android.util.Log
import android.webkit.JavascriptInterface

private const val TAG = "Bridge"

/**
 * JavaScript bridge exposed as `_TimePadNative` in the WebView.
 * The injected JS wraps it into the `window.TimePadBridge` / `window.timePadSaveConfig`
 * contract expected by use-native-bridge.ts.
 */
class Bridge(
    private val cfg: Config,
    private val onDeviceKeyUpdated: (String) -> Unit,
) {
    @JavascriptInterface
    fun getDeviceKey(): String = cfg.deviceKey

    @JavascriptInterface
    fun getPlatform(): String = "android"

    @JavascriptInterface
    fun saveConfig(accessToken: String, refreshToken: String, deviceKey: String) {
        Log.d(
            TAG,
            "saveConfig called device_key=$deviceKey tokens_present=${accessToken.isNotEmpty()}"
        )
        cfg.setTokens(accessToken, refreshToken)
        if (deviceKey.isNotEmpty()) {
            cfg.setDeviceKey(deviceKey)
            onDeviceKeyUpdated(deviceKey)
        }
        Log.d(TAG, "config saved")
    }
}
