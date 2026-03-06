package com.timepad.android

import android.util.Log

/** Single logcat tag "TIMEPAD" for all app logs. Filter: adb logcat TIMEPAD:D *:S */
internal object TPLog {
    const val TAG = "TIMEPAD"

    fun d(ctx: String, msg: String) = Log.d(TAG, "[$ctx] $msg")
    fun w(ctx: String, msg: String) = Log.w(TAG, "[$ctx] $msg")
    fun e(ctx: String, msg: String) = Log.e(TAG, "[$ctx] $msg")
    fun e(ctx: String, msg: String, t: Throwable) = Log.e(TAG, "[$ctx] $msg", t)
}
