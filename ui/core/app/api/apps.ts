import type { ApiEnvelope, App } from "~/app/types";
import { client } from "./client";

export const appsApi = {
    /** List all tracked apps for the authenticated user. */
    list: () =>
        client.get<ApiEnvelope<App[]>>("/apps").then((r) => r.data.data ?? []),

    /**
     * Directly assign (or clear) a specific category on an app.
     * Pass categoryId = null to clear.
     */
    setCategory: (appId: string, categoryId: string | null) =>
        client
            .patch<ApiEnvelope<App>>(`/apps/${appId}/category`, {
                category_id: categoryId,
            })
            .then((r) => r.data.data),

    /**
     * Quick-classify: finds-or-creates a Productive / Distraction category
     * and assigns it to the app. Pass isProductive = null to clear (Neutral).
     */
    classify: (appId: string, isProductive: boolean | null) =>
        client
            .patch<ApiEnvelope<App>>(`/apps/${appId}/classify`, {
                is_productive: isProductive,
            })
            .then((r) => r.data.data),

    /** Mark or unmark an app as a system app (excluded from productivity stats). */
    setSystem: (appId: string, isSystem: boolean) =>
        client
            .patch<ApiEnvelope<App>>(`/apps/${appId}/system`, {
                is_system: isSystem,
            })
            .then((r) => r.data.data),
};
