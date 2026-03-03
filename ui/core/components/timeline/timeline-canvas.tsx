import dayjs from "dayjs";
import { AppIcon } from "~/components/ui/app-icon";
import { formatDuration } from "~/components/ui/duration";
import type { TimelineEntry } from "~/app/types";
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "~/components/ui/tooltip";

const DAY_START_HOUR = 0;
const DAY_END_HOUR = 24;
const TOTAL_MINS = (DAY_END_HOUR - DAY_START_HOUR) * 60;

/** Palette used when an app has no category colour assigned. */
const TRACK_PALETTE = [
    "#4f8ef7",
    "#e06c75",
    "#98c379",
    "#e5c07b",
    "#c678dd",
    "#56b6c2",
    "#d19a66",
    "#61afef",
    "#f08080",
    "#7ec8a4",
    "#b48ead",
    "#a3be8c",
];

/** Stable colour derived from an app name when no category colour exists. */
function paletteColor(name: string): string {
    let h = 0;
    for (let i = 0; i < name.length; i++) {
        h = (h * 31 + name.charCodeAt(i)) >>> 0;
    }
    return TRACK_PALETTE[h % TRACK_PALETTE.length];
}

function appColor(event: TimelineEntry): string {
    return event.app?.category?.color ?? paletteColor(event.app_name);
}

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
        <div className="space-y-6">
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

            {/* Device sections */}
            {deviceList.length === 0 ? (
                <div className="relative h-10 bg-surface-alt rounded-lg" />
            ) : (
                deviceList.map((deviceId) => {
                    const deviceEvents = events.filter(
                        (e) => e.device_id === deviceId && !e.is_idle,
                    );
                    const label = deviceNames[deviceId] ?? deviceId.slice(0, 8);

                    // Stable ordered list of unique app names (order of first appearance)
                    const appNames: string[] = [];
                    for (const e of deviceEvents) {
                        if (!appNames.includes(e.app_name)) {
                            appNames.push(e.app_name);
                        }
                    }

                    return (
                        <div key={deviceId} className="space-y-1.5">
                            {deviceList.length > 1 && (
                                <p className="text-xs font-medium text-secondary-text mb-2">
                                    {label}
                                </p>
                            )}

                            {appNames.map((appName) => {
                                const trackEvents = deviceEvents.filter(
                                    (e) => e.app_name === appName,
                                );
                                // All events for the same app share the same color
                                const color = appColor(trackEvents[0]);
                                // Representative event for icon
                                const rep = trackEvents[0];

                                return (
                                    <div
                                        key={appName}
                                        className="flex items-center gap-2"
                                    >
                                        {/* App label */}
                                        <div className="flex items-center gap-1.5 w-32 shrink-0">
                                            <AppIcon
                                                appName={rep.app_name}
                                                url={rep.url}
                                                size="sm"
                                            />
                                            <span
                                                className="text-xs truncate text-secondary-text"
                                                title={appName}
                                            >
                                                {appName}
                                            </span>
                                        </div>

                                        {/* Track */}
                                        <div className="relative flex-1 h-7 bg-surface-alt rounded-md overflow-hidden">
                                            {trackEvents.map((evt) => (
                                                <EventBar
                                                    key={evt.id}
                                                    event={evt}
                                                    color={color}
                                                    onClick={() =>
                                                        onEventClick(evt)
                                                    }
                                                />
                                            ))}
                                            {nowPct !== null && (
                                                <div
                                                    className="absolute top-0 bottom-0 w-0.5 bg-accent z-10"
                                                    style={{
                                                        left: `${nowPct}%`,
                                                    }}
                                                />
                                            )}
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    );
                })
            )}
        </div>
    );
}

function EventBar({
    event,
    color,
    onClick,
}: {
    event: TimelineEntry;
    color: string;
    onClick: () => void;
}) {
    const left = positionPercent(event.start_time);
    const width = widthPercent(event.start_time, event.end_time);

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    onClick={onClick}
                    className="absolute top-1 bottom-1 rounded transition-opacity hover:opacity-75 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
                    style={{
                        left: `${left}%`,
                        width: `${width}%`,
                        minWidth: 4,
                        background: color + "55",
                        borderLeft: `2px solid ${color}`,
                    }}
                />
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-52">
                <div className="flex items-center gap-2 mb-1">
                    <AppIcon
                        appName={event.app_name}
                        url={event.url}
                        size="sm"
                    />
                    <p className="font-medium">{event.app_name}</p>
                </div>
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
