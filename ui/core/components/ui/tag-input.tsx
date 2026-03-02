import { X } from "lucide-react";
import { useRef, useState } from "react";
import { cn } from "~/lib/utils";
import { Badge } from "~/components/ui/badge";
import { Input } from "~/components/ui/input";

interface TagInputProps {
    value: string[];
    onChange: (value: string[]) => void;
    placeholder?: string;
    className?: string;
}

export function TagInput({
    value,
    onChange,
    placeholder,
    className,
}: TagInputProps) {
    const [input, setInput] = useState("");
    const inputRef = useRef<HTMLInputElement>(null);

    const add = () => {
        const trimmed = input.trim();
        if (trimmed && !value.includes(trimmed)) {
            onChange([...value, trimmed]);
        }
        setInput("");
    };

    const remove = (tag: string) => onChange(value.filter((t) => t !== tag));

    return (
        <div
            className={cn(
                "flex flex-wrap gap-1.5 rounded-lg border border-divider bg-surface p-2 min-h-10 cursor-text",
                className,
            )}
            onClick={() => inputRef.current?.focus()}
        >
            {value.map((tag) => (
                <Badge key={tag} variant="secondary" className="gap-1 pr-1">
                    {tag}
                    <button
                        type="button"
                        onClick={(e) => {
                            e.stopPropagation();
                            remove(tag);
                        }}
                        className="ml-0.5 rounded-sm hover:text-ink"
                    >
                        <X className="h-3 w-3" />
                    </button>
                </Badge>
            ))}
            <Input
                ref={inputRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === ",") {
                        e.preventDefault();
                        add();
                    }
                    if (e.key === "Backspace" && !input && value.length > 0) {
                        onChange(value.slice(0, -1));
                    }
                }}
                onBlur={add}
                placeholder={value.length === 0 ? placeholder : undefined}
                className="h-auto flex-1 min-w-24 border-none shadow-none focus-visible:ring-0 p-0 text-sm"
            />
        </div>
    );
}
