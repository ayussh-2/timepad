import dayjs from "dayjs";
import { useState } from "react";
import { eventsApi } from "~/app/api/events";
import { categoriesApi } from "~/app/api/categories";
import { formatDuration } from "~/components/ui/duration";
import { useActivityStore } from "~/store/activity.store";
import type { Category, TimelineEntry } from "~/app/types";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "~/components/ui/alert-dialog";
import { Button } from "~/components/ui/button";
import { Label } from "~/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "~/components/ui/select";
import { Separator } from "~/components/ui/separator";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
} from "~/components/ui/sheet";
import { Toggle } from "~/components/ui/toggle";
import { useEffect } from "react";

interface EventDetailDrawerProps {
    event: TimelineEntry | null;
    onClose: () => void;
}

export function EventDetailDrawer({ event, onClose }: EventDetailDrawerProps) {
    const updateEvent = useActivityStore((s) => s.updateEvent);
    const removeEvent = useActivityStore((s) => s.removeEvent);
    const [categories, setCategories] = useState<Category[]>([]);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        categoriesApi
            .list()
            .then(setCategories)
            .catch(() => {});
    }, []);

    if (!event) return null;

    const handleCategoryChange = async (categoryId: string) => {
        setSaving(true);
        try {
            const val = categoryId === "__none__" ? null : categoryId;
            await eventsApi.patch(event.id, { category_id: val });
            const cat = categories.find((c) => c.id === categoryId) ?? null;
            updateEvent(event.id, { category_id: val, category: cat });
        } finally {
            setSaving(false);
        }
    };

    const handlePrivacyToggle = async () => {
        setSaving(true);
        try {
            await eventsApi.patch(event.id, { is_private: !event.is_private });
            updateEvent(event.id, { is_private: !event.is_private });
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = async () => {
        await eventsApi.delete(event.id);
        removeEvent(event.id);
        onClose();
    };

    return (
        <Sheet open={!!event} onOpenChange={(open) => !open && onClose()}>
            <SheetContent side="right" className="w-full sm:max-w-md px-5">
                <SheetHeader>
                    <SheetTitle className="text-left text-ink">
                        {event.app_name}
                    </SheetTitle>
                </SheetHeader>

                <div className="mt-4 space-y-4">
                    {event.window_title && (
                        <p className="text-sm text-secondary-text line-clamp-2">
                            {event.window_title}
                        </p>
                    )}
                    {event.url && (
                        <p className="text-xs text-secondary-text truncate">
                            {event.url}
                        </p>
                    )}

                    <Separator />

                    <div className="grid grid-cols-2 gap-3 text-sm">
                        <div>
                            <p className="text-xs text-secondary-text">Start</p>
                            <p className="text-ink">
                                {dayjs(event.start_time).format("h:mm A")}
                            </p>
                        </div>
                        <div>
                            <p className="text-xs text-secondary-text">End</p>
                            <p className="text-ink">
                                {dayjs(event.end_time).format("h:mm A")}
                            </p>
                        </div>
                        <div>
                            <p className="text-xs text-secondary-text">
                                Duration
                            </p>
                            <p className="text-ink">
                                {formatDuration(event.duration_secs)}
                            </p>
                        </div>
                        <div>
                            <p className="text-xs text-secondary-text">
                                Device
                            </p>
                            <p className="text-ink">
                                {event.device?.name ?? "—"}
                            </p>
                        </div>
                    </div>

                    <Separator />

                    <div className="space-y-1.5">
                        <Label>Category</Label>
                        <Select
                            value={event.category_id ?? "__none__"}
                            onValueChange={handleCategoryChange}
                            disabled={saving}
                        >
                            <SelectTrigger>
                                <SelectValue placeholder="Uncategorized" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="__none__">
                                    Uncategorized
                                </SelectItem>
                                {categories.map((c) => (
                                    <SelectItem key={c.id} value={c.id}>
                                        <span className="flex items-center gap-2">
                                            <span
                                                className="h-2 w-2 rounded-full shrink-0"
                                                style={{ background: c.color }}
                                            />
                                            {c.name}
                                        </span>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="flex items-center justify-between">
                        <Label>Private</Label>
                        <Toggle
                            pressed={event.is_private}
                            onPressedChange={handlePrivacyToggle}
                            disabled={saving}
                            size="sm"
                        >
                            {event.is_private ? "Hidden" : "Visible"}
                        </Toggle>
                    </div>

                    <Separator />

                    <AlertDialog>
                        <AlertDialogTrigger asChild>
                            <Button
                                variant="destructive"
                                size="sm"
                                className="w-full"
                            >
                                Delete event
                            </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                            <AlertDialogHeader>
                                <AlertDialogTitle>
                                    Delete this event?
                                </AlertDialogTitle>
                                <AlertDialogDescription>
                                    This cannot be undone.
                                </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction onClick={handleDelete}>
                                    Delete
                                </AlertDialogAction>
                            </AlertDialogFooter>
                        </AlertDialogContent>
                    </AlertDialog>
                </div>
            </SheetContent>
        </Sheet>
    );
}
