import type {
    ApiEnvelope,
    App,
    Device,
    TimelineEntry,
    TimelineResponse,
} from "~/app/types";
import { client } from "./client";

// Server now emits snake_case JSON \u2014 the response shape matches our types directly.
interface RawTimelineResponse {
    events: TimelineEntry[];
    next_cursor?: string | null;
}

export const timelineApi = {
    get: (
        date: string,
        cursor?: string | null,
        limit = 100,
        appName?: string,
    ) =>
        client
            .get<ApiEnvelope<RawTimelineResponse>>("/timeline", {
                params: {
                    date,
                    ...(cursor ? { cursor } : {}),
                    limit,
                    ...(appName ? { app_name: appName } : {}),
                },
            })
            .then(
                (r) =>
                    ({
                        events: r.data.data.events ?? [],
                        next_cursor: r.data.data.next_cursor ?? null,
                    }) satisfies TimelineResponse,
            ),
};
