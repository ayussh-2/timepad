import { useEffect } from "react";
import { useActivityStore } from "~/store/activity.store";

export function useSummary(date: string) {
    const dailySummary = useActivityStore((s) => s.dailySummary);
    const weeklySummary = useActivityStore((s) => s.weeklySummary);
    const isLoading = useActivityStore((s) => s.isLoading);
    const fetchDailySummary = useActivityStore((s) => s.fetchDailySummary);
    const fetchWeeklySummary = useActivityStore((s) => s.fetchWeeklySummary);

    useEffect(() => {
        fetchDailySummary(date);
        fetchWeeklySummary(date);
    }, [date, fetchDailySummary, fetchWeeklySummary]);

    return { dailySummary, weeklySummary, isLoading };
}
