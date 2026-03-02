import { useState } from "react";
import { getAppIconUrl, getAppInitials } from "~/utils/app-icon";
import { cn } from "~/lib/utils";

interface AppIconProps {
    appName: string;
    url?: string;
    size?: "sm" | "md" | "lg";
    className?: string;
}

const SIZE = {
    sm: "h-5 w-5 text-[10px]",
    md: "h-7 w-7 text-xs",
    lg: "h-9 w-9 text-sm",
};

const IMG_SIZE = {
    sm: "h-3.5 w-3.5",
    md: "h-4 w-4",
    lg: "h-5 w-5",
};

export function AppIcon({
    appName,
    url,
    size = "md",
    className,
}: AppIconProps) {
    const iconUrl = getAppIconUrl(appName, url);
    const initials = getAppInitials(appName);
    const [failed, setFailed] = useState(false);

    return (
        <span
            className={cn(
                "inline-flex items-center justify-center rounded-md bg-divider shrink-0 font-medium text-secondary-text select-none",
                SIZE[size],
                className,
            )}
        >
            {iconUrl && !failed ? (
                <img
                    src={iconUrl}
                    alt={appName}
                    className={cn(IMG_SIZE[size], "object-contain")}
                    onError={() => setFailed(true)}
                    loading="lazy"
                />
            ) : (
                initials
            )}
        </span>
    );
}
