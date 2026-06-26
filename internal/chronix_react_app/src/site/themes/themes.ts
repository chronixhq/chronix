import {alpha, createTheme, type Theme} from '@mui/material/styles';

const darkBackground = {
    default: '#0e1727',
    paper: '#172339',
};

const commonTheme: any = {
    palette: {
        secondary: {
            main: '#f06a23',
        },
    },
    shape: {
        borderRadius: 5,
    },
    typography: {
        fontFamily: 'Inter, Arial, sans-serif',
    },
    components: {
        MuiCard: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    backgroundImage: 'none',
                    border: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.16 : 0.12)}`,
                    boxShadow: theme.palette.mode === 'dark'
                        ? `0 14px 30px ${alpha('#030712', 0.18)}`
                        : `0 10px 22px ${alpha('#0f172a', 0.07)}`,
                    backdropFilter: 'blur(18px)',
                }),
            },
        },
        MuiCardHeader: {
            styleOverrides: {
                root: {
                    padding: '12px 16px',
                },
            },
        },
        MuiPaper: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    backgroundImage: 'none',
                    position: 'relative',
                    zIndex: 1,
                    border: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.12 : 0.1)}`,
                    boxShadow: theme.palette.mode === 'dark'
                        ? `0 12px 26px ${alpha('#020617', 0.16)}`
                        : `0 10px 20px ${alpha('#0f172a', 0.07)}`,
                }),
            },
        },
        MuiDataGrid: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    backgroundColor: theme.palette.background.paper,
                    color: theme.palette.text.primary,
                    border: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.14 : 0.12)}`,
                    borderRadius: 6,
                    position: 'relative',
                    zIndex: 1,
                    isolation: 'isolate',
                    '& .MuiDataGrid-columnHeaders': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.06),
                        borderBottom: `1px solid ${theme.palette.divider}`,
                    },
                    '& .MuiDataGrid-footerContainer': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.06),
                        borderTop: `1px solid ${theme.palette.divider}`,
                    },
                    '& .MuiDataGrid-row': {
                        backgroundColor: theme.palette.background.paper,
                        '&:hover': {
                            backgroundColor: theme.palette.action.hover,
                        },
                    },
                    '& .MuiDataGrid-cell': {
                        color: 'inherit',
                        borderBottom: `1px solid ${theme.palette.divider}`,
                    },
                    '& .MuiDataGrid-virtualScroller': {
                        zIndex: 1,
                    },
                    '& .MuiDataGrid-overlay': {
                        backgroundColor: theme.palette.background.paper,
                        color: 'inherit',
                        zIndex: 10,
                    },
                }),
            },
        },
        MuiButton: {
            defaultProps: {
                variant: 'contained',
            },
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    borderRadius: 6,
                    fontWeight: 600,
                    letterSpacing: '0.02em',
                    textTransform: 'none',
                    boxShadow: theme.palette.mode === 'dark'
                        ? `0 6px 16px ${alpha(theme.palette.primary.main, 0.14)}`
                        : `0 5px 14px ${alpha(theme.palette.primary.main, 0.1)}`,
                }),
                containedPrimary: ({theme}: { theme: Theme }) => ({
                    backgroundImage: `linear-gradient(135deg, ${alpha(theme.palette.primary.light, 0.98)} 0%, ${alpha(theme.palette.primary.main, 0.94)} 55%, ${alpha('#6e95ef', 0.94)} 100%)`,
                    color: theme.palette.mode === 'dark' ? '#08101d' : theme.palette.common.white,
                }),
                outlined: ({theme}: { theme: Theme }) => ({
                    borderColor: alpha(theme.palette.primary.light, 0.35),
                    backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.05 : 0.02),
                }),
            },
        },
        MuiIconButton: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    borderRadius: 8,
                    '&:hover': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.12 : 0.08),
                    },
                }),
            },
        },
        MuiChip: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    borderRadius: 999,
                    border: `1px solid ${alpha(theme.palette.common.white, theme.palette.mode === 'dark' ? 0.12 : 0.08)}`,
                }),
            },
        },
        MuiListItemButton: {
            styleOverrides: {
                root: ({theme}: { theme: Theme }) => ({
                    borderRadius: 8,
                    transition: 'background-color 140ms ease, border-color 140ms ease, transform 140ms ease',
                    '&:hover': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.1 : 0.06),
                    },
                    '&.Mui-selected': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.18 : 0.12),
                        border: `1px solid ${alpha(theme.palette.primary.light, theme.palette.mode === 'dark' ? 0.28 : 0.22)}`,
                        boxShadow: `inset 0 1px 0 ${alpha(theme.palette.common.white, theme.palette.mode === 'dark' ? 0.07 : 0.2)}`,
                    },
                    '&.Mui-selected:hover': {
                        backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.22 : 0.14),
                    },
                }),
            },
        },
        MuiAppBar: {
            defaultProps: {
                elevation: 0,
            },
        },
        MuiTextField: {
            defaultProps: {
                size: 'medium',
            },
        },
        MuiFormControl: {
            defaultProps: {
                size: 'medium',
            },
        },
        MuiCssBaseline: {
            styleOverrides: (theme: Theme) => ({
                html: {
                    minHeight: '100%',
                },
                body: {
                    minHeight: '100%',
                    backgroundColor: theme.palette.background.default,
                    backgroundImage: theme.palette.mode === 'dark'
                        ? `radial-gradient(circle at top left, ${alpha(theme.palette.primary.main, 0.18)} 0%, transparent 34%), radial-gradient(circle at top right, ${alpha('#7dd3fc', 0.12)} 0%, transparent 24%), linear-gradient(180deg, ${alpha('#15243d', 0.72)} 0%, ${theme.palette.background.default} 18%, ${theme.palette.background.default} 100%)`
                        : `linear-gradient(180deg, ${alpha(theme.palette.primary.light, 0.08)} 0%, ${theme.palette.background.default} 16%, ${theme.palette.background.default} 100%)`,
                    backgroundAttachment: 'fixed',
                },
                '#root': {
                    minHeight: '100vh',
                },
                '::selection': {
                    backgroundColor: alpha(theme.palette.primary.main, 0.32),
                },
                '@media (max-width:899.95px)': {
                    '.MuiInputBase-root.MuiInputBase-sizeSmall': {
                        minHeight: '44px',
                    },
                    '.MuiOutlinedInput-root.MuiInputBase-sizeSmall': {
                        minHeight: '44px',
                    },
                },
            }),
        },
    },
};

export const darkTheme: Theme = createTheme({
    palette: {
        mode: 'dark',
        primary: {
            main: '#91b8ff',
            light: '#c2d8ff',
            dark: '#5f87d6',
            contrastText: '#07101e',
        },
        background: {
            default: darkBackground.default,
            paper: darkBackground.paper,
        },
        text: {
            primary: 'rgba(241, 246, 255, 0.94)',
            secondary: 'rgba(205, 218, 240, 0.78)',
        },
        divider: alpha('#b9d2ff', 0.14),
        success: {
            main: '#59d38c',
        },
        error: {
            main: '#ff6d63',
        },
        warning: {
            main: '#f4ae42',
        },
    },
    components: {
        MuiDrawer: {
            styleOverrides: {
                paper: {
                    backgroundImage: 'none',
                },
            },
        },
    },
}, commonTheme);

export const lightTheme: Theme = createTheme({
    palette: {
        mode: 'light',
        primary: {
            main: '#2f65cb',
            light: '#6e95ef',
            dark: '#204a98',
        },
        background: {
            default: '#f4f8ff',
            paper: '#ffffff',
        },
        text: {
            primary: 'rgba(0,0,0,0.87)',
            secondary: 'rgb(88, 102, 125)',
        },
        divider: alpha('#0f172a', 0.1),
    },
}, commonTheme);
