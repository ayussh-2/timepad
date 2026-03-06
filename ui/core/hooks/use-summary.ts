import { useEffect } from "react";
import { hasFresh } from "~/lib/request-cache";
import { useActivityStore } from "~/store/activity.store";

export function useSummary(date: string) {
    const dailySummary = useActivityStore((s) => s.dailySummary);
    const weeklySummary = useActivityStore((s) => s.weeklySummary);
    const isLoading = useActivityStore((s) => s.summaryLoading);
    const fetchDailySummary = useActivityStore((s) => s.fetchDailySummary);
    const fetchWeeklySummary = useActivityStore((s) => s.fetchWeeklySummary);

    useEffect(() => {
        fetchDailySummary(date);
        // Only fetch weekly if not already warm — it's non-critical.
        if (!hasFresh(`weekly:${date}`)) fetchWeeklySummary(date);
    }, [date, fetchDailySummary, fetchWeeklySummary]);

    return { dailySummary, weeklySummary, isLoading };
}
