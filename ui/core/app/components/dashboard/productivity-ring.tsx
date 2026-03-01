import { Cell, Pie, PieChart, ResponsiveContainer } from "recharts";
import type { DailySummary } from "~/app/types";
import { formatDuration } from "~/app/components/ui/duration";

const PRODUCTIVE_COLOR = "#7a9a6d";
const DISTRACTION_COLOR = "#c4a77d";
const NEUTRAL_COLOR = "#e5e1d8";

interface ProductivityRingProps {
    summary: DailySummary;
}

export function ProductivityRing({ summary }: ProductivityRingProps) {
    const neutral = Math.max(
        0,
        summary.total_active_secs -
            summary.productive_secs -
            summary.distraction_secs,
    );

    const data = [
        {
            name: "Productive",
            value: summary.productive_secs,
            color: PRODUCTIVE_COLOR,
        },
        {
            name: "Distraction",
            value: summary.distraction_secs,
            color: DISTRACTION_COLOR,
        },
        { name: "Neutral", value: neutral, color: NEUTRAL_COLOR },
    ].filter((d) => d.value > 0);

    if (data.length === 0) {
        data.push({ name: "None", value: 1, color: NEUTRAL_COLOR });
    }

    const productivePercent =
        summary.total_active_secs > 0
            ? Math.round(
                  (summary.productive_secs / summary.total_active_secs) * 100,
              )
            : 0;

    return (
        <div className="relative flex items-center justify-center h-40">
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
                    {productivePercent}%
                </span>
                <span className="text-xs text-secondary-text">productive</span>
            </div>

            <div className="absolute -bottom-6 left-0 right-0 flex justify-center gap-4">
                {[
                    {
                        label: "Productive",
                        color: PRODUCTIVE_COLOR,
                        secs: summary.productive_secs,
                    },
                    {
                        label: "Distraction",
                        color: DISTRACTION_COLOR,
                        secs: summary.distraction_secs,
                    },
                ].map(({ label, color, secs }) => (
                    <div key={label} className="flex items-center gap-1.5">
                        <span
                            className="h-2 w-2 rounded-full shrink-0"
                            style={{ background: color }}
                        />
                        <span className="text-xs text-secondary-text">
                            {label} · {formatDuration(secs)}
                        </span>
                    </div>
                ))}
            </div>
        </div>
    );
}
