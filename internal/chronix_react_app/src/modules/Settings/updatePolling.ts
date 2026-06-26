import type {UpdaterStatus} from './types.ts'

type WaitForServerVersionOptions = {
    maxAttempts?: number
    intervalMs?: number
    initialDelayMs?: number
}

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => {
        globalThis.setTimeout(resolve, ms)
    })
}

export async function waitForServerVersion(
    fetchStatus: () => Promise<UpdaterStatus>,
    targetVersion: string,
    options: WaitForServerVersionOptions = {},
): Promise<UpdaterStatus> {
    const maxAttempts = options.maxAttempts ?? 90
    const intervalMs = options.intervalMs ?? 2000
    const initialDelayMs = options.initialDelayMs ?? 1500
    let lastSeenVersion = ''

    if (initialDelayMs > 0) {
        await sleep(initialDelayMs)
    }

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
        try {
            const status = await fetchStatus()
            lastSeenVersion = status.currentVersion
            if (status.currentVersion === targetVersion) {
                return status
            }
        } catch {
            // Ignore transient restart failures while the process is cycling.
        }

        if (attempt < maxAttempts - 1) {
            await sleep(intervalMs)
        }
    }

    if (lastSeenVersion) {
        throw new Error(`Timed out waiting for Chronix ${targetVersion}; last reported version was ${lastSeenVersion}.`)
    }
    throw new Error(`Timed out waiting for Chronix ${targetVersion} to come back online.`)
}
