import { create } from "zustand";
import { settingsApi } from "~/app/api/settings";
import type { UserSettings } from "~/app/types";

interface SettingsState {
    settings: UserSettings | null;
    isLoading: boolean;
    fetchSettings: () => Promise<void>;
    updateSettings: (
        payload: Partial<Omit<UserSettings, "user_id" | "updated_at">>,
    ) => Promise<void>;
}

export const useSettingsStore = create<SettingsState>((set) => ({
    settings: null,
    isLoading: false,

    fetchSettings: async () => {
        set({ isLoading: true });
        try {
            const data = await settingsApi.get();
            set({ settings: data, isLoading: false });
        } catch {
            set({ isLoading: false });
        }
    },

    updateSettings: async (payload) => {
        await settingsApi.update(payload);
        set((s) => ({
            settings: s.settings ? { ...s.settings, ...payload } : null,
        }));
    },
}));
