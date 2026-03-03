package com.timepad.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/**
 * Restarts the collector service after a device reboot.
 * Requires RECEIVE_BOOT_COMPLETED permission.
 */
class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            context.startForegroundService(Intent(context, CollectorService::class.java))
        }
    }
}
