import {Box, Divider, Drawer, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Typography} from '@mui/material'
import {useTheme} from '@mui/material/styles'
import useMediaQuery from '@mui/material/useMediaQuery'
import {useLocation, useNavigate} from 'react-router'
import {useAuthContext} from '../data/useAuthContext.ts'
import {globalNavItems, hasModuleSideNav} from './appShellManifest.ts'
import {getGlobalNavIcon} from './appShellIcons.tsx'
import {ModuleSideNavContent} from './ModuleSideNav.tsx'

export const MobileNavDrawer = ({open, onClose, appbarHeight}: { open: boolean; onClose: () => void; appbarHeight: number }) => {
    const theme = useTheme()
    const isMobile = useMediaQuery(theme.breakpoints.down('md'))
    const navigate = useNavigate()
    const location = useLocation()
    const {user} = useAuthContext()
    const showModuleSideNav = hasModuleSideNav(location.pathname)

    if (!isMobile) return null

    return (
        <Drawer
            anchor="left"
            open={open}
            onClose={onClose}
            variant="temporary"
            ModalProps={{keepMounted: true}}
            sx={{
                '& .MuiDrawer-paper': {
                    boxSizing: 'border-box',
                    width: 320,
                    maxWidth: '85vw',
                    pt: `${appbarHeight}px`,
                },
            }}
        >
            <Box sx={{overflowY: 'auto', height: '100%'}}>
                <List>
                    {globalNavItems.map((item) => {
                        if (item.adminOnly && !user?.admin) return null
                        const selected = item.matches(location.pathname)
                        const Icon = getGlobalNavIcon(item.icon)
                        return (
                            <ListItem key={item.id} disablePadding>
                                <ListItemButton
                                    selected={selected}
                                    onClick={() => {
                                        navigate(item.path)
                                        onClose()
                                    }}
                                >
                                    <ListItemIcon><Icon/></ListItemIcon>
                                    <ListItemText primary={item.label}/>
                                </ListItemButton>
                            </ListItem>
                        )
                    })}
                </List>
                {showModuleSideNav && (
                    <>
                        <Divider/>
                        <Box sx={{px: 2, pt: 1}}>
                            <Typography variant="caption" sx={{
                                color: "text.secondary"
                            }}>Current Module</Typography>
                        </Box>
                        <ModuleSideNavContent onNavigate={onClose}/>
                    </>
                )}
            </Box>
        </Drawer>
    );
}
