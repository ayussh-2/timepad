import type { ApiEnvelope, Device } from "~/app/types";
import { client } from "./client";

interface RawDevice {
    id: string;
    user_id: string;
    name: string;
    platform: string;
    device_key: string;
    last_seen_at: string | null;
    created_at: string;
}

function normalizeDevice(d: RawDevice): Device {
    return {
        id: d.id,
        user_id: d.user_id,
        name: d.name,
        platform: d.platform as Device["platform"],
        device_key: d.device_key,
        last_seen_at: d.last_seen_at,
        created_at: d.created_at,
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
