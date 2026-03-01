import {
    type RouteConfig,
    index,
    layout,
    route,
} from "@react-router/dev/routes";

export default [
    layout("routes/_auth.tsx", [
        route("login", "routes/_auth.login.tsx"),
        route("register", "routes/_auth.register.tsx"),
    ]),
    layout("routes/_app.tsx", [
        index("routes/_app.dashboard.tsx"),
        route("timeline", "routes/_app.timeline.tsx"),
        route("reports", "routes/_app.reports.tsx"),
        route("categories", "routes/_app.categories.tsx"),
        route("devices", "routes/_app.devices.tsx"),
        route("settings", "routes/_app.settings.tsx"),
    ]),
] satisfies RouteConfig;
