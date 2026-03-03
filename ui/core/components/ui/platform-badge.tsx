import { cn } from "~/lib/utils";

type Platform = "windows" | "android" | "browser";

interface PlatformBadgeProps {
    platform: Platform;
    className?: string;
}

function WindowsIcon() {
    return (
        <img
            src="https://upload.wikimedia.org/wikipedia/commons/8/87/Windows_logo_-_2021.svg"
            alt="Windows"
            className="w-full h-full"
        />
    );
}

function AndroidIcon() {
    return (
        <img
            src="https://upload.wikimedia.org/wikipedia/commons/6/64/Android_logo_2019_%28stacked%29.svg"
            alt="Android"
            className="w-full h-full"
        />
    );
}

function ChromeIcon() {
    return (
        <img
            src="https://upload.wikimedia.org/wikipedia/commons/e/e1/Google_Chrome_icon_%28February_2022%29.svg"
            alt="Chrome"
            className="w-full h-full"
        />
    );
}

export function PlatformBadge({ platform, className }: PlatformBadgeProps) {
    return (
        <span
            className={cn(
                "inline-flex items-center justify-center rounded-sm w-3.5 h-3.5 shrink-0",
                className,
            )}
            title={
                platform === "windows"
                    ? "Windows"
                    : platform === "android"
                      ? "Android"
                      : "Browser"
            }
        >
            {/* <AndroidIcon /> */}
            {platform === "windows" && <WindowsIcon />}
            {platform === "android" && <AndroidIcon />}
            {platform === "browser" && <ChromeIcon />}
        </span>
    );
}
