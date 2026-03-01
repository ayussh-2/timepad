import { useEffect } from "react";
import { useActivityStore } from "~/app/store/activity.store";

export function useTimeline(date: string) {
    const timeline = useActivityStore((s) => s.timeline);
    const cursor = useActivityStore((s) => s.timelineCursor);
    const hasMore = useActivityStore((s) => s.timelineHasMore);
    const isLoading = useActivityStore((s) => s.isLoading);
    const fetchTimeline = useActivityStore((s) => s.fetchTimeline);

    useEffect(() => {
        fetchTimeline(date);
    }, [date, fetchTimeline]);

    const loadMore = () => {
        if (hasMore && cursor) fetchTimeline(date, cursor);
    };

    return { timeline, hasMore, isLoading, loadMore };
}
