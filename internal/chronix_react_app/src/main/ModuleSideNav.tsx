import React from 'react';
import {Box, Divider, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography} from '@mui/material';
import {useLocation, useNavigate, useSearchParams} from 'react-router';
import {AddCircleOutlined, BrandingWatermark, BugReport as BugReportIcon, Dataset, DirectionsRun, Group, Http, Lightbulb as LightbulbIcon, Link as LinkIcon, ListAlt, NotificationsActive, Security, Sms, Storage as StorageIcon, SystemUpdate, Terminal, Webhook, Work} from '@mui/icons-material';
import {useTheme} from '@mui/material/styles';
import {alpha} from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import {useAuthContext} from '../data/useAuthContext';
import {useConnections} from '../data/ConnectionsContext';
import {useActions} from '../data/ActionsContext';
import {useJobs} from '../data/JobsContext';
import {useFeatureAvailability} from "../data/FeatureAvailabilityContext.tsx";
import {GLOBAL_RAIL_WIDTH, MODULE_SIDENAV_WIDTH, getAppModule, hasModuleSideNav} from "./appShellManifest.ts";
import type {Action} from '../modules/Actions/types.ts';
import type {Job} from '../modules/Jobs/types.ts';

export const ModuleSideNavContent = ({onNavigate}: { onNavigate?: () => void }) => {
    const theme = useTheme();
    const navigate = useNavigate();
    const location = useLocation();
    const [searchParams] = useSearchParams();
    const {user} = useAuthContext();

    const path = location.pathname;

    // Determine which module we're in
    const inModule = getAppModule(path);

    const {items: connections, ensureLoaded: ensureConnectionsLoaded} = useConnections();
    const {items: actions, ensureLoaded: ensureActionsLoaded} = useActions();
    const {items: jobs, ensureLoaded: ensureJobsLoaded} = useJobs();
    const {data: featureData, checkLimit} = useFeatureAvailability();

    const dbLimit = checkLimit('db_connections');
    const shLimit = checkLimit('shell_connections');
    const wtLimit = checkLimit('webtask_connections');
    const actionLimit = checkLimit('actions');
    const jobLimit = checkLimit('jobs');

    const connItems = React.useMemo(() => (connections || []).map(c => ({
        id: String(c.id ?? ''),
        name: String(c.name ?? ''),
        kind: c.kind,
        suspended: c.suspended,
        enabled: c.enabled !== false
    })).filter(it => it.id && it.name).sort((a, b) => a.name.localeCompare(b.name)), [connections]);

    const actionItems = React.useMemo(() => (actions || []).map((action: Action) => ({
        id: String(action.id ?? ''),
        name: String(action.name ?? ''),
        actionType: action.actionType,
        suspended: action.suspended
    })).filter(it => it.id && it.name).sort((a, b) => a.name.localeCompare(b.name)), [actions]);

    const jobItems = React.useMemo(() => (jobs || []).map((job: Job) => ({
        id: String(job.id ?? ''),
        name: String(job.name ?? ''),
        suspended: job.suspended
    })).filter(it => it.id && it.name).sort((a, b) => a.name.localeCompare(b.name)), [jobs]);

    React.useEffect(() => {
        if (inModule === 'connections') {
            void ensureConnectionsLoaded();
        } else if (inModule === 'actions') {
            void ensureActionsLoaded();
        } else if (inModule === 'jobs') {
            void ensureJobsLoaded();
        }
    }, [ensureActionsLoaded, ensureConnectionsLoaded, ensureJobsLoaded, inModule]);

    if (!inModule) return null;

    const SectionHeader = ({text}: { text: string }) => (
        <Typography
            variant="overline"
            sx={{
                px: 2,
                pt: 1.25,
                pb: 0.5,
                color: theme.palette.text.secondary,
                letterSpacing: '0.14em',
            }}
        >
            {text}
        </Typography>
    );

    const goTo = (to: string) => {
        navigate(to)
        onNavigate?.()
    }

    const Item = ({label, icon, to, selected, disabled, dimmed}: { label: string; icon: React.ReactNode; to: string; selected: boolean; disabled?: boolean; dimmed?: boolean }) => (
        <ListItem disablePadding sx={{opacity: dimmed ? 0.6 : 1}}>
            <ListItemButton
                selected={selected}
                onClick={() => goTo(to)}
                disabled={disabled}
                sx={{
                    mx: 1,
                    my: 0.18,
                    border: `1px solid ${selected ? alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.14 : 0.12) : 'transparent'}`,
                    background: selected
                        ? `linear-gradient(180deg, ${alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.12 : 0.09)} 0%, ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.07 : 0.04)} 100%)`
                        : 'transparent',
                    '& .MuiListItemIcon-root': {
                        color: selected ? theme.palette.common.white : 'inherit',
                    },
                }}
            >
                <ListItemIcon sx={{minWidth: 34, opacity: disabled ? 0.5 : (selected ? 1 : 0.8)}}>{icon}</ListItemIcon>
                <ListItemText primary={<Typography variant="body2" sx={{color: disabled ? 'text.disabled' : 'inherit', fontSize: '0.88rem'}}>{label}</Typography>}/>
            </ListItemButton>
        </ListItem>
    );

    return (
        <List dense disablePadding>
            {inModule === 'connections' && (<>
                <SectionHeader text="Connections"/>
                <Item label="All Connections" icon={<Dataset/>} to="/connections/all" selected={path.startsWith('/connections/all')}/>
                <Item label="Database Connections" icon={<ListAlt sx={{color: '#1976d2'}}/>} to="/databases/list" selected={path.startsWith('/databases/list')}/>
                <Item label="Shell Connections" icon={<Terminal sx={{color: '#9c27b0'}}/>} to="/shell/list" selected={path.startsWith('/shell/list')}/>
                <Item label="Web Task Connections" icon={<Http sx={{color: '#ed6c02'}}/>} to="/webtasks/list" selected={path.startsWith('/webtasks/list')}/>
                <Divider sx={{my: 1}}/>
                <Item label="New Database Connection" icon={<AddCircleOutlined sx={{color: '#1976d2'}}/>} to="/databases/add" selected={path.startsWith('/databases/add')} disabled={!dbLimit.allowed}/>
                <Item label="New Shell Connection" icon={<AddCircleOutlined sx={{color: '#9c27b0'}}/>} to="/shell/add" selected={path.startsWith('/shell/add')} disabled={!shLimit.allowed}/>
                <Item label="New Web Task Connection" icon={<AddCircleOutlined sx={{color: '#ed6c02'}}/>} to="/webtasks/add" selected={path.startsWith('/webtasks/add')} disabled={!wtLimit.allowed}/>
                <Divider sx={{my: 1}}/>
                {connItems.map(ci => {
                    const kind = ci.kind || 'database';
                    let icon = <ListAlt sx={{color: '#1976d2'}}/>;
                    let prefix = '/databases/edit/';
                    if (kind === 'shell') {
                        icon = <Terminal sx={{color: '#9c27b0'}}/>;
                        prefix = '/shell/edit/';
                    } else if (kind === 'webtask') {
                        icon = <Http sx={{color: '#ed6c02'}}/>;
                        prefix = '/webtasks/edit/';
                    }
                    const to = `${prefix}${encodeURIComponent(ci.id)}`;
                    return (
                        <Item key={`${kind}-${ci.id}`} label={ci.name} icon={icon} to={to} selected={path === to} disabled={ci.suspended} dimmed={!ci.enabled}/>
                    );
                })}
            </>)}

            {inModule === 'actions' && (<>
                <SectionHeader text="Actions"/>
                <Item label="All Actions" icon={<ListAlt/>} to="/actions/list" selected={path === '/actions/list' && !searchParams.get('type')}/>
                <Item label="Database Actions" icon={<StorageIcon sx={{color: '#1976d2'}}/>} to="/actions/list?type=database" selected={path === '/actions/list' && searchParams.get('type') === 'database'}/>
                <Item label="Shell Actions" icon={<Terminal sx={{color: '#9c27b0'}}/>} to="/actions/list?type=shell" selected={path === '/actions/list' && searchParams.get('type') === 'shell'}/>
                <Item label="Web Task Actions" icon={<Http sx={{color: '#ed6c02'}}/>} to="/actions/list?type=webtask" selected={path === '/actions/list' && searchParams.get('type') === 'webtask'}/>
                <Divider sx={{my: 1}}/>
                <Item label="Create DB Action" icon={<AddCircleOutlined sx={{color: '#1976d2'}}/>} to="/actions/create" selected={path.startsWith('/actions/create') && !path.includes('shell') && !path.includes('webtask')} disabled={!actionLimit.allowed}/>
                <Item label="Create Shell Action" icon={<AddCircleOutlined sx={{color: '#9c27b0'}}/>} to="/actions/create-shell" selected={path.startsWith('/actions/create-shell')} disabled={!actionLimit.allowed}/>
                <Item label="Create Web Task Action" icon={<AddCircleOutlined sx={{color: '#ed6c02'}}/>} to="/actions/create-webtask" selected={path.startsWith('/actions/create-webtask')} disabled={!actionLimit.allowed}/>
                <Divider sx={{my: 1}}/>
                {actionItems.map(ai => {
                    const type = ai.actionType || 'database';
                    let icon = <StorageIcon sx={{color: '#1976d2'}}/>;
                    let prefix = '/actions/edit/';
                    if (type === 'shell') {
                        icon = <Terminal sx={{color: '#9c27b0'}}/>;
                        prefix = '/actions/edit-shell/';
                    } else if (type === 'webtask') {
                        icon = <Http sx={{color: '#ed6c02'}}/>;
                        prefix = '/actions/edit-webtask/';
                    }
                    const to = `${prefix}${encodeURIComponent(ai.id)}`;
                    return (
                        <Item key={`${type}-${ai.id}`} label={ai.name} icon={icon} to={to} selected={path === to} disabled={ai.suspended}/>
                    );
                })}
            </>)}

            {inModule === 'jobs' && (<>
                <SectionHeader text="Jobs"/>
                <Item label="All Jobs" icon={<ListAlt/>} to="/jobs/list" selected={path.startsWith('/jobs/list')}/>
                <Item label="Job Create" icon={<AddCircleOutlined/>} to="/jobs/create" selected={path.startsWith('/jobs/create')} disabled={!jobLimit.allowed}/>
                <Item label="Runs" icon={<DirectionsRun/>} to="/runs" selected={path.startsWith('/runs')}/>
                <Divider sx={{my: 1}}/>
                {jobItems.map(ji => (
                    <Item key={ji.id} label={ji.name} icon={<Work/>} to={`/jobs/edit/${encodeURIComponent(ji.id)}`} selected={path === `/jobs/edit/${encodeURIComponent(ji.id)}`} disabled={ji.suspended}/>
                ))}
            </>)}

            {inModule === 'settings' && (<>
                <SectionHeader text="Settings"/>
                {user?.admin && (
                    <>
                        <Item label="Overview" icon={<Security/>} to="/settings/overview" selected={path.startsWith('/settings/overview')}/>
                        <Item label="Server URL" icon={<LinkIcon/>} to="/settings/server-url" selected={path.startsWith('/settings/server-url')}/>
                        <Item label="Email Notifier" icon={<Security/>} to="/settings/email" selected={path.startsWith('/settings/email')}/>
                        <Item label="HTTP / HTTPS / Agent" icon={<Security/>} to="/settings/https" selected={path.startsWith('/settings/https')}/>
                        <Item label="SMS Notifier" icon={<Sms/>} to="/settings/sms" selected={path.startsWith('/settings/sms')}/>
                        <Item label="Alerts" icon={<NotificationsActive/>} to="/settings/alerts" selected={path.startsWith('/settings/alerts')}/>
                        <Item label="Webhooks" icon={<Webhook/>} to="/settings/webhooks" selected={path.startsWith('/settings/webhooks')}/>
                        <Item label="Users" icon={<Group/>} to="/settings/users" selected={path.startsWith('/settings/users')}/>
                        {featureData?.features.branding && (
                            <Item label="Branding" icon={<BrandingWatermark/>} to="/settings/branding" selected={path.startsWith('/settings/branding')}/>
                        )}
                        <Item label="Updates" icon={<SystemUpdate/>} to="/settings/updates" selected={path.startsWith('/settings/updates')}/>
                        {featureData?.feedbackEnabled && (
                            <>
                                <Item label="Bug Reports" icon={<BugReportIcon sx={{color: '#ffffff'}}/>} to="/settings/bug-reports" selected={path.startsWith('/settings/bug-reports')}/>
                                <Item label="Feature Requests" icon={<LightbulbIcon sx={{color: '#ffffff'}}/>} to="/settings/feature-requests" selected={path.startsWith('/settings/feature-requests')}/>
                            </>
                        )}
                    </>
                )}
            </>)}

            {inModule === 'user' && (<>
                <SectionHeader text="User"/>
                <Item label="Reset Password" icon={<Security/>} to="/user/reset" selected={path.startsWith('/user/reset')}/>
                <Item label="Profile" icon={<Group/>} to="/user/profile" selected={path.startsWith('/user/profile')}/>
            </>)}
        </List>
    );
};

export const ModuleSideNav = ({appbarHeight}: { appbarHeight: number }) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const location = useLocation();

    if (isMobile || !hasModuleSideNav(location.pathname)) return null;

    return (
        <Box
            sx={{
                width: MODULE_SIDENAV_WIDTH,
                flex: '0 0 auto',
                position: 'fixed',
                left: GLOBAL_RAIL_WIDTH,
                top: appbarHeight,
                bottom: 0,
                overflowY: 'auto',
                overflowX: 'hidden',
                borderRight: `1px solid ${theme.palette.divider}`,
                background: theme.palette.mode === 'dark'
                    ? `linear-gradient(180deg, ${alpha('#17253c', 0.98)} 0%, ${alpha('#162238', 0.96)} 100%)`
                    : theme.palette.background.paper,
                '&::before': {
                    content: '""',
                    position: 'absolute',
                    inset: 0,
                    pointerEvents: 'none',
                    background: `linear-gradient(180deg, ${alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.03)} 0%, transparent 30%, transparent 100%)`,
                },
            }}
        >
            <ModuleSideNavContent/>
        </Box>
    );
};
