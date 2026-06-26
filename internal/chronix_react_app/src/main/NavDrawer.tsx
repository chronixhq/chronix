import {Box, Divider, Drawer, List, ListItem, ListItemButton, ListItemIcon, type Theme, Tooltip, Typography} from "@mui/material";
import {alpha, useTheme} from "@mui/material/styles";
import useMediaQuery from '@mui/material/useMediaQuery';
import {useLocation, useNavigate} from "react-router";
import {useAuthContext} from "../data/useAuthContext.ts";
import pkg from '../../package.json';
import {GLOBAL_RAIL_WIDTH, globalNavItems} from "./appShellManifest.ts";
import {getGlobalNavIcon} from './appShellIcons.tsx';
import ChronixGearMark from "../assets/Chronix-gears.png";

export const NavDrawer = ({appbarHeight}: { appbarHeight: number }) => {
    const theme: Theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const navigate = useNavigate();
    const location = useLocation();
    const {user} = useAuthContext();

    if (isMobile) return null;

    const drawer = (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
                position: 'relative',
                overflow: 'hidden',
                '&::before': {
                    content: '""',
                    position: 'absolute',
                    inset: 0,
                    pointerEvents: 'none',
                    background: `linear-gradient(180deg, ${alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.04)} 0%, transparent 22%, transparent 78%, ${alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.05 : 0.03)} 100%)`,
                },
            }}
        >
            <Box
                component="img"
                src={ChronixGearMark}
                alt=""
                sx={{
                    position: 'absolute',
                    left: '50%',
                    bottom: 120,
                    width: 142,
                    transform: 'translateX(-50%)',
                    opacity: theme.palette.mode === 'dark' ? 0.09 : 0.06,
                    filter: 'blur(0.4px)',
                    pointerEvents: 'none',
                    userSelect: 'none',
                }}
            />
            <Box sx={{overflowY: 'auto', flexGrow: 1, minHeight: 0, borderRight: `1px solid ${theme.palette.divider}`, position: 'relative', zIndex: 1}}>
                <List sx={{py: 1.1, px: 0.75}}>
                    {globalNavItems.map((item, index) => {
                        if (item.adminOnly && !user?.admin) return null;
                        const selected = item.matches(location.pathname);
                        const Icon = getGlobalNavIcon(item.icon);
                        return (
                            <ListItem key={index} disablePadding sx={{display: 'block'}}>
                                <Tooltip title={item.label} placement="right">
                                    <ListItemButton
                                        selected={selected}
                                        onClick={() => navigate(item.path)}
                                        sx={{
                                            minHeight: 66,
                                            justifyContent: 'center',
                                            px: 0.5,
                                            flexDirection: 'column',
                                            gap: 0.35,
                                            textAlign: 'center',
                                            mb: 0.5,
                                            border: `1px solid ${selected ? alpha(theme.palette.primary.light, 0.22) : 'transparent'}`,
                                            background: selected
                                                ? `linear-gradient(180deg, ${alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.16 : 0.12)} 0%, ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.1 : 0.08)} 100%)`
                                                : 'transparent',
                                        }}
                                    >
                                        <ListItemIcon sx={{color: 'inherit', minWidth: 0}}><Icon/></ListItemIcon>
                                        <Typography
                                            variant="caption"
                                            sx={{
                                                color: "text.primary",
                                                fontSize: '0.68rem',
                                                lineHeight: 1.12,
                                                whiteSpace: 'normal',
                                                textWrap: 'balance',
                                                maxWidth: '100%',
                                            }}
                                        >
                                            {item.label}
                                        </Typography>
                                    </ListItemButton>
                                </Tooltip>
                            </ListItem>
                        );
                    })}
                </List>
                <Divider/>
            </Box>

            {/* Drawer Footer (Chronix text always at bottom) */}
            <Box sx={{
                py: 2,
                px: 2,
                background: `linear-gradient(180deg, ${alpha(theme.palette.background.paper, 0.3)} 0%, ${alpha(theme.palette.background.paper, 0.92)} 100%)`,
                textAlign: 'center',
                flexShrink: 0,
                borderTop: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.14 : 0.1)}`,
                position: 'relative',
                zIndex: 1,
            }}>
                <Typography variant="body2" sx={{
                    color: "text.secondary"
                }}>
                    &copy; {new Date().getFullYear()} Chronix
                </Typography>
                <Typography variant="caption" sx={{color: "text.secondary", display: 'block', mt: 0.25}}>
                    v{pkg.version}
                </Typography>
            </Box>
        </Box>
    );

    return (<>
        <Drawer
            variant="permanent"
            open
            sx={{
                '& .MuiDrawer-paper': {
                    boxSizing: 'border-box',
                    width: GLOBAL_RAIL_WIDTH,
                    background: theme.palette.mode === 'dark'
                        ? `linear-gradient(180deg, ${alpha('#14253d', 0.98)} 0%, ${alpha('#162238', 0.96)} 100%)`
                        : theme.palette.background.paper,
                    display: 'flex',
                    flexDirection: 'column',
                    height: '100%',
                    paddingTop: `${appbarHeight}px`,
                    overflow: 'hidden',
                    borderRight: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.14 : 0.08)}`,
                    boxShadow: theme.palette.mode === 'dark'
                        ? `inset -1px 0 0 ${alpha(theme.palette.common.white, 0.03)}`
                        : 'none',
                },
            }}
        >
            {drawer}
        </Drawer>
    </>);
}
