import { create } from "zustand";
import { persist } from "zustand/middleware";
import { authApi } from "~/app/api/auth";
import type { User } from "~/app/types";

interface AuthState {
    user: User | null;
    accessToken: string | null;
    refreshToken: string | null;
    login: (email: string, password: string) => Promise<void>;
    register: (
        email: string,
        password: string,
        displayName: string,
    ) => Promise<void>;
    logout: () => void;
    deleteAccount: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            user: null,
            accessToken: null,
            refreshToken: null,

            login: async (email, password) => {
                const data = await authApi.login(email, password);
                set({
                    user: data.user,
                    accessToken: data.access_token,
                    refreshToken: data.refresh_token,
                });
                (window as any).timePadSaveConfig?.(
                    data.access_token,
                    data.refresh_token,
                    "",
                );
            },

            register: async (email, password, displayName) => {
                const data = await authApi.register(
                    email,
                    password,
                    displayName,
                );
                set({
                    user: data.user,
                    accessToken: data.access_token,
                    refreshToken: data.refresh_token,
                });
                (window as any).timePadSaveConfig?.(
                    data.access_token,
                    data.refresh_token,
                    "",
                );
            },

            logout: () =>
                set({ user: null, accessToken: null, refreshToken: null }),

            deleteAccount: async () => {
                await authApi.deleteAccount();
                set({ user: null, accessToken: null, refreshToken: null });
            },
        }),
        { name: "auth-store" },
    ),
);
