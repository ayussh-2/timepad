import type { ApiEnvelope, DailySummary, WeeklySummary } from "~/app/types";
import { client } from "./client";

// Server emits snake_case JSON — response types match our TS types directly.
export const summaryApi = {
    daily: (date: string) =>
        client
            .get<ApiEnvelope<DailySummary>>("/summary/daily", {
                params: { date },
            })
            .then((r) => r.data.data),

    weekly: (date: string) =>
        client
            .get<ApiEnvelope<WeeklySummary>>("/summary/weekly", {
                params: { date },
            })
            .then((r) => r.data.data),
};
