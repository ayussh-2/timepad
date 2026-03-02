import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { Copy, Monitor, Plus, Smartphone, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { devicesApi } from "~/app/api/devices";
import { EmptyState } from "~/components/ui/empty-state";
import { useNativeBridge } from "~/hooks/use-native-bridge";
import type { Device } from "~/app/types";
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
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Card, CardContent } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "~/components/ui/select";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
} from "~/components/ui/sheet";

dayjs.extend(relativeTime);

function PlatformIcon({ platform }: { platform: string }) {
    if (platform === "android") return <Smartphone className="h-5 w-5" />;
    if (platform === "browser")
        return <Monitor className="h-5 w-5 opacity-60" />;
    return <Monitor className="h-5 w-5" />;
}

interface RegisterSheetProps {
    open: boolean;
    onClose: () => void;
    onRegistered: (device: Device) => void;
}

function RegisterDeviceSheet({
    open,
    onClose,
    onRegistered,
}: RegisterSheetProps) {
    const bridge = useNativeBridge();
    const [name, setName] = useState("");
    const [platform, setPlatform] = useState<"android" | "windows" | "browser">(
        "windows",
    );
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");
    const [created, setCreated] = useState<Device | null>(null);

    useEffect(() => {
        if (open) {
            setName("");
            setPlatform(
                bridge.platform === "android"
                    ? "android"
                    : bridge.platform === "windows"
                      ? "windows"
                      : "windows",
            );
            setError("");
            setCreated(null);
        }
    }, [open, bridge.platform]);

    const handleRegister = async () => {
        if (!name.trim()) {
            setError("Name is required");
            return;
        }
        setSaving(true);
        setError("");
        try {
            const device = await devicesApi.register(name.trim(), platform);
            setCreated(device);
            onRegistered(device);
        } catch {
            setError("Registration failed.");
        } finally {
            setSaving(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
            <SheetContent side="right" className="w-full sm:max-w-md">
                <SheetHeader>
                    <SheetTitle>Register device</SheetTitle>
                </SheetHeader>

                <div className="mt-6 space-y-5">
                    {bridge.isNative && (
                        <div className="rounded-lg bg-accent/10 border border-accent/20 px-3 py-2 text-sm text-accent">
                            Running on {bridge.platform} — device key
                            pre-configured
                        </div>
                    )}

                    {!created ? (
                        <>
                            <div className="space-y-1.5">
                                <Label>Device name</Label>
                                <Input
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    placeholder="e.g. Work Laptop"
                                />
                            </div>
                            <div className="space-y-1.5">
                                <Label>Platform</Label>
                                <Select
                                    value={platform}
                                    onValueChange={(v) =>
                                        setPlatform(v as typeof platform)
                                    }
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="windows">
                                            Windows
                                        </SelectItem>
                                        <SelectItem value="android">
                                            Android
                                        </SelectItem>
                                        <SelectItem value="browser">
                                            Browser
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            {error && (
                                <p className="text-sm text-destructive">
                                    {error}
                                </p>
                            )}
                            <Button
                                onClick={handleRegister}
                                disabled={saving}
                                className="w-full"
                            >
                                {saving ? "Registering..." : "Register"}
                            </Button>
                        </>
                    ) : (
                        <div className="space-y-3">
                            <p className="text-sm text-secondary-text">
                                Device registered. Copy the key and paste it
                                into your collector config.
                            </p>
                            <div className="flex gap-2">
                                <code className="flex-1 rounded-lg bg-surface-alt border border-divider px-3 py-2 text-sm font-mono break-all">
                                    {created.device_key}
                                </code>
                                <Button
                                    variant="outline"
                                    size="icon"
                                    onClick={() =>
                                        navigator.clipboard.writeText(
                                            created.device_key,
                                        )
                                    }
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                            <Button
                                variant="outline"
                                onClick={onClose}
                                className="w-full"
                            >
                                Done
                            </Button>
                        </div>
                    )}
                </div>
            </SheetContent>
        </Sheet>
    );
}

export default function DevicesPage() {
    const [devices, setDevices] = useState<Device[]>([]);
    const [loading, setLoading] = useState(true);
    const [sheetOpen, setSheetOpen] = useState(false);

    useEffect(() => {
        devicesApi
            .list()
            .then(setDevices)
            .finally(() => setLoading(false));
    }, []);

    const handleDelete = async (id: string) => {
        await devicesApi.delete(id);
        setDevices((prev) => prev.filter((d) => d.id !== id));
    };

    return (
        <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-lg font-semibold text-ink">Devices</h1>
                <Button size="sm" onClick={() => setSheetOpen(true)}>
                    <Plus className="h-4 w-4 mr-1" /> Register
                </Button>
            </div>

            {loading ? (
                <div className="space-y-3">
                    {[...Array(2)].map((_, i) => (
                        <div
                            key={i}
                            className="h-20 rounded-xl bg-divider animate-pulse"
                        />
                    ))}
                </div>
            ) : devices.length === 0 ? (
                <EmptyState
                    title="No devices registered"
                    description="Register a device to start syncing activity data."
                >
                    <Button size="sm" onClick={() => setSheetOpen(true)}>
                        Register first device
                    </Button>
                </EmptyState>
            ) : (
                <div className="space-y-3">
                    {devices.map((device) => (
                        <Card key={device.id}>
                            <CardContent className="flex items-center gap-4 py-4">
                                <div className="text-secondary-text">
                                    <PlatformIcon platform={device.platform} />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <p className="text-sm font-medium text-ink">
                                        {device.name}
                                    </p>
                                    <p className="text-xs text-secondary-text capitalize">
                                        {device.platform}
                                        {device.last_seen_at && (
                                            <>
                                                {" "}
                                                · Last seen{" "}
                                                {dayjs(
                                                    device.last_seen_at,
                                                ).fromNow()}
                                            </>
                                        )}
                                    </p>
                                </div>
                                <AlertDialog>
                                    <AlertDialogTrigger asChild>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="shrink-0 text-destructive hover:text-destructive h-8 w-8"
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </AlertDialogTrigger>
                                    <AlertDialogContent>
                                        <AlertDialogHeader>
                                            <AlertDialogTitle>
                                                Remove &ldquo;{device.name}
                                                &rdquo;?
                                            </AlertDialogTitle>
                                            <AlertDialogDescription>
                                                All activity events from this
                                                device will be permanently
                                                deleted.
                                            </AlertDialogDescription>
                                        </AlertDialogHeader>
                                        <AlertDialogFooter>
                                            <AlertDialogCancel>
                                                Cancel
                                            </AlertDialogCancel>
                                            <AlertDialogAction
                                                onClick={() =>
                                                    handleDelete(device.id)
                                                }
                                            >
                                                Remove
                                            </AlertDialogAction>
                                        </AlertDialogFooter>
                                    </AlertDialogContent>
                                </AlertDialog>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            <RegisterDeviceSheet
                open={sheetOpen}
                onClose={() => setSheetOpen(false)}
                onRegistered={(device) =>
                    setDevices((prev) => [...prev, device])
                }
            />
        </div>
    );
}
