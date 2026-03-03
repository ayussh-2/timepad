import { Eye, EyeOff } from "lucide-react";
import { useEffect, useState } from "react";
import { devicesApi } from "~/app/api/devices";
import { DateNavigator } from "~/components/ui/date-navigator";
import { EmptyState } from "~/components/ui/empty-state";
import { EventDetailDrawer } from "~/components/timeline/event-detail-drawer";
import { TimelineCanvas } from "~/components/timeline/timeline-canvas";
import { useTimeline } from "~/hooks/use-timeline";
import { useActivityStore } from "~/store/activity.store";
import type { TimelineEntry } from "~/app/types";
import { Button } from "~/components/ui/button";
import { Skeleton } from "~/components/ui/skeleton";
import { Toggle } from "~/components/ui/toggle";
import { isSystemApp } from "~/utils/app-icon";

export default function TimelinePage() {
    const selectedDate = useActivityStore((s) => s.selectedDate);
    const setSelectedDate = useActivityStore((s) => s.setSelectedDate);
    const { timeline, hasMore, isLoading, loadMore } =
        useTimeline(selectedDate);

    const [activeEvent, setActiveEvent] = useState<TimelineEntry | null>(null);
    const [filteredDeviceIds, setFilteredDeviceIds] = useState<Set<string>>(
        new Set(),
    );
    const [deviceNames, setDeviceNames] = useState<Record<string, string>>({});
    const [hideSystem, setHideSystem] = useState(true);

    useEffect(() => {
        devicesApi
            .list()
            .then((devs) => {
                const map: Record<string, string> = {};
                devs.forEach((d) => {
                    map[d.id] = d.name;
                });
                setDeviceNames(map);
            })
            .catch(() => {});
    }, []);

    const uniqueDeviceIds = Array.from(
        new Set(timeline.map((e) => e.device_id)),
    );

    const toggleDevice = (id: string) => {
        setFilteredDeviceIds((prev) => {
            const next = new Set(prev);
            next.has(id) ? next.delete(id) : next.add(id);
            return next;
        });
    };

    const visibleTimeline = hideSystem
        ? timeline.filter((e) => !isSystemApp(e.app_name) && !e.app?.is_system)
        : timeline;

    return (
        <div className="max-w-5xl mx-auto px-4 py-6 space-y-6">
            <div className="flex flex-wrap items-center gap-3">
                <DateNavigator date={selectedDate} onChange={setSelectedDate} />

                {uniqueDeviceIds.length > 1 && (
                    <div className="flex gap-1.5 flex-wrap">
                        {uniqueDeviceIds.map((id) => (
                            <Toggle
                                key={id}
                                pressed={!filteredDeviceIds.has(id)}
                                onPressedChange={() => toggleDevice(id)}
                                size="sm"
                                variant="outline"
                                className="text-xs h-7"
                            >
                                {deviceNames[id] ?? id.slice(0, 8)}
                            </Toggle>
                        ))}
                    </div>
                )}

                <Toggle
                    pressed={hideSystem}
                    onPressedChange={setHideSystem}
                    size="sm"
                    variant="outline"
                    className="h-7 gap-1.5 text-xs px-2 ml-auto"
                    aria-label="Toggle system apps"
                >
                    {hideSystem ? (
                        <Eye className="h-3 w-3" />
                    ) : (
                        <EyeOff className="h-3 w-3" />
                    )}
                    {hideSystem ? "Show system" : "Hide system"}
                </Toggle>
            </div>

            {isLoading && timeline.length === 0 ? (
                <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                        <Skeleton key={i} className="h-12 w-full rounded-lg" />
                    ))}
                </div>
            ) : timeline.length === 0 ? (
                <EmptyState
                    title="Nothing here yet"
                    description="No activity recorded for this day."
                />
            ) : (
                <>
                    <TimelineCanvas
                        events={visibleTimeline}
                        deviceIds={filteredDeviceIds}
                        onEventClick={setActiveEvent}
                        date={selectedDate}
                        deviceNames={deviceNames}
                    />
                    {hasMore && (
                        <div className="flex justify-center">
                            <Button
                                variant="outline"
                                onClick={loadMore}
                                disabled={isLoading}
                            >
                                {isLoading ? "Loading..." : "Load more"}
                            </Button>
                        </div>
                    )}
                </>
            )}

            <EventDetailDrawer
                event={activeEvent}
                onClose={() => setActiveEvent(null)}
            />
        </div>
    );
}
