import dayjs from "dayjs";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/utils";

interface DateNavigatorProps {
    date: string;
    onChange: (date: string) => void;
    className?: string;
}

export function DateNavigator({
    date,
    onChange,
    className,
}: DateNavigatorProps) {
    const d = dayjs(date);
    const isToday = d.isSame(dayjs(), "day");

    return (
        <div className={cn("flex items-center gap-2", className)}>
            <Button
                variant="ghost"
                size="icon"
                onClick={() =>
                    onChange(d.subtract(1, "day").format("YYYY-MM-DD"))
                }
                className="h-8 w-8"
            >
                <ChevronLeft className="h-4 w-4" />
            </Button>

            <input
                type="date"
                value={date}
                max={dayjs().format("YYYY-MM-DD")}
                onChange={(e) => {
                    if (e.target.value) onChange(e.target.value);
                }}
                className="text-sm font-medium text-ink bg-transparent border-none outline-none cursor-pointer"
            />

            <Button
                variant="ghost"
                size="icon"
                onClick={() => onChange(d.add(1, "day").format("YYYY-MM-DD"))}
                disabled={isToday}
                className="h-8 w-8"
            >
                <ChevronRight className="h-4 w-4" />
            </Button>
        </div>
    );
}
