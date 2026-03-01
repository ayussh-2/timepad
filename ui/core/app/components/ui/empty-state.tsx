import { cn } from "~/lib/utils";

interface EmptyStateProps {
    title: string;
    description?: string;
    className?: string;
    children?: React.ReactNode;
}

export function EmptyState({
    title,
    description,
    className,
    children,
}: EmptyStateProps) {
    return (
        <div
            className={cn(
                "flex flex-col items-center justify-center py-16 px-4 text-center",
                className,
            )}
        >
            <p className="font-display text-2xl text-ink">{title}</p>
            {description && (
                <p className="mt-2 text-sm text-secondary-text max-w-xs">
                    {description}
                </p>
            )}
            {children && <div className="mt-6">{children}</div>}
        </div>
    );
}
