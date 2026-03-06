import type { ApiEnvelope, ReportData } from "~/app/types";
import { cached } from "~/lib/request-cache";
import { client } from "./client";

const TTL = 5 * 60_000;

export const reportsApi = {
    get: (start_date?: string, end_date?: string) => {
        const key = `reports:${start_date ?? ""}:${end_date ?? ""}`;
        return cached(
            key,
            () =>
                client
                    .get<ApiEnvelope<ReportData>>("/reports", {
                        params: {
                            ...(start_date ? { start_date } : {}),
                            ...(end_date ? { end_date } : {}),
                        },
                    })
                    .then((r) => r.data.data),
            TTL,
        );
    },
};
