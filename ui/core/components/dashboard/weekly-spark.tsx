import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from "recharts";
import type { WeeklySummary } from "~/app/types";

interface WeeklySparkProps {
    weekly: WeeklySummary;
}

export function WeeklySpark({ weekly }: WeeklySparkProps) {
    const data = weekly.daily_breakdown.map((day) => ({
        date: day.date.slice(5), // "MM-DD"
        active: Math.round(day.total_active_secs / 60), // minutes
    }));

    return (
        <ResponsiveContainer width="100%" height={64}>
            <BarChart data={data} barSize={8}>
                <XAxis
                    dataKey="date"
                    tick={{ fontSize: 10, fill: "#6e6a63" }}
                    axisLine={false}
                    tickLine={false}
                />
                <Tooltip
                    formatter={(v) => [`${v ?? 0}m`, "Active"]}
                    contentStyle={{
                        fontSize: 12,
                        borderColor: "#e5e1d8",
                        borderRadius: 8,
                        boxShadow: "0 1px 4px rgba(0,0,0,.06)",
                    }}
                />
                <Bar dataKey="active" fill="#5b7c99" radius={[3, 3, 0, 0]} />
            </BarChart>
        </ResponsiveContainer>
    );
}
