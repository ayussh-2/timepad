import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { categoriesApi } from "~/app/api/categories";
import { appsApi } from "~/app/api/apps";
import { timelineApi } from "~/app/api/timeline";
import type { AppUsage, Category, TimelineEntry } from "~/app/types";
import { AppIcon } from "~/components/ui/app-icon";
import { formatDuration } from "~/components/ui/duration";
import { PlatformBadge } from "~/components/ui/platform-badge";

import { Separator } from "~/components/ui/separator";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
} from "~/components/ui/sheet";
import { Skeleton } from "~/components/ui/skeleton";
import { cn } from "~/lib/utils";
import { detectPlatform } from "~/utils/app-icon";
import { EventDetailDrawer } from "~/components/timeline/event-detail-drawer";

interface AppTimelineDrawerProps {
    app: AppUsage | null;
    date: string;
    onClose: () => void;
    onCategoryChanged?: (appName: string, category: Category | null) => void;
}

export function AppTimelineDrawer({
    app,
    date,
    onClose,
    onCategoryChanged,
}: AppTimelineDrawerProps) {
    const [events, setEvents] = useState<TimelineEntry[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [activeEvent, setActiveEvent] = useState<TimelineEntry | null>(null);

    const [categories, setCategories] = useState<Category[]>([]);
    const [currentCategory, setCurrentCategory] = useState<Category | null>(
        null,
    );
    const [saving, setSaving] = useState(false);
    const [isSystem, setIsSystem] = useState(false);
    // Sync state when a different app is selected
    useEffect(() => {
        setCurrentCategory(app?.category ?? null);
        setIsSystem(app?.is_system ?? false);
    }, [app?.app_name]);

    useEffect(() => {
        categoriesApi
            .list()
            .then(setCategories)
            .catch(() => {});
    }, []);

    useEffect(() => {
        if (!app) {
            setEvents([]);
            return;
        }
        setIsLoading(true);
        timelineApi
            .get(date, null, 500, app.app_name)
            .then((r) => setEvents(r.events))
            .catch(() => setEvents([]))
            .finally(() => setIsLoading(false));
    }, [app?.app_name, date]);

    const applyCategory = (cat: Category | null) => {
        setCurrentCategory(cat);
        setEvents((prev) =>
            prev.map((e) => ({
                ...e,
                app: e.app
                    ? { ...e.app, category_id: cat?.id ?? null, category: cat }
                    : e.app,
            })),
        );
        onCategoryChanged?.(app!.app_name, cat);
    };

    const handleToggleSystem = async () => {
        if (!app || saving) return;
        setSaving(true);
        try {
            await appsApi.setSystem(app.app_id, !isSystem);
            setIsSystem((v) => !v);
        } catch {
            // silent
        } finally {
            setSaving(false);
        }
    };

    const handleCategoryChange = async (value: string) => {
        if (!app || saving) return;
        setSaving(true);
        try {
            const catId = value === "__none__" ? null : value;
            await appsApi.setCategory(app.app_id, catId);
            const cat = catId
                ? (categories.find((c) => c.id === catId) ?? null)
                : null;
            applyCategory(cat);
        } catch {
            // silent
        } finally {
            setSaving(false);
        }
    };

    const platform = app ? detectPlatform(app.app_name, app.platforms) : null;
    const totalSecs = events
        .filter((e) => !e.is_idle)
        .reduce((s, e) => s + e.duration_secs, 0);

    return (
        <>
            <Sheet open={!!app} onOpenChange={(open) => !open && onClose()}>
                <SheetContent
                    side="right"
                    className="w-full sm:max-w-md flex flex-col gap-0 p-0"
                >
                    <SheetHeader className="px-5 pt-5 pb-4 border-b border-divider space-y-4">
                        <SheetTitle className="flex items-center gap-2.5 text-ink">
                            <div className="relative shrink-0">
                                {app && (
                                    <AppIcon appName={app.app_name} size="md" />
                                )}
                                {platform && (
                                    <span className="absolute -bottom-1 -right-1">
                                        <PlatformBadge platform={platform} />
                                    </span>
                                )}
                            </div>
                            <span className="truncate">{app?.app_name}</span>
                        </SheetTitle>

                        {app && (
                            <div className="flex items-center justify-between gap-4 text-xs text-secondary-text">
                                <div className="flex items-center gap-4">
                                    <span>
                                        Total:{" "}
                                        <span className="font-medium text-ink">
                                            {formatDuration(app.total_secs)}
                                        </span>
                                    </span>
                                    {!isLoading && (
                                        <span>
                                            {
                                                events.filter((e) => !e.is_idle)
                                                    .length
                                            }{" "}
                                            sessions
                                        </span>
                                    )}
                                </div>
                                <button
                                    onClick={handleToggleSystem}
                                    disabled={saving}
                                    title={
                                        isSystem
                                            ? "Unmark as system app"
                                            : "Mark as system app"
                                    }
                                    className={cn(
                                        "flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium border transition-all",
                                        isSystem
                                            ? "border-amber-400/50 bg-amber-400/10 text-amber-600 dark:text-amber-400"
                                            : "border-divider text-secondary-text hover:border-amber-400/40 hover:text-amber-600",
                                    )}
                                >
                                    <span>⚙</span>
                                    {isSystem ? "System" : "Mark system"}
                                </button>
                            </div>
                        )}

                        {app && categories.length > 0 && (
                            <div className="space-y-2">
                                <p className="text-xs font-medium text-secondary-text">
                                    Category
                                </p>
                                <div className="flex flex-wrap gap-1.5">
                                    <button
                                        onClick={() =>
                                            handleCategoryChange("__none__")
                                        }
                                        disabled={saving}
                                        className={cn(
                                            "px-3 py-1.5 rounded-lg text-xs font-medium border transition-all",
                                            currentCategory === null
                                                ? "border-ink/40 bg-ink/10 text-ink"
                                                : "border-divider text-secondary-text hover:text-ink hover:border-ink/30",
                                        )}
                                    >
                                        None
                                    </button>
                                    {categories.map((c) => (
                                        <button
                                            key={c.id}
                                            onClick={() =>
                                                handleCategoryChange(c.id)
                                            }
                                            disabled={saving}
                                            className={cn(
                                                "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium border transition-all",
                                                currentCategory?.id === c.id
                                                    ? "border-ink/40 bg-ink/10 text-ink"
                                                    : "border-divider text-secondary-text hover:text-ink hover:border-ink/30",
                                            )}
                                        >
                                            <span
                                                className="h-2 w-2 rounded-full shrink-0"
                                                style={{ background: c.color }}
                                            />
                                            {c.name}
                                        </button>
                                    ))}
                                </div>
                                {saving && (
                                    <p className="text-[11px] text-accent animate-pulse">
                                        Saving...
                                    </p>
                                )}
                            </div>
                        )}
                    </SheetHeader>

                    <div className="flex-1 overflow-y-auto px-5 py-4 space-y-1">
                        {isLoading ? (
                            <div className="space-y-2">
                                {[...Array(5)].map((_, i) => (
                                    <Skeleton
                                        key={i}
                                        className="h-14 w-full rounded-lg"
                                    />
                                ))}
                            </div>
                        ) : events.length === 0 ? (
                            <p className="text-sm text-secondary-text py-8 text-center">
                                No sessions found.
                            </p>
                        ) : (
                            events
                                .filter((e) => !e.is_idle)
                                .map((event) => (
                                    <button
                                        key={event.id}
                                        onClick={() => setActiveEvent(event)}
                                        className="w-full text-left rounded-lg px-3 py-2.5 hover:bg-divider/60 transition-colors group"
                                    >
                                        <div className="flex items-center justify-between gap-3">
                                            <div className="min-w-0 flex-1">
                                                <div className="flex items-center gap-2 text-xs">
                                                    <span className="font-medium text-ink tabular-nums">
                                                        {dayjs(
                                                            event.start_time,
                                                        ).format("h:mm A")}
                                                    </span>
                                                    <span className="text-secondary-text">
                                                        {"\u2192"}
                                                    </span>
                                                    <span className="text-secondary-text tabular-nums">
                                                        {dayjs(
                                                            event.end_time,
                                                        ).format("h:mm A")}
                                                    </span>
                                                </div>
                                                {event.window_title && (
                                                    <p className="text-xs text-secondary-text mt-0.5 truncate">
                                                        {event.window_title}
                                                    </p>
                                                )}
                                                {event.url && (
                                                    <p className="text-xs text-secondary-text/70 mt-0.5 truncate">
                                                        {event.url}
                                                    </p>
                                                )}
                                            </div>
                                            <span className="text-xs tabular-nums text-secondary-text shrink-0 group-hover:text-ink transition-colors">
                                                {formatDuration(
                                                    event.duration_secs,
                                                )}
                                            </span>
                                        </div>
                                    </button>
                                ))
                        )}
                    </div>

                    {!isLoading && events.length > 0 && (
                        <>
                            <Separator />
                            <div className="px-5 py-3 flex items-center justify-between text-xs text-secondary-text">
                                <span>
                                    {events.filter((e) => !e.is_idle).length}{" "}
                                    sessions on{" "}
                                    {dayjs(date).format("MMM D, YYYY")}
                                </span>
                                <span className="font-medium text-ink">
                                    {formatDuration(totalSecs)}
                                </span>
                            </div>
                        </>
                    )}
                </SheetContent>
            </Sheet>

            <EventDetailDrawer
                event={activeEvent}
                onClose={() => setActiveEvent(null)}
            />
        </>
    );
}
