import type { ApiEnvelope } from "~/app/types";
import type { Category } from "~/app/types";
import { client } from "./client";
import { type RawCategory, normalizeCategory } from "./categories";

export const eventsApi = {
    patch: (
        id: string,
        payload: { category_id?: string | null; is_private?: boolean },
    ) =>
        client
            .patch<ApiEnvelope<null>>(`/events/${id}`, payload)
            .then((r) => r.data),

    delete: (id: string) =>
        client.delete<ApiEnvelope<null>>(`/events/${id}`).then((r) => r.data),

    /**
     * Mark an entire app as Productive / Distraction / Neutral (null).
     * Automatically finds-or-creates a matching default category on the server.
     * Returns the assigned Category, or null when cleared.
     */
    classifyApp: (appName: string, isProductive: boolean | null) =>
        client
            .patch<
                ApiEnvelope<{ category: RawCategory | null }>
            >("/events/classify-app", { app_name: appName, is_productive: isProductive })
            .then((r) => {
                const raw = r.data.data?.category;
                return raw ? normalizeCategory(raw) : null;
            }),

    /** Set (or clear) the category for every event belonging to appName. */
    categorizeApp: (appName: string, categoryId: string | null) =>
        client
            .patch<ApiEnvelope<{ updated: number }>>("/events/categorize-app", {
                app_name: appName,
                category_id: categoryId,
            })
            .then((r) => r.data.data),
};
