/**
 * AppTimelineDrawer â€” slide-in sheet showing sessions for a specific app,
 * with a first-class app-level productivity classifier.
 */
import dayjs from "dayjs";
import { CheckCircle2, XCircle, Minus, ChevronDown } from "lucide-react";
import { useEffect, useState } from "react";
import { categoriesApi } from "~/app/api/categories";
import { eventsApi } from "~/app/api/events";
import { timelineApi } from "~/app/api/timeline";
import type { AppUsage, Category, TimelineEntry } from "~/app/types";
import { AppIcon } from "~/components/ui/app-icon";
import { formatDuration } from "~/components/ui/duration";
import { PlatformBadge } from "~/components/ui/platform-badge";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "~/components/ui/select";
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

type Productivity = "productive" | "distraction" | null;

function productivityFromCategory(cat: Category | null): Productivity {
    if (!cat) return null;
    if (cat.is_productive === true) return "productive";
    if (cat.is_productive === false) return "distraction";
    return null;
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
    const [showCategoryPicker, setShowCategoryPicker] = useState(false);

    // Sync state when a different app is selected
    useEffect(() => {
        setCurrentCategory(app?.category ?? null);
        setShowCategoryPicker(false);
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
                category_id: cat?.id ?? null,
                category: cat,
            })),
        );
        onCategoryChanged?.(app!.app_name, cat);
    };

    /** Called by the 3 big buttons */
    const handleQuickClassify = async (value: Productivity) => {
        if (!app || saving) return;
        setSaving(true);
        try {
            const isProductive =
                value === "productive"
                    ? true
                    : value === "distraction"
                      ? false
                      : null;
            const cat = await eventsApi.classifyApp(app.app_name, isProductive);
            applyCategory(cat);
        } catch {
            // silent
        } finally {
            setSaving(false);
        }
    };

    /** Called by the specific-category dropdown */
    const handleCategoryChange = async (value: string) => {
        if (!app || saving) return;
        setSaving(true);
        try {
            const catId = value === "__none__" ? null : value;
            await eventsApi.categorizeApp(app.app_name, catId);
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
    const currentProductivity = productivityFromCategory(currentCategory);
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
                    {/* â”€â”€ Header â”€â”€ */}
                    <SheetHeader className="px-5 pt-5 pb-4 border-b border-divider space-y-4">
                        {/* App name + icon */}
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

                        {/* Stats */}
                        {app && (
                            <div className="flex items-center gap-4 text-xs text-secondary-text">
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
                        )}

                        {/* â”€â”€ Productivity classifier â”€â”€ */}
                        {app && (
                            <div className="space-y-3">
                                <p className="text-xs font-medium text-secondary-text">
                                    How do you use{" "}
                                    <span className="text-ink">
                                        {app.app_name}
                                    </span>
                                    ?
                                </p>

                                {/* 3 big buttons */}
                                <div className="grid grid-cols-3 gap-2">
                                    <button
                                        onClick={() =>
                                            handleQuickClassify("productive")
                                        }
                                        disabled={saving}
                                        className={cn(
                                            "flex flex-col items-center gap-1 rounded-lg border py-3 px-2 text-xs font-medium transition-all",
                                            currentProductivity === "productive"
                                                ? "border-[#7a9a6d] bg-[#7a9a6d]/10 text-[#7a9a6d]"
                                                : "border-divider text-secondary-text hover:border-[#7a9a6d]/50 hover:text-[#7a9a6d]",
                                        )}
                                    >
                                        <CheckCircle2 className="h-5 w-5" />
                                        Productive
                                    </button>

                                    <button
                                        onClick={() =>
                                            handleQuickClassify(null)
                                        }
                                        disabled={saving}
                                        className={cn(
                                            "flex flex-col items-center gap-1 rounded-lg border py-3 px-2 text-xs font-medium transition-all",
                                            currentProductivity === null
                                                ? "border-secondary-text bg-secondary-text/10 text-secondary-text"
                                                : "border-divider text-secondary-text hover:border-secondary-text/50",
                                        )}
                                    >
                                        <Minus className="h-5 w-5" />
                                        Neutral
                                    </button>

                                    <button
                                        onClick={() =>
                                            handleQuickClassify("distraction")
                                        }
                                        disabled={saving}
                                        className={cn(
                                            "flex flex-col items-center gap-1 rounded-lg border py-3 px-2 text-xs font-medium transition-all",
                                            currentProductivity ===
                                                "distraction"
                                                ? "border-[#b45a5a] bg-[#b45a5a]/10 text-[#b45a5a]"
                                                : "border-divider text-secondary-text hover:border-[#b45a5a]/50 hover:text-[#b45a5a]",
                                        )}
                                    >
                                        <XCircle className="h-5 w-5" />
                                        Distraction
                                    </button>
                                </div>

                                {saving && (
                                    <p className="text-[11px] text-accent animate-pulse">
                                        Applying to all sessionsâ€¦
                                    </p>
                                )}

                                {/* Optional: specific category */}
                                <button
                                    onClick={() =>
                                        setShowCategoryPicker((v) => !v)
                                    }
                                    className="flex items-center gap-1 text-[11px] text-secondary-text hover:text-ink transition-colors"
                                >
                                    <ChevronDown
                                        className={cn(
                                            "h-3 w-3 transition-transform",
                                            showCategoryPicker && "rotate-180",
                                        )}
                                    />
                                    {currentCategory
                                        ? `Category: ${currentCategory.name}`
                                        : "Assign to a specific category"}
                                </button>

                                {showCategoryPicker && (
                                    <Select
                                        value={
                                            currentCategory?.id ?? "__none__"
                                        }
                                        onValueChange={handleCategoryChange}
                                        disabled={saving}
                                    >
                                        <SelectTrigger className="h-8 text-xs">
                                            <SelectValue placeholder="Select categoryâ€¦" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="__none__">
                                                <span className="text-secondary-text">
                                                    No category
                                                </span>
                                            </SelectItem>
                                            {categories.map((c) => (
                                                <SelectItem
                                                    key={c.id}
                                                    value={c.id}
                                                >
                                                    <span className="flex items-center gap-2">
                                                        <span
                                                            className="h-2.5 w-2.5 rounded-full shrink-0"
                                                            style={{
                                                                background:
                                                                    c.color,
                                                            }}
                                                        />
                                                        {c.name}
                                                        {c.is_productive ===
                                                            true && (
                                                            <CheckCircle2 className="h-3 w-3 text-[#7a9a6d] shrink-0" />
                                                        )}
                                                        {c.is_productive ===
                                                            false && (
                                                            <XCircle className="h-3 w-3 text-[#b45a5a] shrink-0" />
                                                        )}
                                                    </span>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                )}
                            </div>
                        )}
                    </SheetHeader>

                    {/* â”€â”€ Timeline list â”€â”€ */}
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
                                                        â†’
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

                    {/* â”€â”€ Footer â”€â”€ */}
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
