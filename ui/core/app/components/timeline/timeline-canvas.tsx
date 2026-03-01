import dayjs from "dayjs";
import { useState } from "react";
import { formatDuration } from "~/app/components/ui/duration";
import type { TimelineEntry } from "~/app/types";
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "~/components/ui/tooltip";

const DAY_START_HOUR = 0;
const DAY_END_HOUR = 24;
const TOTAL_MINS = (DAY_END_HOUR - DAY_START_HOUR) * 60;

function positionPercent(time: string): number {
    const d = dayjs(time);
    const mins = d.hour() * 60 + d.minute();
    return Math.max(0, Math.min(100, (mins / TOTAL_MINS) * 100));
}

function widthPercent(start: string, end: string): number {
    const s = dayjs(start);
    const e = dayjs(end);
    const durMins = e.diff(s, "minute");
    return Math.max(0.3, Math.min(100, (durMins / TOTAL_MINS) * 100));
}

interface TimelineCanvasProps {
    events: TimelineEntry[];
    deviceIds: Set<string>;
    onEventClick: (event: TimelineEntry) => void;
    date: string;
    deviceNames?: Record<string, string>;
}

export function TimelineCanvas({
    events,
    deviceIds,
    onEventClick,
    date,
    deviceNames = {},
}: TimelineCanvasProps) {
    const deviceList = Array.from(
        new Set(events.map((e) => e.device_id)),
    ).filter((id) => deviceIds.size === 0 || deviceIds.has(id));

    const isToday = dayjs(date).isSame(dayjs(), "day");
    const nowPct = isToday ? positionPercent(dayjs().toISOString()) : null;

    const hourTicks = Array.from({ length: 13 }, (_, i) => i * 2); // 0,2,4...24

    return (
        <div className="space-y-4">
            {/* X-axis ticks */}
            <div className="relative h-5 border-b border-divider">
                {hourTicks.map((h) => (
                    <div
                        key={h}
                        className="absolute text-xs text-secondary-text"
                        style={{
                            left: `${((h * 60) / TOTAL_MINS) * 100}%`,
                            transform: "translateX(-50%)",
                        }}
                    >
                        {h === 0
                            ? "12a"
                            : h === 12
                              ? "12p"
                              : h < 12
                                ? `${h}a`
                                : `${h - 12}p`}
                    </div>
                ))}
            </div>

            {/* Device rows */}
            {deviceList.length === 0 ? (
                <div className="relative h-10 bg-surface-alt rounded-lg" />
            ) : (
                deviceList.map((deviceId) => {
                    const row = events.filter(
                        (e) => e.device_id === deviceId && !e.is_idle,
                    );
                    const label = deviceNames[deviceId] ?? deviceId.slice(0, 8);
                    return (
                        <div key={deviceId} className="space-y-1">
                            <p className="text-xs text-secondary-text">
                                {label}
                            </p>
                            <div className="relative h-10 bg-surface-alt rounded-lg overflow-hidden">
                                {row.map((evt) => (
                                    <EventBar
                                        key={evt.id}
                                        event={evt}
                                        onClick={() => onEventClick(evt)}
                                    />
                                ))}
                                {nowPct !== null && (
                                    <div
                                        className="absolute top-0 bottom-0 w-0.5 bg-accent z-10"
                                        style={{ left: `${nowPct}%` }}
                                    />
                                )}
                            </div>
                        </div>
                    );
                })
            )}
        </div>
    );
}

function EventBar({
    event,
    onClick,
}: {
    event: TimelineEntry;
    onClick: () => void;
}) {
    const left = positionPercent(event.start_time);
    const width = widthPercent(event.start_time, event.end_time);
    const color = event.category?.color ?? "#5b7c99";

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    onClick={onClick}
                    className="absolute top-1 bottom-1 rounded transition-opacity hover:opacity-80 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
                    style={{
                        left: `${left}%`,
                        width: `${width}%`,
                        minWidth: 4,
                        background: color + "66",
                        borderLeft: `2px solid ${color}`,
                    }}
                />
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-52">
                <p className="font-medium">{event.app_name}</p>
                {event.window_title && (
                    <p className="text-xs text-secondary-text truncate">
                        {event.window_title}
                    </p>
                )}
                <p className="text-xs mt-0.5">
                    {dayjs(event.start_time).format("h:mm")} –{" "}
                    {dayjs(event.end_time).format("h:mm A")} ·{" "}
                    {formatDuration(event.duration_secs)}
                </p>
            </TooltipContent>
        </Tooltip>
    );
}
