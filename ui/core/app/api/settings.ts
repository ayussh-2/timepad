import type { ApiEnvelope, UserSettings } from "~/app/types";
import { cached } from "~/lib/request-cache";
import { client } from "./client";

const CACHE_KEY = "settings";
const TTL = 5 * 60_000;

interface RawSettings {
    UserID: string;
    ExcludedApps: string[] | null;
    ExcludedUrls: string[] | null;
    IdleThreshold: number;
    TrackingEnabled: boolean;
    DataRetentionDays: number;
    UpdatedAt: string;
}

function normalizeSettings(s: RawSettings): UserSettings {
    return {
        user_id: s.UserID,
        excluded_apps: s.ExcludedApps ?? [],
        excluded_urls: s.ExcludedUrls ?? [],
        idle_threshold: s.IdleThreshold,
        tracking_enabled: s.TrackingEnabled,
        data_retention_days: s.DataRetentionDays,
        updated_at: s.UpdatedAt,
    };
}

export const settingsApi = {
    get: () =>
        cached(
            CACHE_KEY,
            () =>
                client
                    .get<ApiEnvelope<RawSettings>>("/settings")
                    .then((r) => normalizeSettings(r.data.data)),
            TTL,
        ),

    update: (payload: Partial<Omit<UserSettings, "user_id" | "updated_at">>) =>
        client.put<ApiEnvelope<null>>("/settings", payload).then((r) => r.data),
};
