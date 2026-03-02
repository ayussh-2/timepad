import { Plus, Trash2, Pencil } from "lucide-react";
import { useEffect, useState } from "react";
import { categoriesApi } from "~/app/api/categories";
import { EmptyState } from "~/components/ui/empty-state";
import type { Category, CategoryRule } from "~/app/types";
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
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
} from "~/components/ui/sheet";

const PRESET_COLORS = [
    "#5b7c99",
    "#7a9a6d",
    "#c4a77d",
    "#8b7b9e",
    "#a3796a",
    "#6e6a63",
    "#4a7c7c",
    "#9a6d6d",
];

const PRODUCTIVE_OPTIONS = [
    { value: "true", label: "Productive" },
    { value: "false", label: "Distraction" },
    { value: "null", label: "Neutral" },
];

function productiveLabel(v: boolean | null) {
    if (v === true) return { label: "Productive", color: "#7a9a6d" };
    if (v === false) return { label: "Distraction", color: "#c4a77d" };
    return { label: "Neutral", color: "#6e6a63" };
}

interface CategoryFormSheetProps {
    open: boolean;
    onClose: () => void;
    onSave: (cat: Category) => void;
    editing: Category | null;
}

function CategoryFormSheet({
    open,
    onClose,
    onSave,
    editing,
}: CategoryFormSheetProps) {
    const [name, setName] = useState(editing?.name ?? "");
    const [color, setColor] = useState(editing?.color ?? PRESET_COLORS[0]);
    const [icon, setIcon] = useState(editing?.icon ?? "");
    const [productive, setProductive] = useState<string>(
        editing?.is_productive === true
            ? "true"
            : editing?.is_productive === false
              ? "false"
              : "null",
    );
    const [rules, setRules] = useState<CategoryRule[]>(editing?.rules ?? []);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        if (open) {
            setName(editing?.name ?? "");
            setColor(editing?.color ?? PRESET_COLORS[0]);
            setIcon(editing?.icon ?? "");
            setProductive(
                editing?.is_productive === true
                    ? "true"
                    : editing?.is_productive === false
                      ? "false"
                      : "null",
            );
            setRules(editing?.rules ?? []);
            setError("");
        }
    }, [open, editing]);

    const addRule = () =>
        setRules([...rules, { type: "app_name", op: "contains", value: "" }]);
    const removeRule = (i: number) =>
        setRules(rules.filter((_, idx) => idx !== i));
    const updateRule = (i: number, patch: Partial<CategoryRule>) =>
        setRules(rules.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));

    const handleSave = async () => {
        if (!name.trim()) {
            setError("Name is required");
            return;
        }
        setSaving(true);
        setError("");
        try {
            const productiveVal =
                productive === "true"
                    ? true
                    : productive === "false"
                      ? false
                      : null;
            const payload = {
                name: name.trim(),
                color,
                icon,
                is_productive: productiveVal,
                rules,
            };
            let cat: Category;
            if (editing) {
                await categoriesApi.update(editing.id, payload);
                cat = { ...editing, ...payload };
            } else {
                cat = await categoriesApi.create(payload);
            }
            onSave(cat);
            onClose();
        } catch {
            setError("Failed to save.");
        } finally {
            setSaving(false);
        }
    };

    return (
        <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
            <SheetContent
                side="right"
                className="w-full sm:max-w-md overflow-y-auto px-5"
            >
                <SheetHeader>
                    <SheetTitle>
                        {editing ? "Edit category" : "New category"}
                    </SheetTitle>
                </SheetHeader>

                <div className="mt-6 space-y-5">
                    <div className="space-y-1.5">
                        <Label>Name</Label>
                        <Input
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                        />
                    </div>

                    <div className="space-y-1.5">
                        <Label>Color</Label>
                        <div className="flex gap-2 flex-wrap">
                            {PRESET_COLORS.map((c) => (
                                <button
                                    key={c}
                                    type="button"
                                    onClick={() => setColor(c)}
                                    className="h-7 w-7 rounded-full border-2 transition-all"
                                    style={{
                                        background: c,
                                        borderColor:
                                            color === c
                                                ? "#2b2b2b"
                                                : "transparent",
                                    }}
                                />
                            ))}
                            <input
                                type="color"
                                value={color}
                                onChange={(e) => setColor(e.target.value)}
                                className="h-7 w-7 rounded-full cursor-pointer border border-divider"
                            />
                        </div>
                    </div>

                    <div className="space-y-1.5">
                        <Label>Type</Label>
                        <Select
                            value={productive}
                            onValueChange={setProductive}
                        >
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {PRODUCTIVE_OPTIONS.map((o) => (
                                    <SelectItem key={o.value} value={o.value}>
                                        {o.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <Separator />

                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <Label>Auto-match rules</Label>
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={addRule}
                            >
                                <Plus className="h-4 w-4 mr-1" /> Add rule
                            </Button>
                        </div>
                        {rules.map((rule, i) => (
                            <div key={i} className="flex gap-2 items-start">
                                <Select
                                    value={rule.type}
                                    onValueChange={(v) =>
                                        updateRule(i, {
                                            type: v as CategoryRule["type"],
                                        })
                                    }
                                >
                                    <SelectTrigger className="w-32">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="app_name">
                                            App name
                                        </SelectItem>
                                        <SelectItem value="url_domain">
                                            URL domain
                                        </SelectItem>
                                        <SelectItem value="window_title">
                                            Window title
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                                <Select
                                    value={rule.op}
                                    onValueChange={(v) =>
                                        updateRule(i, {
                                            op: v as CategoryRule["op"],
                                        })
                                    }
                                >
                                    <SelectTrigger className="w-28">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="contains">
                                            contains
                                        </SelectItem>
                                        <SelectItem value="equals">
                                            equals
                                        </SelectItem>
                                        <SelectItem value="starts_with">
                                            starts with
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                                <Input
                                    className="flex-1"
                                    value={rule.value}
                                    onChange={(e) =>
                                        updateRule(i, { value: e.target.value })
                                    }
                                    placeholder="value"
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => removeRule(i)}
                                    className="shrink-0"
                                >
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </div>
                        ))}
                    </div>

                    {error && (
                        <p className="text-sm text-destructive">{error}</p>
                    )}

                    <Button
                        onClick={handleSave}
                        disabled={saving}
                        className="w-full"
                    >
                        {saving ? "Saving..." : "Save"}
                    </Button>
                </div>
            </SheetContent>
        </Sheet>
    );
}

export default function CategoriesPage() {
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [sheetOpen, setSheetOpen] = useState(false);
    const [editing, setEditing] = useState<Category | null>(null);

    useEffect(() => {
        categoriesApi
            .list()
            .then(setCategories)
            .finally(() => setLoading(false));
    }, []);

    const system = categories.filter((c) => c.is_system);
    const user = categories.filter((c) => !c.is_system);

    const handleSave = (cat: Category) => {
        setCategories((prev) => {
            const idx = prev.findIndex((c) => c.id === cat.id);
            if (idx >= 0) {
                const next = [...prev];
                next[idx] = cat;
                return next;
            }
            return [...prev, cat];
        });
    };

    const handleDelete = async (id: string) => {
        await categoriesApi.delete(id);
        setCategories((prev) => prev.filter((c) => c.id !== id));
    };

    const openNew = () => {
        setEditing(null);
        setSheetOpen(true);
    };
    const openEdit = (cat: Category) => {
        setEditing(cat);
        setSheetOpen(true);
    };

    function CategoryRow({
        cat,
        editable,
    }: {
        cat: Category;
        editable: boolean;
    }) {
        const { label, color } = productiveLabel(cat.is_productive);
        return (
            <div className="flex items-center gap-3 py-3">
                <span
                    className="h-3 w-3 rounded-full shrink-0"
                    style={{ background: cat.color }}
                />
                <span className="flex-1 text-sm text-ink">{cat.name}</span>
                <Badge
                    variant="secondary"
                    className="text-xs"
                    style={{ color }}
                >
                    {label}
                </Badge>
                {editable && (
                    <div className="flex gap-1">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => openEdit(cat)}
                        >
                            <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        <AlertDialog>
                            <AlertDialogTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-7 w-7 text-destructive hover:text-destructive"
                                >
                                    <Trash2 className="h-3.5 w-3.5" />
                                </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                                <AlertDialogHeader>
                                    <AlertDialogTitle>
                                        Delete &ldquo;{cat.name}&rdquo;?
                                    </AlertDialogTitle>
                                    <AlertDialogDescription>
                                        All events in this category will become
                                        uncategorized.
                                    </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                    <AlertDialogCancel>
                                        Cancel
                                    </AlertDialogCancel>
                                    <AlertDialogAction
                                        onClick={() => handleDelete(cat.id)}
                                    >
                                        Delete
                                    </AlertDialogAction>
                                </AlertDialogFooter>
                            </AlertDialogContent>
                        </AlertDialog>
                    </div>
                )}
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto px-4 py-6 space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-lg font-semibold text-ink">Categories</h1>
                <Button size="sm" onClick={openNew}>
                    <Plus className="h-4 w-4 mr-1" /> New
                </Button>
            </div>

            {loading ? (
                <div className="space-y-2">
                    {[...Array(4)].map((_, i) => (
                        <div
                            key={i}
                            className="h-11 rounded-lg bg-divider animate-pulse"
                        />
                    ))}
                </div>
            ) : (
                <>
                    {user.length > 0 && (
                        <Card>
                            <CardHeader className="pb-0">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    Your categories
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="divide-y divide-divider">
                                {user.map((cat) => (
                                    <CategoryRow
                                        key={cat.id}
                                        cat={cat}
                                        editable
                                    />
                                ))}
                            </CardContent>
                        </Card>
                    )}

                    {user.length === 0 && (
                        <EmptyState
                            title="No categories yet"
                            description="Create categories to automatically group your activity."
                        />
                    )}

                    {system.length > 0 && (
                        <Card>
                            <CardHeader className="pb-0">
                                <CardTitle className="text-sm font-medium text-secondary-text">
                                    System categories
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="divide-y divide-divider">
                                {system.map((cat) => (
                                    <CategoryRow
                                        key={cat.id}
                                        cat={cat}
                                        editable={false}
                                    />
                                ))}
                            </CardContent>
                        </Card>
                    )}
                </>
            )}

            <CategoryFormSheet
                open={sheetOpen}
                onClose={() => setSheetOpen(false)}
                onSave={handleSave}
                editing={editing}
            />
        </div>
    );
}
