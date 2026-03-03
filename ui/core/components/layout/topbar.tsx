import dayjs from "dayjs";
import { LogOut, Settings, User } from "lucide-react";
import { useNavigate } from "react-router";
import { Avatar, AvatarFallback } from "~/components/ui/avatar";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { useAuthStore } from "~/store/auth.store";

export function Topbar() {
    const user = useAuthStore((s) => s.user);
    const logout = useAuthStore((s) => s.logout);
    const navigate = useNavigate();
    console.log(user);
    const initials =
        user?.display_name
            ?.split(" ")
            .map((n) => n[0])
            .join("")
            .toUpperCase()
            .slice(0, 2) ?? "?";

    return (
        <header className="flex items-center justify-between px-6 py-4 border-b border-divider bg-surface">
            <p className="text-sm text-secondary-text">
                {dayjs().format("dddd, MMMM D")}
            </p>

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <button className="rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent">
                        <Avatar className="h-8 w-8">
                            <AvatarFallback className="text-xs bg-accent/10 text-accent">
                                {initials}
                            </AvatarFallback>
                        </Avatar>
                    </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-44">
                    <div className="px-2 py-1.5">
                        <p className="text-sm font-medium text-ink capitalize">
                            {user?.display_name}
                        </p>
                        <p className="text-xs text-secondary-text truncate">
                            {user?.email}
                        </p>
                    </div>
                    <DropdownMenuSeparator />
                    {/* <DropdownMenuItem onClick={() => navigate("/settings")}>
                        <Settings className="h-4 w-4 mr-2" />
                        Settings
                    </DropdownMenuItem> */}
                    <DropdownMenuItem
                        onClick={() => {
                            logout();
                            navigate("/login");
                        }}
                        className="text-destructive focus:text-white"
                    >
                        <LogOut className="h-4 w-4 mr-2 focus:text-white" />
                        Sign out
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </header>
    );
}
