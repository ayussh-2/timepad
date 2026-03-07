import {
    isRouteErrorResponse,
    Links,
    Meta,
    Outlet,
    Scripts,
    ScrollRestoration,
} from "react-router";

import type { Route } from "./+types/root";
import "./app.css";

export function Layout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="en">
            <head>
                <meta charSet="utf-8" />
                <meta
                    name="viewport"
                    content="width=device-width, initial-scale=1"
                />
                <link rel="preconnect" href="https://fonts.googleapis.com" />
                <link
                    rel="preconnect"
                    href="https://fonts.gstatic.com"
                    crossOrigin="anonymous"
                />
                <link
                    href="https://fonts.googleapis.com/css2?family=Geist:wght@100..900&display=swap"
                    rel="stylesheet"
                />
                <title>Timepad</title>
                <meta
                    name="description"
                    content="A personal sketchpad for your day"
                />
                <link rel="icon" type="image/x-icon" href="/favicon.ico" />
                <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
                <link
                    rel="icon"
                    type="image/png"
                    sizes="96x96"
                    href="/favicon-96x96.png"
                />
                <link
                    rel="apple-touch-icon"
                    sizes="180x180"
                    href="/apple-touch-icon.png"
                />
                <link rel="manifest" href="/site.webmanifest" />
                <Meta />
                <Links />
            </head>
            <body>
                {children}
                <ScrollRestoration />
                <Scripts />
            </body>
        </html>
    );
}

export default function App() {
    return <Outlet />;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
    let message = "Oops!";
    let details = "An unexpected error occurred.";
    let stack: string | undefined;

    if (isRouteErrorResponse(error)) {
        message = error.status === 404 ? "404" : "Error";
        details =
            error.status === 404
                ? "The requested page could not be found."
                : error.statusText || details;
    } else if (import.meta.env.DEV && error && error instanceof Error) {
        details = error.message;
        stack = error.stack;
    }

    return (
        <main className="flex min-h-screen items-center justify-center p-4">
            <div className="text-center">
                <h1 className="font-display text-4xl text-ink">{message}</h1>
                <p className="mt-2 text-secondary-text">{details}</p>
                {stack && (
                    <pre className="mt-4 max-w-2xl overflow-x-auto rounded-lg bg-surface-alt p-4 text-left text-sm">
                        <code>{stack}</code>
                    </pre>
                )}
            </div>
        </main>
    );
}
