/**
 * PlatformBadge — tiny corner badge that indicates whether an app
 * originates from Windows, Android, or a browser (Chrome-style).
 */
import { cn } from "~/lib/utils";

type Platform = "windows" | "android" | "browser";

interface PlatformBadgeProps {
    platform: Platform;
    className?: string;
}

function WindowsIcon() {
    return (
        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M0 2.286L6.545 1.4V7.43H0V2.286Z" fill="#0078D4" />
            <path d="M7.273 1.29L16 0v7.43H7.273V1.29Z" fill="#0078D4" />
            <path d="M0 8.57h6.545v6.03L0 13.714V8.57Z" fill="#0078D4" />
            <path d="M7.273 8.57H16V16l-8.727-1.29V8.57Z" fill="#0078D4" />
        </svg>
    );
}

function AndroidIcon() {
    return (
        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
                d="M2.75 10.5A5.25 5.25 0 0113.25 10.5H2.75Z"
                fill="#3DDC84"
            />
            <rect x="2" y="5.5" width="12" height="7" rx="1.5" fill="#3DDC84" />
            <circle cx="5.5" cy="8.75" r="0.75" fill="white" />
            <circle cx="10.5" cy="8.75" r="0.75" fill="white" />
            <line
                x1="5"
                y1="4.5"
                x2="3.5"
                y2="2.5"
                stroke="#3DDC84"
                strokeWidth="1.2"
                strokeLinecap="round"
            />
            <line
                x1="11"
                y1="4.5"
                x2="12.5"
                y2="2.5"
                stroke="#3DDC84"
                strokeWidth="1.2"
                strokeLinecap="round"
            />
        </svg>
    );
}

function ChromeIcon() {
    return (
        <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="8" cy="8" r="7.5" fill="white" />
            {/* outer ring segments */}
            <path d="M8 0.5A7.5 7.5 0 0114.99 7.5H8V0.5Z" fill="#EA4335" />
            <path
                d="M14.99 7.5A7.5 7.5 0 018 15.5L4.75 9.75 8 7.5h6.99Z"
                fill="#FBBC05"
            />
            <path
                d="M8 15.5A7.5 7.5 0 011.01 7.5H8L4.75 9.75 8 15.5Z"
                fill="#34A853"
            />
            <path d="M1.01 7.5A7.5 7.5 0 018 0.5V7.5H1.01Z" fill="#4285F4" />
            {/* inner white circle */}
            <circle cx="8" cy="8" r="3" fill="white" />
            {/* inner blue circle */}
            <circle cx="8" cy="8" r="2.25" fill="#4285F4" />
        </svg>
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
            {platform === "windows" && <WindowsIcon />}
            {platform === "android" && <AndroidIcon />}
            {platform === "browser" && <ChromeIcon />}
        </span>
    );
}
