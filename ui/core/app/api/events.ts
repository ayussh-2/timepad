import type { ApiEnvelope } from "~/app/types";
import { client } from "./client";

export const eventsApi = {
    patch: (id: string, payload: { is_private?: boolean }) =>
        client
            .patch<ApiEnvelope<null>>(`/events/${id}`, payload)
            .then((r) => r.data),

    delete: (id: string) =>
        client.delete<ApiEnvelope<null>>(`/events/${id}`).then((r) => r.data),
};
