import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { TagInput } from "~/components/ui/tag-input";
import { useAuthStore } from "~/store/auth.store";
import { useSettingsStore } from "~/store/settings.store";
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
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "~/components/ui/select";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { Slider } from "~/components/ui/slider";
import { Toggle } from "~/components/ui/toggle";

const TIMEZONES = Intl.supportedValuesOf("timeZone");
const RETENTION_OPTIONS = [
    { value: "30", label: "30 days" },
    { value: "90", label: "90 days" },
    { value: "180", label: "180 days" },
    { value: "365", label: "1 year" },
    { value: "0", label: "Forever" },
];

export default function SettingsPage() {
    const { settings, isLoading, fetchSettings, updateSettings } =
        useSettingsStore();
    const deleteAccount = useAuthStore((s) => s.deleteAccount);
    const logout = useAuthStore((s) => s.logout);
    const navigate = useNavigate();

    const [confirmEmail, setConfirmEmail] = useState("");
    const userEmail = useAuthStore((s) => s.user?.email ?? "");

    useEffect(() => {
        fetchSettings();
    }, [fetchSettings]);

    if (isLoading || !settings) {
        return (
            <div className="max-w-2xl mx-auto px-4 py-6 space-y-4">
                {[...Array(5)].map((_, i) => (
                    <Skeleton key={i} className="h-24 w-full rounded-xl" />
                ))}
            </div>
        );
    }

    const save = (patch: Partial<typeof settings>) => updateSettings(patch);

    return (
        <div className="max-w-2xl mx-auto px-4 py-6 space-y-4">
            <h1 className="text-lg font-semibold text-ink">Settings</h1>

            {/* Tracking toggle */}
            <Card>
                <CardContent className="flex items-center justify-between py-5">
                    <div>
                        <p className="text-sm font-medium text-ink">
                            Activity tracking
                        </p>
                        <p className="text-xs text-secondary-text mt-0.5">
                            {settings.tracking_enabled
                                ? "Tracking is active"
                                : "Tracking is paused — no data is being collected"}
                        </p>
                    </div>
                    <Toggle
                        pressed={settings.tracking_enabled}
                        onPressedChange={(v) => save({ tracking_enabled: v })}
                        className="data-[state=on]:bg-accent data-[state=on]:text-white"
                    >
                        {settings.tracking_enabled ? "On" : "Off"}
                    </Toggle>
                </CardContent>
            </Card>

            {/* Idle threshold */}
            <Card>
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-secondary-text">
                        Idle detection
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                    <p className="text-sm text-ink">
                        Mark idle after{" "}
                        <strong>
                            {Math.round(settings.idle_threshold / 60)} minutes
                        </strong>{" "}
                        of inactivity
                    </p>
                    <Slider
                        min={30}
                        max={600}
                        step={30}
                        value={[settings.idle_threshold]}
                        onValueChange={([v]) => save({ idle_threshold: v })}
                        className="max-w-xs"
                    />
                </CardContent>
            </Card>

            {/* Excluded apps */}
            <Card>
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-secondary-text">
                        Excluded apps
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <TagInput
                        value={settings.excluded_apps}
                        onChange={(v) => save({ excluded_apps: v })}
                        placeholder="Type app name and press Enter"
                    />
                    <p className="mt-2 text-xs text-secondary-text">
                        Activity from these apps is silently ignored during
                        ingestion.
                    </p>
                </CardContent>
            </Card>

            {/* Excluded URLs */}
            <Card>
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-secondary-text">
                        Excluded URLs
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <TagInput
                        value={settings.excluded_urls}
                        onChange={(v) => save({ excluded_urls: v })}
                        placeholder="Type URL and press Enter"
                    />
                </CardContent>
            </Card>

            {/* Data retention */}
            <Card>
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-secondary-text">
                        Data retention
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <Select
                        value={String(settings.data_retention_days)}
                        onValueChange={(v) =>
                            save({ data_retention_days: Number(v) })
                        }
                    >
                        <SelectTrigger className="w-40">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {RETENTION_OPTIONS.map((o) => (
                                <SelectItem key={o.value} value={o.value}>
                                    {o.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <p className="mt-2 text-xs text-secondary-text">
                        Events older than this are automatically removed.
                    </p>
                </CardContent>
            </Card>

            {/* Danger zone */}
            <Separator />
            <Card className="border-destructive/30">
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-destructive">
                        Danger zone
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="text-sm text-ink">Sign out</p>
                            <p className="text-xs text-secondary-text">
                                Sign out of this device
                            </p>
                        </div>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                                logout();
                                navigate("/login");
                            }}
                        >
                            Sign out
                        </Button>
                    </div>
                    <Separator />
                    <div className="space-y-3">
                        <div>
                            <p className="text-sm text-ink">Delete account</p>
                            <p className="text-xs text-secondary-text">
                                Permanently deletes your account and all data.
                                This cannot be undone.
                            </p>
                        </div>
                        <AlertDialog>
                            <AlertDialogTrigger asChild>
                                <Button variant="destructive" size="sm">
                                    Delete account
                                </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                                <AlertDialogHeader>
                                    <AlertDialogTitle>
                                        Delete your account?
                                    </AlertDialogTitle>
                                    <AlertDialogDescription>
                                        All your devices, events, categories,
                                        and settings will be permanently
                                        deleted. Type your email to confirm.
                                    </AlertDialogDescription>
                                </AlertDialogHeader>
                                <Input
                                    value={confirmEmail}
                                    onChange={(e) =>
                                        setConfirmEmail(e.target.value)
                                    }
                                    placeholder={userEmail}
                                    className="mt-2"
                                />
                                <AlertDialogFooter>
                                    <AlertDialogCancel
                                        onClick={() => setConfirmEmail("")}
                                    >
                                        Cancel
                                    </AlertDialogCancel>
                                    <AlertDialogAction
                                        disabled={confirmEmail !== userEmail}
                                        onClick={async () => {
                                            await deleteAccount();
                                            navigate("/login");
                                        }}
                                        className="bg-destructive hover:bg-destructive/90"
                                    >
                                        Delete permanently
                                    </AlertDialogAction>
                                </AlertDialogFooter>
                            </AlertDialogContent>
                        </AlertDialog>
                    </div>
                </CardContent>
            </Card>

            <p className="text-center text-xs text-secondary-text pb-4">
                {import.meta.env.VITE_APP_VERSION ?? "dev"}
            </p>
        </div>
    );
}
