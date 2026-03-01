import { useCallback, useState } from "react";
import { reportsApi } from "~/app/api/reports";
import type { ReportData } from "~/app/types";

export function useReports() {
    const [data, setData] = useState<ReportData | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async (startDate?: string, endDate?: string) => {
        setIsLoading(true);
        setError(null);
        try {
            const result = await reportsApi.get(startDate, endDate);
            setData(result);
        } catch {
            setError("Failed to load reports");
        } finally {
            setIsLoading(false);
        }
    }, []);

    return { data, isLoading, error, fetch };
}
