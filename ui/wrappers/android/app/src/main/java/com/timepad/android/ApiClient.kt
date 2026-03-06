package com.timepad.android

import java.util.concurrent.TimeUnit
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject

private const val CTX = "ApiClient"

data class EventInput(
        val appName: String,
        val windowTitle: String,
        val startTime: String, // ISO-8601
        val endTime: String,
        val isIdle: Boolean,
)

class ApiClient(private val cfg: Config) {

    private val http =
            OkHttpClient.Builder()
                    .connectTimeout(15, TimeUnit.SECONDS)
                    .readTimeout(15, TimeUnit.SECONDS)
                    .build()

    private val jsonType = "application/json; charset=utf-8".toMediaType()

    fun postEvents(events: List<EventInput>) {
        if (cfg.deviceKey.isEmpty()) {
            TPLog.w(CTX, "no device_key — register device in the dashboard")
            return
        }
        if (cfg.accessToken.isEmpty()) {
            TPLog.w(CTX, "not authenticated — dropping ${events.size} event(s)")
            return
        }

        val body = buildPayload(events)
        TPLog.d(CTX, "POST ${cfg.serverURL}/events — ${events.size} event(s), ${body.length}B")

        var status = doPost("/events", body)

        if (status == 401) {
            TPLog.d(CTX, "token expired — refreshing and retrying")
            refreshToken()
            status = doPost("/events", buildPayload(events))
        }

        if (status in 200..299) {
            TPLog.d(CTX, "POST /events -> $status OK")
        } else {
            TPLog.e(CTX, "POST /events -> $status")
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
        return JSONObject()
                .apply {
                    put("device_key", cfg.deviceKey)
                    put("events", arr)
                }
                .toString()
    }

    private fun doPost(path: String, body: String): Int {
        val url = cfg.serverURL + path
        val req =
                Request.Builder()
                        .url(url)
                        .post(body.toRequestBody(jsonType))
                        .header("Authorization", "Bearer ${cfg.accessToken}")
                        .build()
        val t0 = System.currentTimeMillis()
        return try {
            http.newCall(req).execute().use { resp ->
                val ms = System.currentTimeMillis() - t0
                TPLog.d(CTX, "POST $url -> ${resp.code} (${ms}ms)")
                resp.code
            }
        } catch (e: Exception) {
            TPLog.e(
                    CTX,
                    "POST $url failed after ${System.currentTimeMillis() - t0}ms: ${e.message}"
            )
            0
        }
    }

    private fun refreshToken() {
        val rt = cfg.refreshToken
        if (rt.isEmpty()) {
            TPLog.w(CTX, "no refresh token")
            return
        }
        val url = cfg.serverURL + "/auth/refresh"
        val body = JSONObject().put("refresh_token", rt).toString()
        val req = Request.Builder().url(url).post(body.toRequestBody(jsonType)).build()
        val t0 = System.currentTimeMillis()
        try {
            http.newCall(req).execute().use { resp ->
                val ms = System.currentTimeMillis() - t0
                if (resp.code != 200) {
                    TPLog.e(CTX, "POST $url -> ${resp.code} (${ms}ms)")
                    return
                }
                val obj = JSONObject(resp.body!!.string())
                cfg.setTokens(
                        obj.getString("access_token"),
                        obj.getString("refresh_token"),
                )
                TPLog.d(CTX, "POST $url -> ${resp.code} (${ms}ms) — tokens saved")
            }
        } catch (e: Exception) {
            TPLog.e(CTX, "POST $url failed: ${e.message}")
        }
    }
}
