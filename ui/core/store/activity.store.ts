import dayjs from "dayjs";
import { create } from "zustand";
import { summaryApi } from "~/app/api/summary";
import { timelineApi } from "~/app/api/timeline";
import { bust, cached, hasFresh, ttlFor } from "~/lib/request-cache";
import type { DailySummary, TimelineEntry, WeeklySummary } from "~/app/types";

interface DateCache {
    timeline: TimelineEntry[];
    timelineCursor: string | null;
    timelineHasMore: boolean;
    dailySummary: DailySummary | null;
    weeklySummary: WeeklySummary | null;
    lastFetchedAt: number;
}

// Module-level — persists across renders, not subject to Zustand re-renders.
const dateCache = new Map<string, DateCache>();

function emptyDateCache(): DateCache {
    return {
        timeline: [],
        timelineCursor: null,
        timelineHasMore: false,
        dailySummary: null,
        weeklySummary: null,
        lastFetchedAt: 0,
    };
}

function getOrCreate(date: string): DateCache {
    if (!dateCache.has(date)) dateCache.set(date, emptyDateCache());
    return dateCache.get(date)!;
}

interface ActivityState {
    selectedDate: string;
    timeline: TimelineEntry[];
    timelineCursor: string | null;
    timelineHasMore: boolean;
    dailySummary: DailySummary | null;
    weeklySummary: WeeklySummary | null;
    // Granular flags so timeline and summary don't block each other's UI.
    summaryLoading: boolean;
    timelineLoading: boolean;
    isLoading: boolean; // union of both — kept for backward compat
    error: string | null;

    setSelectedDate: (date: string) => void;
    fetchTimeline: (date: string, cursor?: string | null) => Promise<void>;
    fetchDailySummary: (date: string) => Promise<void>;
    fetchWeeklySummary: (date: string) => Promise<void>;
    invalidate: () => void;
    updateEvent: (id: string, patch: Partial<TimelineEntry>) => void;
    removeEvent: (id: string) => void;
}

// Monotonically increasing IDs — lets each fetch verify it's still the latest.
let tlGen = 0;
let sumGen = 0;
let wkGen = 0;

const todayStr = () => dayjs().format("YYYY-MM-DD");

export const useActivityStore = create<ActivityState>((set, get) => ({
    selectedDate: todayStr(),
    timeline: [],
    timelineCursor: null,
    timelineHasMore: false,
    dailySummary: null,
    weeklySummary: null,
    summaryLoading: false,
    timelineLoading: false,
    isLoading: false,
    error: null,

    setSelectedDate: (date) => {
        const c = getOrCreate(date);
        set({
            selectedDate: date,
            timeline: c.timeline,
            timelineCursor: c.timelineCursor,
            timelineHasMore: c.timelineHasMore,
            dailySummary: c.dailySummary,
            weeklySummary: c.weeklySummary,
            // Reset loading flags so a discarded in-flight fetch for a previous
            // date never leaves the new date's UI in a stuck loading/empty state.
            summaryLoading: false,
            timelineLoading: false,
            isLoading: false,
            error: null,
        });
    },

    fetchTimeline: async (date, cursor = null) => {
        const gen = ++tlGen;
        const warm = !cursor && hasFresh(`timeline:${date}`);
        if (!warm) set({ timelineLoading: true, isLoading: true, error: null });
        try {
            let data: Awaited<ReturnType<typeof timelineApi.get>>;
            if (cursor) {
                data = await timelineApi.get(date, cursor);
            } else {
                data = await cached(
                    `timeline:${date}`,
                    () => timelineApi.get(date),
                    ttlFor(date),
                );
            }
            if (gen !== tlGen) return; // superseded by a newer fetch
            const c = getOrCreate(date);
            c.timeline = cursor ? [...c.timeline, ...data.events] : data.events;
            c.timelineCursor = data.next_cursor ?? null;
            c.timelineHasMore = !!data.next_cursor;
            set({
                timeline: c.timeline,
                timelineCursor: c.timelineCursor,
                timelineHasMore: c.timelineHasMore,
                timelineLoading: false,
                isLoading: get().summaryLoading,
            });
        } catch {
            if (gen !== tlGen) return;
            set({
                error: "Failed to load timeline",
                timelineLoading: false,
                isLoading: get().summaryLoading,
            });
        }
    },

    fetchDailySummary: async (date) => {
        const gen = ++sumGen;
        if (!hasFresh(`daily:${date}`))
            set({ summaryLoading: true, isLoading: true, error: null });
        try {
            const data = await cached(
                `daily:${date}`,
                () => summaryApi.daily(date),
                ttlFor(date),
            );
            if (gen !== sumGen) return;
            const c = getOrCreate(date);
            c.dailySummary = data;
            c.lastFetchedAt = Date.now();
            set({
                dailySummary: data,
                summaryLoading: false,
                isLoading: get().timelineLoading,
            });
        } catch {
            if (gen !== sumGen) return;
            set({
                error: "Failed to load summary",
                summaryLoading: false,
                isLoading: get().timelineLoading,
            });
        }
    },

    fetchWeeklySummary: async (date) => {
        const gen = ++wkGen;
        try {
            const data = await cached(
                `weekly:${date}`,
                () => summaryApi.weekly(date),
                ttlFor(date),
            );
            if (gen !== wkGen) return;
            const c = getOrCreate(date);
            c.weeklySummary = data;
            set({ weeklySummary: data });
        } catch {
            // non-critical
        }
    },

    invalidate: () => {
        const {
            selectedDate,
            fetchDailySummary,
            fetchWeeklySummary,
            fetchTimeline,
            setSelectedDate,
        } = get();
        const today = todayStr();
        // If the date has drifted past midnight, snap to today first.
        if (selectedDate !== today) {
            setSelectedDate(today);
            return;
        }
        const c = getOrCreate(selectedDate);
        if (Date.now() - c.lastFetchedAt < 10_000) return;
        bust(`timeline:${selectedDate}`);
        bust(`daily:${selectedDate}`);
        bust(`weekly:${selectedDate}`);
        fetchTimeline(selectedDate);
        fetchDailySummary(selectedDate);
        fetchWeeklySummary(selectedDate);
    },

    updateEvent: (id, patch) => {
        const { selectedDate, timeline } = get();
        const updated = timeline.map((e) =>
            e.id === id ? { ...e, ...patch } : e,
        );
        getOrCreate(selectedDate).timeline = updated;
        set({ timeline: updated });
    },

    removeEvent: (id) => {
        const { selectedDate, timeline } = get();
        const updated = timeline.filter((e) => e.id !== id);
        getOrCreate(selectedDate).timeline = updated;
        set({ timeline: updated });
    },
}));
