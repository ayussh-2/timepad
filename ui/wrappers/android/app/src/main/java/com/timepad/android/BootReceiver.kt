package com.timepad.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

private const val CTX = "BootReceiver"

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            TPLog.d(CTX, "boot completed — starting collector")
            context.startForegroundService(Intent(context, CollectorService::class.java))
        }
    }
}
