import type { ApiEnvelope, App } from "~/app/types";
import { bust, cached } from "~/lib/request-cache";
import { client } from "./client";

const CACHE_KEY = "apps:list";
const TTL = 2 * 60_000;

export const appsApi = {
    list: () =>
        cached(
            CACHE_KEY,
            () =>
                client
                    .get<ApiEnvelope<App[]>>("/apps")
                    .then((r) => r.data.data ?? []),
            TTL,
        ),

    setCategory: (appId: string, categoryId: string | null) =>
        client
            .patch<
                ApiEnvelope<App>
            >(`/apps/${appId}/category`, { category_id: categoryId })
            .then((r) => {
                bust(CACHE_KEY);
                return r.data.data;
            }),

    classify: (appId: string, isProductive: boolean | null) =>
        client
            .patch<
                ApiEnvelope<App>
            >(`/apps/${appId}/classify`, { is_productive: isProductive })
            .then((r) => {
                bust(CACHE_KEY);
                return r.data.data;
            }),

    setSystem: (appId: string, isSystem: boolean) =>
        client
            .patch<
                ApiEnvelope<App>
            >(`/apps/${appId}/system`, { is_system: isSystem })
            .then((r) => {
                bust(CACHE_KEY);
                return r.data.data;
            }),
};
