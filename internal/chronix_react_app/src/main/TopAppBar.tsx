import {AppBar, Avatar, Badge, Box, Divider, IconButton, ListItemText, Menu, MenuItem, SvgIcon, type Theme, Toolbar, Tooltip, Typography} from "@mui/material";
import type {AppBarProps} from '@mui/material/AppBar';
import ThemeToggle from "../site/themes/ThemeToggle.tsx";
import NotificationsIcon from "@mui/icons-material/Notifications";
import HelpOutlinedIcon from "@mui/icons-material/HelpOutlined";
import {alpha, useTheme} from "@mui/material/styles";
import useMediaQuery from '@mui/material/useMediaQuery';
import {forwardRef, useEffect, useState} from "react";
import ChronixSvg from "../assets/Chronix.svg?react";
import ChronixLettersSvg from "../assets/Chronix-letters.svg?react";
import ChronixGearMark from "../assets/Chronix-gears.png";
import {VStack} from "@dsherwin/mui-kit";
import {useNavigate} from "react-router";
import {useAuthContext} from "../data/useAuthContext.ts";
import {useNotifications} from "../data/NotificationsContext.tsx";
import {formatDateTime} from "../lib/utilities.tsx";
import {useFeatureAvailability} from "../data/FeatureAvailabilityContext.tsx";
import {useHelp} from "../data/HelpContext.tsx";
import BugReportIcon from "@mui/icons-material/BugReport";
import LightbulbIcon from "@mui/icons-material/Lightbulb";
import MenuIcon from '@mui/icons-material/Menu';

export const TopAppBar = forwardRef<HTMLDivElement, AppBarProps & { onToggleNavigation?: () => void }>((props, ref) => {
    const theme: Theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const navigate = useNavigate();
    const {logout, user} = useAuthContext();
    const {items: notifications, unseenCount, markSeen, refresh} = useNotifications();
    const {data: featureData} = useFeatureAvailability();
    const branding = featureData?.branding;
    const brandingEnabled = featureData?.features.branding;
    const {openHelp} = useHelp();

    const [alertsAnchor, setAlertsAnchor] = useState<null | HTMLElement>(null);
    const alertsOpen = Boolean(alertsAnchor);
    const [userAnchor, setUserAnchor] = useState<null | HTMLElement>(null);
    const userOpen = Boolean(userAnchor);

    useEffect(() => {
        if (alertsOpen) {
            const visibleIds = notifications.map(n => n.id);
            if (visibleIds.length) {
                void markSeen(visibleIds);
            }
        } else {
            void refresh();
        }
    }, [alertsOpen]); // eslint-disable-line react-hooks/exhaustive-deps -- [deps-intentional] only react to menu open/close

    const surfaceBorder = `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.18 : 0.14)}`;
    const appBarBackground = theme.palette.mode === 'dark'
        ? `linear-gradient(180deg, ${alpha('#14253f', 0.98)} 0%, ${alpha('#122037', 0.96)} 100%)`
        : `linear-gradient(180deg, ${alpha(theme.palette.primary.main, 0.98)} 0%, ${alpha('#376fd7', 0.96)} 100%)`;
    const actionIconSx = {
        color: theme.palette.common.white,
        borderRadius: 2,
        border: `1px solid ${alpha(theme.palette.common.white, 0.05)}`,
        backgroundColor: alpha(theme.palette.common.white, 0.015),
    };

    const brandLockup = (
        <Box sx={{display: 'flex', alignItems: 'center', gap: 0.95, px: 0.95, py: 0.62}}>
            <Box
                component="img"
                src={ChronixGearMark}
                alt="Chronix gear mark"
                sx={{
                    width: 28,
                    height: 28,
                    borderRadius: 1.75,
                    objectFit: 'cover',
                    border: `1px solid ${alpha(theme.palette.common.white, 0.14)}`,
                    boxShadow: `0 5px 14px ${alpha(theme.palette.primary.main, 0.16)}`,
                }}
            />
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    color: theme.palette.common.white,
                    '& svg': {
                        display: 'block',
                    },
                    '& svg path': {
                        fill: 'currentColor',
                    },
                }}
            >
                <Box sx={{display: {xs: 'none', sm: 'block'}}}>
                    <ChronixLettersSvg style={{display: 'block', width: 114, height: 17}}/>
                </Box>
                <Box sx={{display: {xs: 'block', sm: 'none'}}}>
                    <SvgIcon component={ChronixSvg} inheritViewBox sx={{height: 28, width: 102}}/>
                </Box>
            </Box>
        </Box>
    );

    return (
        <AppBar
            position="fixed"
            ref={ref}
            {...props}
            sx={{
                zIndex: theme.zIndex.drawer + 1,
                width: '100%',
                background: appBarBackground,
                borderBottom: surfaceBorder,
                boxShadow: theme.palette.mode === 'dark'
                    ? `0 14px 28px ${alpha('#020617', 0.22)}`
                    : `0 12px 26px ${alpha('#1d4ed8', 0.12)}`,
                backdropFilter: 'blur(20px)',
            }}
        >
            <Toolbar
                sx={{
                    justifyContent: 'space-between',
                    color: theme.palette.common.white,
                    minHeight: 64,
                    px: {xs: 1.5, md: 2},
                }}
            >
                <Box sx={{display: 'flex', alignItems: 'center'}}>
                    {isMobile && (
                        <IconButton color="inherit" onClick={props.onToggleNavigation} aria-label="open navigation" sx={{mr: 1}}>
                            <MenuIcon/>
                        </IconButton>
                    )}
                    <VStack
                        gap={0}
                        onClick={() => navigate('/')}
                        sx={{
                            cursor: 'pointer',
                            borderRadius: 2.25,
                            border: surfaceBorder,
                            background: `linear-gradient(135deg, ${alpha(theme.palette.common.white, theme.palette.mode === 'dark' ? 0.07 : 0.12)} 0%, ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.08 : 0.12)} 100%)`,
                            boxShadow: `0 8px 18px ${alpha('#020617', theme.palette.mode === 'dark' ? 0.12 : 0.07)}`,
                            overflow: 'hidden',
                        }}
                    >
                        {brandingEnabled && branding?.logoUrl ? (
                            <Box sx={{display: 'flex', alignItems: 'center', gap: 0.95, px: 0.95, py: 0.62}}>
                                <Box
                                    component="img"
                                    src={ChronixGearMark}
                                    alt=""
                                    sx={{
                                        width: 28,
                                        height: 28,
                                        borderRadius: 1.75,
                                        objectFit: 'cover',
                                        border: `1px solid ${alpha(theme.palette.common.white, 0.14)}`,
                                        boxShadow: `0 5px 14px ${alpha(theme.palette.primary.main, 0.14)}`,
                                    }}
                                />
                                <img src={branding.logoUrl} alt={branding?.name || 'Logo'} style={{height: 30, maxWidth: 150, objectFit: 'contain'}}/>
                            </Box>
                        ) : brandingEnabled && branding?.name ? (
                            <Box sx={{display: 'flex', alignItems: 'center', gap: 0.95, px: 0.95, py: 0.62}}>
                                <Box
                                    component="img"
                                    src={ChronixGearMark}
                                    alt=""
                                    sx={{
                                        width: 28,
                                        height: 28,
                                        borderRadius: 1.75,
                                        objectFit: 'cover',
                                        border: `1px solid ${alpha(theme.palette.common.white, 0.14)}`,
                                        boxShadow: `0 5px 14px ${alpha(theme.palette.primary.main, 0.14)}`,
                                    }}
                                />
                                <Typography variant="h6" sx={{fontWeight: "bold", color: 'white', fontSize: '1.05rem'}}>
                                    {branding.name}
                                </Typography>
                            </Box>
                        ) : (
                            brandLockup
                        )}
                    </VStack>
                </Box>

                <Box sx={{display: 'flex', alignItems: 'center', color: 'inherit'}}>
                    <ThemeToggle/>
                    {featureData?.feedbackEnabled && (
                        <>
                            <Tooltip title="Report a Bug">
                                <IconButton color="inherit" onClick={() => navigate('/bug-report')} aria-label="bug report" sx={actionIconSx}>
                                    <BugReportIcon sx={{color: '#ff5252'}}/>
                                </IconButton>
                            </Tooltip>
                            <Tooltip title="Request a Feature">
                                <IconButton color="inherit" onClick={() => navigate('/feature-request')} aria-label="feature request" sx={actionIconSx}>
                                    <LightbulbIcon sx={{color: '#61daff'}}/>
                                </IconButton>
                            </Tooltip>
                        </>
                    )}
                    <IconButton color="warning" onClick={() => openHelp()} aria-label="help" sx={actionIconSx}>
                        <HelpOutlinedIcon/>
                    </IconButton>
                    <IconButton
                        color="inherit"
                        aria-label="notifications"
                        onClick={(e) => setAlertsAnchor(e.currentTarget)}
                        sx={{...actionIconSx, ml: 0.25}}
                    >
                        <Badge color="error" badgeContent={unseenCount} overlap="circular">
                            <NotificationsIcon/>
                        </Badge>
                    </IconButton>
                    <Menu
                        anchorEl={alertsAnchor}
                        open={alertsOpen}
                        onClose={() => setAlertsAnchor(null)}
                        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
                        transformOrigin={{vertical: 'top', horizontal: 'right'}}
                        slotProps={{
                            paper: {
                                sx: {
                                    minWidth: 260,
                                    maxWidth: 360,
                                },
                            },
                        }}
                    >
                        {notifications.length === 0 ? (
                            <MenuItem disabled>
                                <ListItemText primary="No recent alerts"/>
                            </MenuItem>
                        ) : (
                            notifications.map((n) => (
                                <MenuItem key={n.id} onClick={() => setAlertsAnchor(null)} sx={{alignItems: 'start', whiteSpace: 'normal'}}>
                                    <ListItemText
                                        primary={<Typography variant="body2">{n.subject}</Typography>}
                                        secondary={<Typography variant="caption" sx={{color: "text.secondary"}}>{formatDateTime(n.createdAt)}</Typography>}
                                    />
                                </MenuItem>
                            ))
                        )}
                        <Divider/>
                        <MenuItem onClick={() => {
                            setAlertsAnchor(null);
                            navigate('/notifications');
                        }}>
                            <ListItemText primary="View all alerts"/>
                        </MenuItem>
                    </Menu>

                    <IconButton color="inherit" onClick={(e) => setUserAnchor(e.currentTarget)} sx={{ml: 0.75, p: 0.35}} aria-label="user menu">
                        <Avatar
                            sx={{
                                background: `linear-gradient(135deg, ${theme.palette.secondary.main} 0%, ${alpha('#ff9c55', 0.92)} 100%)`,
                                width: 34,
                                height: 34,
                                fontSize: '0.8rem',
                                color: theme.palette.common.white,
                                border: `1px solid ${alpha(theme.palette.common.white, 0.18)}`,
                                boxShadow: `0 7px 18px ${alpha(theme.palette.secondary.main, 0.18)}`,
                            }}
                        >
                            {(() => {
                                const name = user?.name || user?.email || '';
                                const parts = name.trim().split(/\s+/);
                                const initials = parts.length >= 2 ? (parts[0][0] + parts[1][0]) : (parts[0]?.slice(0, 2) || '?');
                                return initials.toUpperCase();
                            })()}
                        </Avatar>
                    </IconButton>
                    <Menu
                        anchorEl={userAnchor}
                        open={userOpen}
                        onClose={() => setUserAnchor(null)}
                        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
                        transformOrigin={{vertical: 'top', horizontal: 'right'}}
                    >
                        <MenuItem onClick={() => {
                            setUserAnchor(null);
                            navigate('/user/profile');
                        }}>
                            <ListItemText primary="Edit Profile"/>
                        </MenuItem>
                        <MenuItem onClick={() => {
                            setUserAnchor(null);
                            navigate('/activity');
                        }}>
                            <ListItemText primary="My Activity"/>
                        </MenuItem>
                        <Divider/>
                        <MenuItem onClick={() => {
                            setUserAnchor(null);
                            logout();
                        }}>
                            <ListItemText primary="Logout"/>
                        </MenuItem>
                    </Menu>
                </Box>
            </Toolbar>
        </AppBar>
    );
});
