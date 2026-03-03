package com.timepad.android

import android.util.Log
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import java.util.concurrent.TimeUnit

private const val TAG = "ApiClient"

data class EventInput(
    val appName: String,
    val windowTitle: String,
    val startTime: String,   // ISO-8601
    val endTime: String,
    val isIdle: Boolean,
)

class ApiClient(private val cfg: Config) {

    private val http = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .build()

    private val jsonType = "application/json; charset=utf-8".toMediaType()

    fun postEvents(events: List<EventInput>) {
        if (cfg.deviceKey.isEmpty()) {
            Log.w(TAG, "no device_key — register device in the dashboard")
            return
        }
        if (cfg.accessToken.isEmpty()) {
            Log.w(TAG, "not authenticated")
            return
        }

        Log.d(TAG, "posting ${events.size} event(s) to ${cfg.serverURL}/events")
        val body = buildPayload(events)

        var status = doPost("/events", body)
        Log.d(TAG, "POST /events -> HTTP $status")

        if (status == 401) {
            Log.d(TAG, "token expired, refreshing")
            refreshToken()
            status = doPost("/events", buildPayload(events))
            Log.d(TAG, "retry POST /events -> HTTP $status")
        }

        if (status >= 400) {
            Log.e(TAG, "POST /events HTTP $status")
        }
    }

    private fun buildPayload(events: List<EventInput>): String {
        val arr = JSONArray()
        events.forEach { e ->
            arr.put(
                JSONObject().apply {
                    put("app_name", e.appName)
                    put("window_title", e.windowTitle)
                    put("url", "")
                    put("start_time", e.startTime)
                    put("end_time", e.endTime)
                    put("is_idle", e.isIdle)
                }
            )
        }
        return JSONObject().apply {
            put("device_key", cfg.deviceKey)
            put("events", arr)
        }.toString()
    }

    private fun doPost(path: String, body: String): Int {
        val req = Request.Builder()
            .url(cfg.serverURL + path)
            .post(body.toRequestBody(jsonType))
            .header("Authorization", "Bearer ${cfg.accessToken}")
            .build()
        return try {
            http.newCall(req).execute().use { it.code }
        } catch (e: Exception) {
            Log.e(TAG, "doPost $path: ${e.message}")
            0
        }
    }

    private fun refreshToken() {
        val rt = cfg.refreshToken
        if (rt.isEmpty()) {
            Log.w(TAG, "no refresh token")
            return
        }
        val body = JSONObject().put("refresh_token", rt).toString()
        val req = Request.Builder()
            .url(cfg.serverURL + "/auth/refresh")
            .post(body.toRequestBody(jsonType))
            .build()
        try {
            http.newCall(req).execute().use { resp ->
                if (resp.code != 200) {
                    Log.e(TAG, "refresh HTTP ${resp.code}")
                    return
                }
                val obj = JSONObject(resp.body!!.string())
                cfg.setTokens(
                    obj.getString("access_token"),
                    obj.getString("refresh_token"),
                )
                Log.d(TAG, "tokens refreshed and saved")
            }
        } catch (e: Exception) {
            Log.e(TAG, "refreshToken: ${e.message}")
        }
    }
}
