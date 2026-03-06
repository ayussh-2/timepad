import type { ApiEnvelope, Category, CategoryRule } from "~/app/types";
import { bust, cached } from "~/lib/request-cache";
import { client } from "./client";

const CACHE_KEY = "categories:list";
const TTL = 5 * 60_000;

export const categoriesApi = {
    list: () =>
        cached(
            CACHE_KEY,
            () =>
                client
                    .get<ApiEnvelope<Category[]>>("/categories")
                    .then((r) => r.data.data ?? []),
            TTL,
        ),

    create: (payload: {
        name: string;
        color?: string;
        icon?: string;
        is_productive?: boolean | null;
    }) =>
        client.post<ApiEnvelope<Category>>("/categories", payload).then((r) => {
            bust(CACHE_KEY);
            return r.data.data;
        }),

    update: (
        id: string,
        payload: Partial<{
            name: string;
            color: string;
            icon: string;
            is_productive: boolean | null;
            rules: CategoryRule[];
        }>,
    ) =>
        client
            .patch<ApiEnvelope<null>>(`/categories/${id}`, payload)
            .then((r) => {
                bust(CACHE_KEY);
                return r.data;
            }),

    delete: (id: string) =>
        client.delete<ApiEnvelope<null>>(`/categories/${id}`).then((r) => {
            bust(CACHE_KEY);
            return r.data;
        }),
};
