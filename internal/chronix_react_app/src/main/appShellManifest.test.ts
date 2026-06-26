import {describe, expect, it} from 'vitest'
import {getAppModule, globalNavItems, hasModuleSideNav} from './appShellManifest.ts'

describe('appShellManifest', () => {
    it('maps module paths to the expected shell module', () => {
        expect(getAppModule('/connections/all')).toBe('connections')
        expect(getAppModule('/databases/list')).toBe('connections')
        expect(getAppModule('/shell/edit/123')).toBe('connections')
        expect(getAppModule('/actions/list')).toBe('actions')
        expect(getAppModule('/jobs/create')).toBe('jobs')
        expect(getAppModule('/settings/overview')).toBe('settings')
        expect(getAppModule('/user/profile')).toBe('user')
        expect(getAppModule('/runs')).toBeNull()
    })

    it('keeps the global nav matchers aligned with grouped connection routes', () => {
        const connectionsItem = globalNavItems.find((item) => item.id === 'connections')
        const profileItem = globalNavItems.find((item) => item.id === 'profile')

        expect(connectionsItem?.matches('/connections/all')).toBe(true)
        expect(connectionsItem?.matches('/databases/list')).toBe(true)
        expect(connectionsItem?.matches('/shell/list')).toBe(true)
        expect(connectionsItem?.matches('/webtasks/list')).toBe(true)
        expect(connectionsItem?.matches('/actions/list')).toBe(false)

        expect(profileItem?.matches('/user/reset')).toBe(true)
        expect(profileItem?.matches('/user/profile')).toBe(true)
    })

    it('only enables module side navigation for modules that define one', () => {
        expect(hasModuleSideNav('/connections/all')).toBe(true)
        expect(hasModuleSideNav('/settings/overview')).toBe(true)
        expect(hasModuleSideNav('/activity')).toBe(false)
        expect(hasModuleSideNav('/runs')).toBe(false)
    })
})
