import { AppIcon } from "~/components/ui/app-icon";
import { Badge } from "~/components/ui/badge";
import { formatDuration } from "~/components/ui/duration";
import type { AppUsage } from "~/app/types";

interface TopAppsCardProps {
    apps: AppUsage[];
}

export function TopAppsCard({ apps }: TopAppsCardProps) {
    if (apps.length === 0) return null;

    const max = apps[0].total_secs;

    return (
        <div className="space-y-3">
            {apps.slice(0, 8).map((app) => (
                <div key={app.app_name} className="space-y-1">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2 min-w-0">
                            <AppIcon appName={app.app_name} size="sm" />
                            <span className="text-sm text-ink truncate">
                                {app.app_name}
                            </span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                            {app.category && (
                                <Badge
                                    variant="secondary"
                                    className="text-xs py-0"
                                >
                                    {app.category.name}
                                </Badge>
                            )}
                            <span className="text-xs text-secondary-text tabular-nums w-12 text-right">
                                {formatDuration(app.total_secs)}
                            </span>
                        </div>
                    </div>
                    <div className="h-1 rounded-full bg-divider overflow-hidden">
                        <div
                            className="h-full rounded-full bg-accent/50 transition-all"
                            style={{
                                width: `${(app.total_secs / max) * 100}%`,
                            }}
                        />
                    </div>
                </div>
            ))}
        </div>
    );
}
