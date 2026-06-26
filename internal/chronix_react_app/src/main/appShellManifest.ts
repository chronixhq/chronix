export const GLOBAL_RAIL_WIDTH = 92
export const MODULE_SIDENAV_WIDTH = 240

export const APP_SHELL_PATHS = {
  dashboard: '/',
  connectionsAll: '/connections/all',
  actionsList: '/actions/list',
  jobsList: '/jobs/list',
  runs: '/runs',
  activity: '/activity',
  agents: '/agents',
  settingsOverview: '/settings/overview',
  userProfile: '/user/profile',
  userReset: '/user/reset',
} as const

export const HELP_SECTIONS = {
  notifications: '3. Notification Setup',
  connections: '4. Creating Connections',
  actions: '5. Creating Actions',
  jobs: '6. Creating Jobs',
  agents: '7. Chronix Agents',
  updates: '8. Maintenance and Updates',
  activity: '9. Activity, Reporting, and Branding',
  feedback: '10. Feedback and Bug Reporting',
  openSource: '11. Open Source',
} as const

export type AppModule = 'connections' | 'actions' | 'jobs' | 'settings' | 'user'

const MODULE_PREFIXES: Record<AppModule, readonly string[]> = {
  connections: ['/connections', '/databases', '/shell', '/webtasks'],
  actions: ['/actions'],
  jobs: ['/jobs'],
  settings: ['/settings'],
  user: ['/user'],
}

export function getAppModule(pathname: string): AppModule | null {
  const moduleEntry = (Object.entries(MODULE_PREFIXES) as Array<[AppModule, readonly string[]]>)
    .find(([, prefixes]) => prefixes.some((prefix) => pathname.startsWith(prefix)))
  return moduleEntry?.[0] ?? null
}

export function hasModuleSideNav(pathname: string): boolean {
  return getAppModule(pathname) !== null
}

export type AppShellIconKey =
  | 'dashboard'
  | 'connections'
  | 'actions'
  | 'jobs'
  | 'runs'
  | 'activity'
  | 'agents'
  | 'settings'
  | 'profile'

export interface GlobalNavItem {
  id: string
  label: string
  icon: AppShellIconKey
  path: string
  adminOnly?: boolean
  matches: (pathname: string) => boolean
}

export const globalNavItems: readonly GlobalNavItem[] = [
  {
    id: 'dashboard',
    label: 'Dashboard',
    icon: 'dashboard',
    path: APP_SHELL_PATHS.dashboard,
    matches: (pathname) => pathname === APP_SHELL_PATHS.dashboard,
  },
  {
    id: 'connections',
    label: 'Connections',
    icon: 'connections',
    path: APP_SHELL_PATHS.connectionsAll,
    matches: (pathname) => getAppModule(pathname) === 'connections',
  },
  {
    id: 'actions',
    label: 'Actions',
    icon: 'actions',
    path: APP_SHELL_PATHS.actionsList,
    matches: (pathname) => getAppModule(pathname) === 'actions',
  },
  {
    id: 'jobs',
    label: 'Jobs',
    icon: 'jobs',
    path: APP_SHELL_PATHS.jobsList,
    matches: (pathname) => getAppModule(pathname) === 'jobs',
  },
  {
    id: 'runs',
    label: 'Runs',
    icon: 'runs',
    path: APP_SHELL_PATHS.runs,
    matches: (pathname) => pathname.startsWith(APP_SHELL_PATHS.runs),
  },
  {
    id: 'activity',
    label: 'Activity',
    icon: 'activity',
    path: APP_SHELL_PATHS.activity,
    matches: (pathname) => pathname.startsWith(APP_SHELL_PATHS.activity),
  },
  {
    id: 'agents',
    label: 'Agents',
    icon: 'agents',
    path: APP_SHELL_PATHS.agents,
    matches: (pathname) => pathname.startsWith(APP_SHELL_PATHS.agents),
  },
  {
    id: 'settings',
    label: 'Settings',
    icon: 'settings',
    path: APP_SHELL_PATHS.settingsOverview,
    adminOnly: true,
    matches: (pathname) => getAppModule(pathname) === 'settings',
  },
  {
    id: 'profile',
    label: 'My Profile',
    icon: 'profile',
    path: APP_SHELL_PATHS.userProfile,
    matches: (pathname) => getAppModule(pathname) === 'user',
  },
] as const
