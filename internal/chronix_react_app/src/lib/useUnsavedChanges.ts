import {useEffect} from 'react'
import type {RefObject} from 'react'
import type {NavigateFunction} from 'react-router'

// Minimal unsaved-changes guard. Shows a native prompt on page/tab close or reload.
// For in-app navigations, call confirmOnNavigate() before programmatic route changes.
export function useUnsavedChanges(enabled: boolean, allowUnloadRef?: RefObject<boolean>) {
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (!enabled) return
      if (allowUnloadRef?.current) return
      e.preventDefault()
      e.returnValue = ''
    }
    window.addEventListener('beforeunload', handler)
    return () => window.removeEventListener('beforeunload', handler)
  }, [allowUnloadRef, enabled])
}

// Helper to confirm navigation from within components (e.g., Cancel button)
// Requires a confirm function (e.g., useMuiPrompts().confirmPrompt) to avoid native dialogs.
export function confirmOnNavigate(
  enabled: boolean,
  navigate: NavigateFunction,
  confirmFn: (props: { title: string; message: string; buttonText?: string; cancelButtonText?: string }) => Promise<boolean>
) {
  return async (to: string) => {
    if (!enabled) { navigate(to); return }
    let ok: boolean
      try {
      ok = await confirmFn({
        title: 'Unsaved changes',
        message: 'You have unsaved changes. Leave this page?',
        buttonText: 'Leave page',
        cancelButtonText: 'Stay'
      })
    } catch {
      ok = false
    }
    if (ok) navigate(to)
  }
}
