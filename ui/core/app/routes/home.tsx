import type { Route } from "./+types/home";

export function meta({}: Route.MetaArgs) {
    return [
        { title: "Timepad" },
        { name: "description", content: "A personal sketchpad for your day" },
    ];
}

export default function Home() {
    return (
        <main className="flex min-h-screen items-center justify-center">
            <div className="text-center">
                <h1 className="font-display text-5xl text-ink">Timepad</h1>
                <p className="mt-3 text-secondary-text">
                    A personal sketchpad for your day
                </p>
            </div>
        </main>
    );
}
