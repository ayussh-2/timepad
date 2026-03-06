import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import type { Plugin } from "vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

// Intercept browser-automatic & devtools requests before React Router SSR handler sees them
const bypassStaticRequests: Plugin = {
    name: "bypass-static-requests",
    configureServer(server) {
        server.middlewares.use((req, res, next) => {
            if (
                req.url === "/favicon.ico" ||
                req.url?.startsWith("/.well-known/")
            ) {
                res.statusCode = 204;
                res.end();
                return;
            }
            next();
        });
    },
};

export default defineConfig({
    plugins: [
        bypassStaticRequests,
        tailwindcss(),
        reactRouter(),
        tsconfigPaths(),
    ],
    server: {
        allowedHosts: true,
        proxy: {
            "/api": {
                target: "http://localhost:8080",
                changeOrigin: true,
            },
        },
    },
});
