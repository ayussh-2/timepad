import { BarChart2, Clock, LayoutDashboard, Settings } from "lucide-react";
import { NavLink } from "react-router";
import { cn } from "~/lib/utils";

const TABS = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard, end: true },
    { to: "/timeline", label: "Timeline", icon: Clock },
    { to: "/reports", label: "Reports", icon: BarChart2 },
    { to: "/devices", label: "Devices", icon: Settings },
];

export function MobileNav() {
    return (
        <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 flex border-t border-divider bg-surface">
            {TABS.map(({ to, label, icon: Icon, end }) => (
                <NavLink
                    key={to}
                    to={to}
                    end={end}
                    className={({ isActive }) =>
                        cn(
                            "flex flex-1 flex-col items-center gap-1 py-3 text-xs transition-colors",
                            isActive ? "text-accent" : "text-secondary-text",
                        )
                    }
                >
                    {({ isActive }) => (
                        <>
                            <Icon
                                className={cn(
                                    "h-5 w-5",
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
    );
}
