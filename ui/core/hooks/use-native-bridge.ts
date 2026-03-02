interface NativeBridge {
    getDeviceKey: () => string;
    getPlatform: () => string;
}

type SaveConfigFn = (
    accessToken: string,
    refreshToken: string,
    deviceKey: string,
) => void;

export function useNativeBridge() {
    const win =
        typeof window !== "undefined"
            ? (window as unknown as {
                  TimePadBridge?: NativeBridge;
                  timePadSaveConfig?: SaveConfigFn;
              })
            : undefined;
    const bridge = win?.TimePadBridge;
    return {
        isNative: !!bridge,
        deviceKey: bridge?.getDeviceKey() ?? null,
        platform: (bridge?.getPlatform() ?? "web") as
            | "android"
            | "windows"
            | "web",
        saveConfig: (
            accessToken: string,
            refreshToken: string,
            deviceKey: string,
        ) => {
            win?.timePadSaveConfig?.(accessToken, refreshToken, deviceKey);
        },
    };
}
