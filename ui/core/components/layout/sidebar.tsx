import {
    BarChart2,
    Clock,
    FolderOpen,
    LayoutDashboard,
    Monitor,
    Settings,
} from "lucide-react";
import { NavLink } from "react-router";
import { cn } from "~/lib/utils";

const NAV = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard, end: true },
    { to: "/timeline", label: "Timeline", icon: Clock },
    { to: "/reports", label: "Reports", icon: BarChart2 },
    { to: "/categories", label: "Categories", icon: FolderOpen },
    { to: "/devices", label: "Devices", icon: Monitor },
    // { to: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
    return (
        <aside className="hidden lg:flex flex-col w-56 shrink-0 h-screen sticky top-0 border-r border-divider bg-surface-alt px-3 py-6">
            <div className="mb-8 px-3">
                <span className="font-display text-2xl text-ink">Timepad</span>
            </div>

            <nav className="flex flex-col gap-0.5">
                {NAV.map(({ to, label, icon: Icon, end }) => (
                    <NavLink
                        key={to}
                        to={to}
                        end={end}
                        className={({ isActive }) =>
                            cn(
                                "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                                isActive
                                    ? "bg-accent/10 text-accent font-medium"
                                    : "text-secondary-text hover:text-ink hover:bg-paper",
                            )
                        }
                    >
                        {({ isActive }) => (
                            <>
                                <Icon
                                    className={cn(
                                        "h-4 w-4",
                                        isActive
                                            ? "text-accent"
                                            : "text-secondary-text",
                                    )}
                                />
                                {label}
                            </>
                        )}
                    </NavLink>
                ))}
            </nav>
        </aside>
    );
}
