import { useEffect } from "react";
import { useActivityStore } from "~/store/activity.store";

const INTERVAL_MS = 30 * 60 * 1000;

export function useAutoRefresh() {
    const invalidate = useActivityStore((s) => s.invalidate);

    useEffect(() => {
        const id = setInterval(invalidate, INTERVAL_MS);

        const onVisible = () => {
            if (document.visibilityState === "visible") invalidate();
        };
        document.addEventListener("visibilitychange", onVisible);

        return () => {
            clearInterval(id);
            document.removeEventListener("visibilitychange", onVisible);
        };
    }, [invalidate]);
}
