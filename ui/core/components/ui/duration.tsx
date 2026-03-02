interface DurationProps {
    secs: number;
    className?: string;
}

export function formatDuration(secs: number): string {
    if (secs < 60) return `${secs}s`;
    const h = Math.floor(secs / 3600);
    const m = Math.floor((secs % 3600) / 60);
    if (h === 0) return `${m}m`;
    if (m === 0) return `${h}h`;
    return `${h}h ${m}m`;
}

export function Duration({ secs, className }: DurationProps) {
    return <span className={className}>{formatDuration(secs)}</span>;
}
