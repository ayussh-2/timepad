import dayjs from "dayjs";
import { create } from "zustand";
import { summaryApi } from "~/app/api/summary";
import { timelineApi } from "~/app/api/timeline";
import type { DailySummary, TimelineEntry, WeeklySummary } from "~/app/types";

interface ActivityState {
    timeline: TimelineEntry[];
    timelineCursor: string | null;
    timelineHasMore: boolean;
    dailySummary: DailySummary | null;
    weeklySummary: WeeklySummary | null;
    selectedDate: string;
    isLoading: boolean;
    error: string | null;

    setSelectedDate: (date: string) => void;
    fetchTimeline: (date: string, cursor?: string | null) => Promise<void>;
    fetchDailySummary: (date: string) => Promise<void>;
    fetchWeeklySummary: (date: string) => Promise<void>;
    invalidate: () => void;
    updateEvent: (id: string, patch: Partial<TimelineEntry>) => void;
    removeEvent: (id: string) => void;
}

export const useActivityStore = create<ActivityState>((set, get) => ({
    timeline: [],
    timelineCursor: null,
    timelineHasMore: false,
    dailySummary: null,
    weeklySummary: null,
    selectedDate: dayjs().format("YYYY-MM-DD"),
    isLoading: false,
    error: null,

    setSelectedDate: (date) => set({ selectedDate: date }),

    fetchTimeline: async (date, cursor = null) => {
        set({ isLoading: true, error: null });
        try {
            const data = await timelineApi.get(date, cursor);
            set((s) => ({
                timeline: cursor
                    ? [...s.timeline, ...data.events]
                    : data.events,
                timelineCursor: data.next_cursor ?? null,
                timelineHasMore: !!data.next_cursor,
                isLoading: false,
            }));
        } catch {
            set({ error: "Failed to load timeline", isLoading: false });
        }
    },

    fetchDailySummary: async (date) => {
        set({ isLoading: true, error: null });
        try {
            const data = await summaryApi.daily(date);
            set({ dailySummary: data, isLoading: false });
        } catch {
            set({ error: "Failed to load summary", isLoading: false });
        }
    },

    fetchWeeklySummary: async (date) => {
        try {
            const data = await summaryApi.weekly(date);
            set({ weeklySummary: data });
        } catch {
            // non-critical — weekly spark can be absent
        }
    },

    invalidate: () => {
        const {
            selectedDate,
            fetchDailySummary,
            fetchWeeklySummary,
            fetchTimeline,
        } = get();
        fetchTimeline(selectedDate);
        fetchDailySummary(selectedDate);
        fetchWeeklySummary(selectedDate);
    },

    updateEvent: (id, patch) =>
        set((s) => ({
            timeline: s.timeline.map((e) =>
                e.id === id ? { ...e, ...patch } : e,
            ),
        })),

    removeEvent: (id) =>
        set((s) => ({ timeline: s.timeline.filter((e) => e.id !== id) })),
}));
