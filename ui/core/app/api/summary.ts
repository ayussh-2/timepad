import type { ApiEnvelope, DailySummary, WeeklySummary } from "~/app/types";
import { client } from "./client";

export const summaryApi = {
    daily: (date: string) =>
        client
            .get<
                ApiEnvelope<DailySummary>
            >("/summary/daily", { params: { date } })
            .then((r) => r.data.data),

    weekly: (date: string) =>
        client
            .get<
                ApiEnvelope<WeeklySummary>
            >("/summary/weekly", { params: { date } })
            .then((r) => r.data.data),
};
