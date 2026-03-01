import type { ApiEnvelope, TimelineEntry, TimelineResponse } from "~/app/types";
import { client } from "./client";

interface RawEvent {
    ID: string;
    UserID: string;
    DeviceID: string;
    AppName: string;
    WindowTitle: string;
    Url: string;
    CategoryID: string | null;
    StartTime: string;
    EndTime: string;
    DurationSecs: number;
    IsIdle: boolean;
    IsPrivate: boolean;
}

interface RawTimelineResponse {
    events: RawEvent[];
    next_cursor?: string | null;
}

function normalizeEvent(e: RawEvent): TimelineEntry {
    return {
        id: e.ID,
        user_id: e.UserID,
        device_id: e.DeviceID,
        app_name: e.AppName,
        window_title: e.WindowTitle,
        url: e.Url,
        category_id: e.CategoryID,
        category: null,
        device: null,
        start_time: e.StartTime,
        end_time: e.EndTime,
        duration_secs: e.DurationSecs,
        is_idle: e.IsIdle,
        is_private: e.IsPrivate,
    };
}

export const timelineApi = {
    get: (date: string, cursor?: string | null, limit = 100) =>
        client
            .get<ApiEnvelope<RawTimelineResponse>>("/timeline", {
                params: { date, ...(cursor ? { cursor } : {}), limit },
            })
            .then(
                (r) =>
                    ({
                        events: (r.data.data.events ?? []).map(normalizeEvent),
                        next_cursor: r.data.data.next_cursor ?? null,
                    }) satisfies TimelineResponse,
            ),
};
