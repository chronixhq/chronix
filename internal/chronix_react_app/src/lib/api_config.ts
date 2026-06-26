type ViteEnv = {
    MODE?: string
    VITE_API_BASE_URL?: string
}

function resolveBaseURL(): string {
    const env = (import.meta as unknown as { env?: ViteEnv }).env
    const configured = env?.VITE_API_BASE_URL?.trim()
    if (configured) {
        return configured.replace(/\/+$/, '');
    }

    if (env?.MODE === 'development' && typeof window !== 'undefined') {
        return `${window.location.protocol}//${window.location.hostname}:5170`
    }

    return window.location.origin
}

export const baseURL = resolveBaseURL()
