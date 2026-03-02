import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import type { AppUsage } from "~/app/types";
import { AppIcon } from "~/components/ui/app-icon";
import { formatDuration } from "~/components/ui/duration";
import { PlatformBadge } from "~/components/ui/platform-badge";
import { Toggle } from "~/components/ui/toggle";
import { cn } from "~/lib/utils";
import { detectPlatform, isSystemApp } from "~/utils/app-icon";

interface TopAppsCardProps {
    apps: AppUsage[];
    onAppClick: (app: AppUsage) => void;
}

/** Returns the bar + dot colour for an app based on its category productivity. */
function productivityColor(app: AppUsage): { bar: string; dot: string } | null {
    if (!app.category) return null;
    if (app.category.is_productive === true)
        return { bar: "#7a9a6d", dot: "#7a9a6d" };
    if (app.category.is_productive === false)
        return { bar: "#b45a5a", dot: "#b45a5a" };
    return null;
}

export function TopAppsCard({ apps, onAppClick }: TopAppsCardProps) {
    const [hideSystem, setHideSystem] = useState(true);

    if (apps.length === 0) return null;

    const visible = hideSystem
        ? apps.filter((a) => !isSystemApp(a.app_name))
        : apps;
    const displayed = visible.slice(0, 8);
    const hiddenCount = apps.length - visible.length;
    const max = displayed[0]?.total_secs ?? 1;

    return (
        <div className="space-y-1">
            {/* Toggle row */}
            <div className="flex items-center justify-end pb-1">
                <Toggle
                    pressed={hideSystem}
                    onPressedChange={setHideSystem}
                    size="sm"
                    variant="outline"
                    className="h-6 gap-1.5 text-xs px-2"
                    aria-label="Toggle system apps"
                >
                    {hideSystem ? (
                        <EyeOff className="h-3 w-3" />
                    ) : (
                        <Eye className="h-3 w-3" />
                    )}
                    {hideSystem ? "System hidden" : "Show all"}
                </Toggle>
            </div>

            {displayed.length === 0 ? (
                <p className="text-sm text-secondary-text py-4 text-center">
                    No user apps recorded yet.
                </p>
            ) : (
                <div className="space-y-3">
                    {displayed.map((app) => {
                        const platform = detectPlatform(
                            app.app_name,
                            app.platforms,
                        );
                        const pColor = productivityColor(app);
                        return (
                            <button
                                key={app.app_name}
                                onClick={() => onAppClick(app)}
                                className="w-full text-left space-y-1 group"
                            >
                                <div className="flex items-center justify-between gap-2">
                                    <div className="flex items-center gap-2 min-w-0">
                                        <div className="relative shrink-0">
                                            <AppIcon
                                                appName={app.app_name}
                                                size="sm"
                                            />
                                            {platform && (
                                                <span className="absolute -bottom-1 -right-1">
                                                    <PlatformBadge
                                                        platform={platform}
                                                        className="w-3 h-3"
                                                    />
                                                </span>
                                            )}
                                        </div>
                                        <div className="flex items-center gap-1.5 min-w-0">
                                            {/* Productivity dot */}
                                            {pColor && (
                                                <span
                                                    className="h-1.5 w-1.5 rounded-full shrink-0"
                                                    style={{
                                                        background: pColor.dot,
                                                    }}
                                                />
                                            )}
                                            <span className="text-sm text-ink truncate group-hover:text-accent transition-colors">
                                                {app.app_name}
                                            </span>
                                        </div>
                                    </div>
                                    <span className="text-xs text-secondary-text tabular-nums w-12 text-right shrink-0">
                                        {formatDuration(app.total_secs)}
                                    </span>
                                </div>
                                <div className="h-1 rounded-full bg-divider overflow-hidden">
                                    <div
                                        className={cn(
                                            "h-full rounded-full transition-all",
                                            pColor
                                                ? "opacity-60 group-hover:opacity-80"
                                                : "bg-accent/50 group-hover:bg-accent/70",
                                        )}
                                        style={{
                                            width: `${(app.total_secs / max) * 100}%`,
                                            ...(pColor
                                                ? {
                                                      backgroundColor:
                                                          pColor.bar,
                                                  }
                                                : {}),
                                        }}
                                    />
                                </div>
                            </button>
                        );
                    })}
                </div>
            )}

            {hideSystem && hiddenCount > 0 && (
                <p className="text-xs text-secondary-text text-center pt-2">
                    {hiddenCount} system app{hiddenCount !== 1 ? "s" : ""}{" "}
                    hidden
                </p>
            )}
        </div>
    );
}

interface TopAppsCardProps {
    apps: AppUsage[];
    onAppClick: (app: AppUsage) => void;
}
