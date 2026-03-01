export type ApiEnvelope<T> = {
    success: boolean;
    message: string;
    data: T;
    timestamp: number;
};

export interface User {
    id: string;
    email: string;
    display_name: string;
    timezone: string;
    created_at: string;
}

export interface AuthResponse {
    user: User;
    access_token: string;
    refresh_token: string;
    expires_in: number;
}

export interface Device {
    id: string;
    user_id: string;
    name: string;
    platform: "android" | "windows" | "browser";
    device_key: string;
    last_seen_at: string | null;
    created_at: string;
}

export interface Category {
    id: string;
    user_id: string | null;
    name: string;
    color: string;
    icon: string;
    is_system: boolean;
    is_productive: boolean | null;
    rules: CategoryRule[];
}

export interface CategoryRule {
    type: "app_name" | "url_domain" | "window_title";
    op: "contains" | "equals" | "starts_with";
    value: string;
}

export interface TimelineEntry {
    id: string;
    user_id: string;
    device_id: string;
    app_name: string;
    window_title: string;
    url: string;
    category_id: string | null;
    category: Category | null;
    device: Device | null;
    start_time: string;
    end_time: string;
    duration_secs: number;
    is_idle: boolean;
    is_private: boolean;
}

export interface TimelineResponse {
    events: TimelineEntry[];
    next_cursor: string | null;
}

export interface AppUsage {
    app_name: string;
    total_secs: number;
    category?: Category | null;
}

export interface DeviceUsage {
    device_name: string;
    platform: string;
    total_secs: number;
}

export interface DailySummary {
    date: string;
    total_active_secs: number;
    total_idle_secs: number;
    productive_secs: number;
    distraction_secs: number;
    peak_hour: number;
    top_apps: AppUsage[];
    device_breakdown: DeviceUsage[];
}

export interface WeeklySummary {
    start_date: string;
    end_date: string;
    total_active_secs: number;
    total_idle_secs: number;
    productive_secs: number;
    distraction_secs: number;
    daily_breakdown: DailySummary[];
}

export interface ReportData {
    total_active_secs: number;
    total_idle_secs: number;
    category_usage: Record<string, number>;
    app_usage: Record<string, number>;
    device_usage: Record<string, number>;
    daily_active_trend: Record<string, number>;
}

export interface UserSettings {
    user_id: string;
    excluded_apps: string[];
    excluded_urls: string[];
    idle_threshold: number;
    tracking_enabled: boolean;
    data_retention_days: number;
    updated_at: string;
}

export interface PaginatedEvents {
    events: TimelineEntry[];
    next_cursor: string | null;
}
