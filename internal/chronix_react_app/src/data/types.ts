export const ServerStatus = {
    UNKNOWN: "unknown",
    UNINITIALIZED: "uninitialized",
    SUSPENDED: "suspended",
    STARTINGUP: "startingup",
    ACTIVE: "active",
}

export const SiteState = {
    LOADING: "loading",
    READY: "ready",
    ERROR: "error",
}

export type NotificationSeverity = 'info' | 'success' | 'warning' | 'error'
export const NotificationSeverity = {
    INFO: 'info',
    SUCCESS: 'success',
    WARNING: 'warning',
    ERROR: 'error',
} as const

export type NotificationCategory = 'job' | 'system' | 'security' | 'connection' | 'worker' | 'backup' | 'upgrade'
export const NotificationCategory = {
    JOB: 'job',
    SYSTEM: 'system',
    SECURITY: 'security',
    CONNECTION: 'connection',
    WORKER: 'worker',
    BACKUP: 'backup',
    UPGRADE: 'upgrade',
} as const

export interface NotificationItem {
    id: number
    createdAt: string
    category: NotificationCategory
    severity: NotificationSeverity
    subject: string
    origin?: string
    data?: { message?: string } | Record<string, any>
    seen: boolean
}

export type FeatureUsageKind = 'agents' | 'jobs' | 'db_connections' | 'shell_connections' | 'webtask_connections' | 'actions' | 'users'

export interface FeatureAvailabilityStatus {
    usage: Record<FeatureUsageKind, number>
    features: {
        sms: boolean
        webhooks: boolean
        csvReports: boolean
        htmlReports: boolean
        pdfReports: boolean
        branding: boolean
    }
    feedbackEnabled?: boolean
    branding: {
        logoUrl?: string
        name?: string
    }
}

export interface FeedbackAttachment {
    id: number;
    fileName: string;
    filePath: string;
    contentType: string;
    fileSize: number;
    createdAt: string;
}

export interface FeedbackItem {
    id: number;
    summary: string;
    description: string;
    userId: number;
    createdAt: string;
    status: 'open' | 'closed' | 'in-progress';
    feedbackAttachments?: FeedbackAttachment[];
}
