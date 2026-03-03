package com.timepad.android

import android.content.Context
import android.content.SharedPreferences

/**
 * Persistent config backed by SharedPreferences.
 * Default URLs come from BuildConfig (set via local.properties at build time).
 */
class Config(context: Context) {

    private val prefs: SharedPreferences =
        context.getSharedPreferences("timepad", Context.MODE_PRIVATE)

    val serverURL: String
        get() = prefs.getString("server_url", BuildConfig.SERVER_URL)
            ?.takeIf { it.isNotBlank() } ?: BuildConfig.SERVER_URL

    val dashboardURL: String
        get() = prefs.getString("dashboard_url", BuildConfig.DASHBOARD_URL)
            ?.takeIf { it.isNotBlank() } ?: BuildConfig.DASHBOARD_URL

    val accessToken: String
        get() = prefs.getString("access_token", "") ?: ""

    val refreshToken: String
        get() = prefs.getString("refresh_token", "") ?: ""

    val deviceKey: String
        get() = prefs.getString("device_key", "") ?: ""

    fun setTokens(access: String, refresh: String) {
        prefs.edit()
            .putString("access_token", access)
            .putString("refresh_token", refresh)
            .apply()
    }

    fun setDeviceKey(key: String) {
        prefs.edit().putString("device_key", key).apply()
    }
}
