import type { ApiEnvelope, ReportData } from "~/app/types";
import { client } from "./client";

export const reportsApi = {
    get: (start_date?: string, end_date?: string) =>
        client
            .get<ApiEnvelope<ReportData>>("/reports", {
                params: {
                    ...(start_date ? { start_date } : {}),
                    ...(end_date ? { end_date } : {}),
                },
            })
            .then((r) => r.data.data),
};
