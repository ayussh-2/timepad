import { Cell, Pie, PieChart, ResponsiveContainer } from "recharts";
import type { DailySummary } from "~/app/types";
import { formatDuration } from "~/components/ui/duration";

const UNCATEGORIZED_COLOR = "#e5e1d8";

interface CategoryRingProps {
    summary: DailySummary;
}

export function CategoryRing({ summary }: CategoryRingProps) {
    // Aggregate seconds per category from top_apps
    const categoryMap = new Map<
        string,
        { secs: number; color: string; name: string }
    >();

    for (const app of summary.top_apps) {
        if (app.category) {
            const key = app.category.id ?? app.category.name;
            const existing = categoryMap.get(key);
            if (existing) {
                existing.secs += app.total_secs;
            } else {
                categoryMap.set(key, {
                    secs: app.total_secs,
                    color: app.category.color,
                    name: app.category.name,
                });
            }
        }
    }

    // Compute uncategorized time
    let categorizedSecs = 0;
    for (const entry of categoryMap.values()) {
        categorizedSecs += entry.secs;
    }
    const uncategorizedSecs = Math.max(
        0,
        summary.total_active_secs - categorizedSecs,
    );

    const data = Array.from(categoryMap.values())
        .map(({ name, secs, color }) => ({ name, value: secs, color }))
        .sort((a, b) => b.value - a.value);

    if (uncategorizedSecs > 0) {
        data.push({
            name: "Uncategorized",
            value: uncategorizedSecs,
            color: UNCATEGORIZED_COLOR,
        });
    }

    const isEmpty = data.length === 0;
    if (isEmpty) {
        data.push({ name: "None", value: 1, color: UNCATEGORIZED_COLOR });
    }

    const topEntry = data[0];
    const topPercent =
        !isEmpty && summary.total_active_secs > 0
            ? Math.round((topEntry.value / summary.total_active_secs) * 100)
            : 0;

    return (
        <div className="flex flex-col items-center gap-3">
            <div className="relative flex items-center justify-center h-40 w-full">
                <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                        <Pie
                            data={data}
                            cx="50%"
                            cy="50%"
                            innerRadius={52}
                            outerRadius={70}
                            dataKey="value"
                            strokeWidth={0}
                        >
                            {data.map((entry) => (
                                <Cell key={entry.name} fill={entry.color} />
                            ))}
                        </Pie>
                    </PieChart>
                </ResponsiveContainer>

                <div className="absolute flex flex-col items-center pointer-events-none">
                    <span className="font-display text-2xl text-ink">
                        {isEmpty ? "—" : `${topPercent}%`}
                    </span>
                    <span className="text-xs text-secondary-text text-center leading-tight max-w-[80px] truncate">
                        {isEmpty ? "no data" : topEntry.name}
                    </span>
                </div>
            </div>

            {/* Per-category legend */}
            {!isEmpty && (
                <div className="flex flex-wrap justify-center gap-x-4 gap-y-1.5">
                    {data.map(({ name, color, value }) => (
                        <div key={name} className="flex items-center gap-1.5">
                            <span
                                className="h-2 w-2 rounded-full shrink-0"
                                style={{ background: color }}
                            />
                            <span className="text-xs text-secondary-text">
                                {name} · {formatDuration(value)}
                            </span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
