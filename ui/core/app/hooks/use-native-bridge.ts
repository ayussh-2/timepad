interface NativeBridge {
    getDeviceKey: () => string;
    getPlatform: () => string;
}

export function useNativeBridge() {
    const bridge =
        typeof window !== "undefined"
            ? (window as unknown as { TimePadBridge?: NativeBridge })
                  .TimePadBridge
            : undefined;
    return {
        isNative: !!bridge,
        deviceKey: bridge?.getDeviceKey() ?? null,
        platform: (bridge?.getPlatform() ?? "web") as
            | "android"
            | "windows"
            | "web",
    };
}
