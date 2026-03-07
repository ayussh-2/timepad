import type { AuthResponse } from "~/app/types";
import { client } from "./client";

// Actual shape returned by the Go server
interface RawAuthPayload {
    UserId: string;
    Name: string;
    Email?: string;
    AccessToken: string;
    RefreshToken: string;
}
interface RawAuthResponse {
    success: boolean;
    data: RawAuthPayload;
}

function normalize(raw: RawAuthPayload): AuthResponse {
    return {
        user: {
            id: raw.UserId,
            display_name: raw.Name,
            email: raw.Email ?? "",
            timezone: "",
            created_at: "",
        },
        access_token: raw.AccessToken,
        refresh_token: raw.RefreshToken,
        expires_in: 0,
    };
}

export const authApi = {
    register: (email: string, password: string, display_name: string) =>
        client
            .post<RawAuthResponse>("/auth/register", {
                email,
                password,
                name: display_name,
            })
            .then((r) => normalize(r.data.data)),

    login: (email: string, password: string) =>
        client
            .post<RawAuthResponse>("/auth/login", { email, password })
            .then((r) => normalize(r.data.data)),

    refresh: (refresh_token: string) =>
        client
            .post<{
                success: boolean;
                data: { AccessToken: string; RefreshToken: string };
            }>("/auth/refresh", { refresh_token })
            .then((r) => ({
                access_token: r.data.data.AccessToken,
                refresh_token: r.data.data.RefreshToken,
            })),

    deleteAccount: () =>
        client.delete<{ message: string }>("/auth/account").then((r) => r.data),
};
