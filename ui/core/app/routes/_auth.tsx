import { Outlet } from "react-router";

export default function AuthLayout() {
    return (
        <div className="min-h-screen bg-paper flex flex-col items-center justify-center px-4">
            <div className="mb-8">
                <span className="font-display text-4xl text-ink">Timepad</span>
            </div>
            <div className="w-full max-w-sm bg-surface border border-divider rounded-2xl shadow-sm p-8">
                <Outlet />
            </div>
        </div>
    );
}
