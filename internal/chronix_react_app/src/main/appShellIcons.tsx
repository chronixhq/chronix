import DashboardIcon from '@mui/icons-material/Dashboard'
import DatasetIcon from '@mui/icons-material/Dataset'
import DirectionsRunIcon from '@mui/icons-material/DirectionsRun'
import GroupIcon from '@mui/icons-material/Group'
import ListAltIcon from '@mui/icons-material/ListAlt'
import PersonIcon from '@mui/icons-material/Person'
import SecurityIcon from '@mui/icons-material/Security'
import WorkIcon from '@mui/icons-material/Work'
import type {SvgIconComponent} from '@mui/icons-material'
import type {AppShellIconKey} from './appShellManifest.ts'

const ICONS: Record<AppShellIconKey, SvgIconComponent> = {
  dashboard: DashboardIcon,
  connections: DatasetIcon,
  actions: DirectionsRunIcon,
  jobs: WorkIcon,
  runs: DirectionsRunIcon,
  activity: ListAltIcon,
  agents: GroupIcon,
  settings: SecurityIcon,
  profile: PersonIcon,
}

export function getGlobalNavIcon(icon: AppShellIconKey): SvgIconComponent {
  return ICONS[icon]
}
