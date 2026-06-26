import {type ReactNode, useEffect, useMemo, useState} from 'react';
import {CssBaseline, type Theme, ThemeProvider as MuiThemeProvider} from '@mui/material';
import {darkTheme, lightTheme} from './themes.ts';
import {ColorModeContext} from "./colorModeContext.ts";

interface ThemeContextProviderProps {
    children: ReactNode;
}

export function ThemeContextProvider({children}: ThemeContextProviderProps) {
    // Get initial mode from localStorage or default to 'light'
    const [mode, setMode] = useState<'light' | 'dark'>(() => {
        try {
            const storedMode = localStorage.getItem('themeMode');
            return storedMode === 'light' ? 'light' : 'dark';
        } catch (error) {
            console.error("Failed to read theme mode from localStorage:", error);
            return 'light'; // Fallback
        }
    });

    // Save mode to localStorage whenever it changes
    useEffect(() => {
        try {
            localStorage.setItem('themeMode', mode);
        } catch (error) {
            console.error("Failed to save theme mode to localStorage:", error);
        }
    }, [mode]);

    // Memoize the context value to prevent unnecessary re-renders
    const colorMode = useMemo(
        () => ({
            toggleColorMode: () => {
                setMode((prevMode) => (prevMode === 'light' ? 'dark' : 'light'));
            },
            mode, // Expose the current mode if other components need to know it
        }),
        [mode],
    );

    // Create the actual MUI theme based on the current mode
    const theme: Theme = useMemo(
        () => (mode === 'light' ? lightTheme : darkTheme),
        [mode],
    );

    return (
        <ColorModeContext.Provider value={colorMode}>
            <MuiThemeProvider theme={theme}>
                <CssBaseline/> {/* Global CSS Baseline for consistent styling */}
                {children}
            </MuiThemeProvider>
        </ColorModeContext.Provider>
    );
}

