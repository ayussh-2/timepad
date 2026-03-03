import axios from "axios";
import type { AxiosInstance, InternalAxiosRequestConfig } from "axios";

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

export const client: AxiosInstance = axios.create({ baseURL: BASE_URL });

// attach JWT on every request
client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const raw = localStorage.getItem("auth-store");
    if (raw) {
        try {
            const { state } = JSON.parse(raw) as {
                state: { accessToken: string };
            };
            if (state?.accessToken) {
                config.headers.Authorization = `Bearer ${state.accessToken}`;
            }
        } catch {
            // malformed storage — ignore
        }
    }
    return config;
});

let refreshing = false;
let queue: Array<{
    resolve: (token: string) => void;
    reject: (err: unknown) => void;
}> = [];

client.interceptors.response.use(
    (res) => res,
    async (err) => {
        const original = err.config as InternalAxiosRequestConfig & {
            _retry?: boolean;
        };
        // Never retry auth endpoints — a 401 there is just wrong credentials
        const isAuthEndpoint = original.url?.includes("/auth/");
        if (err.response?.status !== 401 || original._retry || isAuthEndpoint) {
            return Promise.reject(err);
        }
        original._retry = true;

        if (refreshing) {
            return new Promise<string>((resolve, reject) => {
                queue.push({ resolve, reject });
            }).then((token) => {
                original.headers.Authorization = `Bearer ${token}`;
                return client(original);
            });
        }

        refreshing = true;
        try {
            const raw = localStorage.getItem("auth-store");
            const { state } = JSON.parse(raw ?? "{}") as {
                state: { refreshToken: string };
            };
            const { data } = await axios.post<{
                success: boolean;
                data: { AccessToken: string; RefreshToken: string };
            }>(`${BASE_URL}/auth/refresh`, {
                refresh_token: state.refreshToken,
            });

            const accessToken = data.data.AccessToken;
            const refreshToken = data.data.RefreshToken;

            // update stored tokens
            const stored = JSON.parse(
                localStorage.getItem("auth-store") ?? "{}",
            );
            stored.state.accessToken = accessToken;
            stored.state.refreshToken = refreshToken;
            localStorage.setItem("auth-store", JSON.stringify(stored));

            queue.forEach(({ resolve }) => resolve(accessToken));
            queue = [];
            original.headers.Authorization = `Bearer ${accessToken}`;
            return client(original);
        } catch (refreshErr: unknown) {
            queue.forEach(({ reject }) => reject(refreshErr));
            queue = [];
            // Only force-logout when the refresh token itself is rejected (truly
            // expired / revoked). Network errors or 5xx should NOT wipe the
            // session — the user is still authenticated, just temporarily offline.
            const status = (refreshErr as any)?.response?.status;
            if (status === 401) {
                localStorage.removeItem("auth-store");
                window.location.href = "/login";
            }
            return Promise.reject(refreshErr);
        } finally {
            refreshing = false;
        }
    },
);
