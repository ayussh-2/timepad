import { useEffect } from "react";
import { useActivityStore } from "~/store/activity.store";

const INTERVAL_MS = 30 * 60_000;
// Only refresh on tab-focus if the user was away for at least this long.
const MIN_HIDDEN_MS = 60_000;

export function useAutoRefresh() {
    const invalidate = useActivityStore((s) => s.invalidate);

    useEffect(() => {
        const id = setInterval(invalidate, INTERVAL_MS);

        let hiddenAt = 0;
        const onVisible = () => {
            if (document.visibilityState === "hidden") {
                hiddenAt = Date.now();
            } else if (document.visibilityState === "visible") {
                if (Date.now() - hiddenAt >= MIN_HIDDEN_MS) invalidate();
            }
        };
        document.addEventListener("visibilitychange", onVisible);

        return () => {
            clearInterval(id);
            document.removeEventListener("visibilitychange", onVisible);
        };
    }, [invalidate]);
}
