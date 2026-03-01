import type { ApiEnvelope, Device } from "~/app/types";
import { client } from "./client";

interface RawDevice {
    ID: string;
    UserID: string;
    Name: string;
    Platform: string;
    DeviceKey: string;
    LastSeenAt: string | null;
    CreatedAt: string;
}

function normalizeDevice(d: RawDevice): Device {
    return {
        id: d.ID,
        user_id: d.UserID,
        name: d.Name,
        platform: d.Platform as Device["platform"],
        device_key: d.DeviceKey,
        last_seen_at: d.LastSeenAt,
        created_at: d.CreatedAt,
    };
}

export const devicesApi = {
    list: () =>
        client
            .get<ApiEnvelope<RawDevice[]>>("/devices")
            .then((r) => (r.data.data ?? []).map(normalizeDevice)),

    register: (name: string, platform: "android" | "windows" | "browser") =>
        client
            .post<ApiEnvelope<RawDevice>>("/devices", { name, platform })
            .then((r) => normalizeDevice(r.data.data)),

    delete: (id: string) =>
        client.delete<ApiEnvelope<null>>(`/devices/${id}`).then((r) => r.data),
};
