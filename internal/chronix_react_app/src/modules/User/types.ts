// Shared types for the end-user module

import type {TimeFormat} from "../../data/SettingsContext.tsx";

export interface UserProfileData {
    id: string;
    displayName: string;
    email: string;
    phone?: string; // optional phone number for SMS notifications
    timeFormat: TimeFormat;
    timeZone: string; // IANA TZ name
    created_at?: string;
    last_login_at?: string;
}

export interface UserActivityItem {
    id: string;
    when: string; // ISO
    action: string;
    details?: string;
}
