import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { useReports } from "~/app/hooks/use-reports";
import { EmptyState } from "~/app/components/ui/empty-state";
import { formatDuration, Duration } from "~/app/components/ui/duration";
import {
    Bar,
    BarChart,
    Cell,
    Pie,
    PieChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Skeleton } from "~/components/ui/skeleton";

const CHART_COLORS = [
    "#5b7c99",
    "#7a9a6d",
    "#c4a77d",
    "#8b7b9e",
    "#6e6a63",
    "#a3796a",
];

export default function ReportsPage() {
    const [startDate, setStartDate] = useState(
        dayjs().subtract(6, "day").format("YYYY-MM-DD"),
    );
    const [endDate, setEndDate] = useState(dayjs().format("YYYY-MM-DD"));
    const { data, isLoading, error, fetch } = useReports();

    useEffect(() => {
        fetch(startDate, endDate);
    }, []);

    const handleFetch = () => fetch(startDate, endDate);

    const categoryData = data
        ? Object.entries(data.category_usage)
              .map(([name, secs]) => ({ name, secs }))
              .sort((a, b) => b.secs - a.secs)
        : [];

    const trendData = data
        ? Object.entries(data.daily_active_trend)
              .sort(([a], [b]) => a.localeCompare(b))
              .map(([date, secs]) => ({
                  date: date.slice(5),
                  minutes: Math.round(secs / 60),
              }))
        : [];

    const appData = data
        ? Object.entries(data.app_usage)
              .map(([name, secs]) => ({ name, secs }))
              .sort((a, b) => b.secs - a.secs)
              .slice(0, 15)
        : [];

    const deviceData = data
        ? Object.entries(data.device_usage).map(([name, secs]) => ({
              name,
              secs,
          }))
        : [];

    return (
        <div className="max-w-4xl mx-auto px-4 py-6 space-y-6">
            <div className="flex flex-wrap gap-3 items-end">
                <div className="space-y-1">
                    <Label>From</Label>
                    <Input
                        type="date"
                        value={startDate}
                        max={endDate}
                        onChange={(e) => setStartDate(e.target.value)}
                        className="w-40"
                    />
                </div>
                <div className="space-y-1">
                    <Label>To</Label>
                    <Input
                        type="date"
                        value={endDate}
                        min={startDate}
                        max={dayjs().format("YYYY-MM-DD")}
                        onChange={(e) => setEndDate(e.target.value)}
                        className="w-40"
                    />
                </div>
                <Button onClick={handleFetch} disabled={isLoading}>
                    {isLoading ? "Loading..." : "Apply"}
                </Button>
            </div>

            {error && <p className="text-sm text-destructive">{error}</p>}

            {isLoading && !data ? (
                <div className="grid gap-4 md:grid-cols-2">
                    {[...Array(4)].map((_, i) => (
                        <Card key={i}>
                            <CardContent className="pt-6">
                                <Skeleton className="h-40 w-full" />
                            </CardContent>
                        </Card>
                    ))}
                </div>
            ) : !data ? (
                <EmptyState
                    title="No data"
                    description="Select a date range and apply."
                />
            ) : (
                <div className="space-y-4">
                    {/* Totals */}
                    <div className="grid gap-4 sm:grid-cols-2">
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    Total active
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="font-display text-3xl text-ink">
                                    <Duration secs={data.total_active_secs} />
                                </p>
                            </CardContent>
                        </Card>
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    Total idle
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <p className="font-display text-3xl text-ink">
                                    <Duration secs={data.total_idle_secs} />
                                </p>
                            </CardContent>
                        </Card>
                    </div>

                    {/* Daily trend */}
                    {trendData.length > 0 && (
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    Daily active time
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <ResponsiveContainer width="100%" height={160}>
                                    <BarChart data={trendData} barSize={12}>
                                        <XAxis
                                            dataKey="date"
                                            tick={{
                                                fontSize: 10,
                                                fill: "#6e6a63",
                                            }}
                                            axisLine={false}
                                            tickLine={false}
                                        />
                                        <YAxis
                                            tick={{
                                                fontSize: 10,
                                                fill: "#6e6a63",
                                            }}
                                            axisLine={false}
                                            tickLine={false}
                                            unit="m"
                                        />
                                        <Tooltip
                                            formatter={(v) => [
                                                `${v ?? 0}m`,
                                                "Active",
                                            ]}
                                            contentStyle={{
                                                fontSize: 12,
                                                borderColor: "#e5e1d8",
                                                borderRadius: 8,
                                            }}
                                        />
                                        <Bar
                                            dataKey="minutes"
                                            fill="#5b7c99"
                                            radius={[3, 3, 0, 0]}
                                        />
                                    </BarChart>
                                </ResponsiveContainer>
                            </CardContent>
                        </Card>
                    )}

                    {/* Category breakdown */}
                    {categoryData.length > 0 && (
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    By category
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="flex flex-col sm:flex-row items-center gap-6">
                                <ResponsiveContainer width={200} height={160}>
                                    <PieChart>
                                        <Pie
                                            data={categoryData}
                                            dataKey="secs"
                                            nameKey="name"
                                            cx="50%"
                                            cy="50%"
                                            innerRadius={45}
                                            outerRadius={70}
                                            strokeWidth={0}
                                        >
                                            {categoryData.map((_, i) => (
                                                <Cell
                                                    key={i}
                                                    fill={
                                                        CHART_COLORS[
                                                            i %
                                                                CHART_COLORS.length
                                                        ]
                                                    }
                                                />
                                            ))}
                                        </Pie>
                                        <Tooltip
                                            formatter={(v) => [
                                                formatDuration(Number(v ?? 0)),
                                                "",
                                            ]}
                                            contentStyle={{
                                                fontSize: 12,
                                                borderColor: "#e5e1d8",
                                                borderRadius: 8,
                                            }}
                                        />
                                    </PieChart>
                                </ResponsiveContainer>
                                <div className="flex-1 space-y-2">
                                    {categoryData
                                        .slice(0, 6)
                                        .map(({ name, secs }, i) => (
                                            <div
                                                key={name}
                                                className="flex items-center gap-2 text-sm"
                                            >
                                                <span
                                                    className="h-2.5 w-2.5 rounded-full shrink-0"
                                                    style={{
                                                        background:
                                                            CHART_COLORS[
                                                                i %
                                                                    CHART_COLORS.length
                                                            ],
                                                    }}
                                                />
                                                <span className="flex-1 truncate text-ink">
                                                    {name}
                                                </span>
                                                <span className="text-secondary-text tabular-nums">
                                                    {formatDuration(secs)}
                                                </span>
                                            </div>
                                        ))}
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* App usage */}
                    {appData.length > 0 && (
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    App usage
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2">
                                    {appData.map(({ name, secs }) => {
                                        const total = appData.reduce(
                                            (s, a) => s + a.secs,
                                            0,
                                        );
                                        const pct =
                                            total > 0
                                                ? Math.round(
                                                      (secs / total) * 100,
                                                  )
                                                : 0;
                                        return (
                                            <div
                                                key={name}
                                                className="flex items-center gap-3 text-sm"
                                            >
                                                <span className="w-36 truncate text-ink">
                                                    {name}
                                                </span>
                                                <div className="flex-1 h-1.5 rounded-full bg-divider overflow-hidden">
                                                    <div
                                                        className="h-full rounded-full bg-accent/60"
                                                        style={{
                                                            width: `${pct}%`,
                                                        }}
                                                    />
                                                </div>
                                                <span className="w-14 text-right text-secondary-text tabular-nums text-xs">
                                                    {formatDuration(secs)}
                                                </span>
                                                <span className="w-8 text-right text-secondary-text tabular-nums text-xs">
                                                    {pct}%
                                                </span>
                                            </div>
                                        );
                                    })}
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* Device usage */}
                    {deviceData.length > 0 && (
                        <Card>
                            <CardHeader className="pb-2">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    By device
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="space-y-2">
                                {deviceData.map(({ name, secs }, i) => {
                                    const total = deviceData.reduce(
                                        (s, a) => s + a.secs,
                                        0,
                                    );
                                    const pct =
                                        total > 0
                                            ? Math.round((secs / total) * 100)
                                            : 0;
                                    const COLORS = [
                                        "#5b7c99",
                                        "#7a9a6d",
                                        "#c4a77d",
                                        "#8b7b9e",
                                    ];
                                    return (
                                        <div key={name} className="space-y-0.5">
                                            <div className="flex justify-between text-sm">
                                                <span className="text-ink">
                                                    {name}
                                                </span>
                                                <span className="text-secondary-text tabular-nums">
                                                    {formatDuration(secs)}
                                                </span>
                                            </div>
                                            <div className="h-2 rounded-full bg-divider overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all"
                                                    style={{
                                                        width: `${pct}%`,
                                                        background:
                                                            COLORS[
                                                                i %
                                                                    COLORS.length
                                                            ],
                                                    }}
                                                />
                                            </div>
                                        </div>
                                    );
                                })}
                            </CardContent>
                        </Card>
                    )}
                </div>
            )}
        </div>
    );
}
