import type { ApiEnvelope, Category, CategoryRule } from "~/app/types";
import { client } from "./client";

// Server now emits snake_case JSON thanks to struct tags — RawCategory = Category.
export type RawCategory = Category;

export function normalizeCategory(c: RawCategory): Category {
    return c;
}

export const categoriesApi = {
    list: () =>
        client
            .get<ApiEnvelope<Category[]>>("/categories")
            .then((r) => r.data.data ?? []),

    create: (payload: {
        name: string;
        color?: string;
        icon?: string;
        is_productive?: boolean | null;
    }) =>
        client
            .post<ApiEnvelope<Category>>("/categories", payload)
            .then((r) => r.data.data),

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
            .then((r) => r.data),

    delete: (id: string) =>
        client
            .delete<ApiEnvelope<null>>(`/categories/${id}`)
            .then((r) => r.data),
};
