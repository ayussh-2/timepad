import { useEffect } from "react";
import { Outlet, redirect, useNavigate } from "react-router";
import { MobileNav } from "~/app/components/layout/mobile-nav";
import { Sidebar } from "~/app/components/layout/sidebar";
import { Topbar } from "~/app/components/layout/topbar";
import { useAutoRefresh } from "~/app/hooks/use-auto-refresh";
import { useAuthStore } from "~/app/store/auth.store";
import { TooltipProvider } from "~/components/ui/tooltip";

export function clientLoader() {
    const raw = localStorage.getItem("auth-store");
    if (!raw) return redirect("/login");
    try {
        const { state } = JSON.parse(raw) as {
            state: { accessToken: string | null };
        };
        if (!state?.accessToken) return redirect("/login");
    } catch {
        return redirect("/login");
    }
    return null;
}

function AutoRefreshProvider() {
    useAutoRefresh();
    return null;
}

export default function AppLayout() {
    const accessToken = useAuthStore((s) => s.accessToken);
    const navigate = useNavigate();

    useEffect(() => {
        if (!accessToken) navigate("/login");
    }, [accessToken, navigate]);

    if (!accessToken) return null;

    return (
        <TooltipProvider>
            <AutoRefreshProvider />
            <div className="flex min-h-screen bg-paper">
                <Sidebar />
                <div className="flex flex-1 flex-col min-w-0">
                    <Topbar />
                    <main className="flex-1 overflow-y-auto pb-20 lg:pb-0">
                        <Outlet />
                    </main>
                </div>
            </div>
            <MobileNav />
        </TooltipProvider>
    );
}
