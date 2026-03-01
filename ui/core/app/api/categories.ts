import type { ApiEnvelope, Category, CategoryRule } from "~/app/types";
import { client } from "./client";

interface RawCategory {
    ID: string;
    UserID: string | null;
    Name: string;
    Color: string;
    Icon: string;
    IsSystem: boolean;
    IsProductive: boolean | null;
    Rules: CategoryRule[] | null;
}

function normalizeCategory(c: RawCategory): Category {
    return {
        id: c.ID,
        user_id: c.UserID,
        name: c.Name,
        color: c.Color,
        icon: c.Icon,
        is_system: c.IsSystem,
        is_productive: c.IsProductive,
        rules: c.Rules ?? [],
    };
}

export const categoriesApi = {
    list: () =>
        client
            .get<ApiEnvelope<RawCategory[]>>("/categories")
            .then((r) => (r.data.data ?? []).map(normalizeCategory)),

    create: (payload: {
        name: string;
        color?: string;
        icon?: string;
        is_productive?: boolean | null;
    }) =>
        client
            .post<ApiEnvelope<RawCategory>>("/categories", payload)
            .then((r) => normalizeCategory(r.data.data)),

    update: (
        id: string,
        payload: Partial<{
            name: string;
            color: string;
            icon: string;
            is_productive: boolean | null;
            rules: import("~/app/types").CategoryRule[];
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
