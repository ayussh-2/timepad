import dayjs from "dayjs";
import { RefreshCw } from "lucide-react";
import { useState } from "react";
import { AppTimelineDrawer } from "~/components/dashboard/app-timeline-drawer";
import { DeviceBreakdownCard } from "~/components/dashboard/device-breakdown-card";
import { ProductivityRing } from "~/components/dashboard/productivity-ring";
import { TopAppsCard } from "~/components/dashboard/top-apps-card";
import { WeeklySpark } from "~/components/dashboard/weekly-spark";
import { DateNavigator } from "~/components/ui/date-navigator";
import { Duration } from "~/components/ui/duration";
import { EmptyState } from "~/components/ui/empty-state";
import { useSummary } from "~/hooks/use-summary";
import { useActivityStore } from "~/store/activity.store";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Skeleton } from "~/components/ui/skeleton";
import type { AppUsage } from "~/app/types";

function SummarySkeleton() {
    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[...Array(4)].map((_, i) => (
                <Card key={i}>
                    <CardHeader>
                        <Skeleton className="h-4 w-24" />
                    </CardHeader>
                    <CardContent>
                        <Skeleton className="h-20 w-full" />
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}

export default function DashboardPage() {
    const selectedDate = useActivityStore((s) => s.selectedDate);
    const setSelectedDate = useActivityStore((s) => s.setSelectedDate);
    const invalidate = useActivityStore((s) => s.invalidate);

    const { dailySummary, weeklySummary, isLoading } = useSummary(selectedDate);
    const [selectedApp, setSelectedApp] = useState<AppUsage | null>(null);

    const handleDateChange = (date: string) => {
        setSelectedDate(date);
    };

    return (
        <>
            <div className="max-w-4xl mx-auto px-4 py-6 space-y-6">
                <div className="flex items-center justify-between">
                    <DateNavigator
                        date={selectedDate}
                        onChange={handleDateChange}
                    />
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={invalidate}
                        disabled={isLoading}
                        className="h-8 w-8"
                    >
                        <RefreshCw
                            className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`}
                        />
                    </Button>
                </div>

                {isLoading && !dailySummary ? (
                    <SummarySkeleton />
                ) : !dailySummary ? (
                    <EmptyState
                        title="No data yet"
                        description="Install a collector on your device to start tracking."
                    />
                ) : (
                    <div className="space-y-4">
                        {/* Active time + productivity row */}
                        <div className="grid gap-4 sm:grid-cols-2">
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-secondary-text">
                                        Active time
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <p className="font-display text-4xl text-ink">
                                        <Duration
                                            secs={
                                                dailySummary.total_active_secs
                                            }
                                        />
                                    </p>
                                    <p className="mt-1 text-xs text-secondary-text">
                                        Idle:{" "}
                                        <Duration
                                            secs={dailySummary.total_idle_secs}
                                        />
                                    </p>
                                    {dailySummary.peak_hour !== undefined && (
                                        <p className="mt-2 text-xs text-secondary-text">
                                            Most active at{" "}
                                            {dayjs()
                                                .hour(dailySummary.peak_hour)
                                                .format("h A")}
                                        </p>
                                    )}
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-secondary-text">
                                        Productivity
                                    </CardTitle>
                                </CardHeader>
                                <CardContent className="pt-2 pb-8">
                                    <ProductivityRing summary={dailySummary} />
                                </CardContent>
                            </Card>
                        </div>

                        {/* Weekly spark */}
                        {weeklySummary && (
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-secondary-text">
                                        This week
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <WeeklySpark weekly={weeklySummary} />
                                </CardContent>
                            </Card>
                        )}

                        {/* Top apps */}
                        {dailySummary.top_apps.length > 0 && (
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-secondary-text">
                                        Top apps
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <TopAppsCard
                                        apps={dailySummary.top_apps}
                                        onAppClick={setSelectedApp}
                                    />
                                </CardContent>
                            </Card>
                        )}

                        {/* Device breakdown */}
                        {dailySummary.device_breakdown.length > 0 && (
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-secondary-text">
                                        Devices
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <DeviceBreakdownCard
                                        devices={dailySummary.device_breakdown}
                                    />
                                </CardContent>
                            </Card>
                        )}
                    </div>
                )}
            </div>

            <AppTimelineDrawer
                app={selectedApp}
                date={selectedDate}
                onClose={() => setSelectedApp(null)}
            />
        </>
    );
}
