import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from "recharts";
import { formatDuration } from "~/app/components/ui/duration";
import type { DeviceUsage } from "~/app/types";

interface DeviceBreakdownCardProps {
    devices: DeviceUsage[];
}

const COLORS = ["#5b7c99", "#7a9a6d", "#c4a77d", "#8b7b9e"];

export function DeviceBreakdownCard({ devices }: DeviceBreakdownCardProps) {
    if (devices.length === 0) return null;

    const data = devices.map((d) => ({
        name: d.device_name,
        minutes: Math.round(d.total_secs / 60),
    }));

    return (
        <div className="space-y-3">
            {devices.map((d, i) => {
                const total = devices.reduce((sum, x) => sum + x.total_secs, 0);
                const pct =
                    total > 0 ? Math.round((d.total_secs / total) * 100) : 0;
                return (
                    <div key={d.device_name} className="space-y-1">
                        <div className="flex justify-between text-sm">
                            <span className="text-ink">{d.device_name}</span>
                            <span className="text-secondary-text tabular-nums">
                                {formatDuration(d.total_secs)} · {pct}%
                            </span>
                        </div>
                        <div className="h-2 rounded-full bg-divider overflow-hidden">
                            <div
                                className="h-full rounded-full transition-all"
                                style={{
                                    width: `${pct}%`,
                                    background: COLORS[i % COLORS.length],
                                }}
                            />
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
