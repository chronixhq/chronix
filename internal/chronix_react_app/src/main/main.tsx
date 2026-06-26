/**
 *       _____ _    _ _____   ____  _   _ _______   __
 *      / ____| |  | |  __ \ / __ \| \ | |_   _\ \ / /
 *     | |    | |__| | |__) | |  | |  \| | | |  \ V /
 *     | |    |  __  |  _  /| |  | | . ` | | |   > <
 *     | |____| |  | | | \ \| |__| | |\  |_| |_ / . \
 *      \_____|_|  |_|_|  \_\\____/|_| \_|_____/_/ \_\
 *
 */

import {createRoot} from 'react-dom/client'
import './main.css'
import {App} from "./App.tsx";
import {BrowserRouter} from "react-router";
import {AuthContextProvider} from "../data/AuthContext.tsx";
import {ThemeContextProvider} from "../site/themes/ThemeContext.tsx";
import {SettingsContextProvider} from "../data/SettingsContext.tsx";
import {LocalizationProvider} from '@mui/x-date-pickers/LocalizationProvider';
import {AdapterDayjs} from '@mui/x-date-pickers/AdapterDayjs';
import '../lib/dayjs';

// Global safety net: log any unhandled promise rejections (APIError or otherwise)
window.addEventListener('unhandledrejection', (event) => {
    try {
        const r: any = event.reason;
        if (r && typeof r === 'object') {
            if ('APIErrorData' in r) {
                console.error('APIError:', r);
                if (r.APIErrorData) console.error('APIErrorData:', r.APIErrorData);
            } else {
                console.error('Unhandled rejection:', r);
            }
        } else {
            console.error('Unhandled rejection:', r);
        }
    } catch (e) {
        console.error('Unhandled rejection (logging failed):', e);
    }
});

createRoot(document.getElementById('root')!).render(
    <BrowserRouter>
        <AuthContextProvider>
            <SettingsContextProvider>
                <LocalizationProvider dateAdapter={AdapterDayjs}>
                    <ThemeContextProvider>
                        <App/>
                    </ThemeContextProvider>
                </LocalizationProvider>
            </SettingsContextProvider>
        </AuthContextProvider>
    </BrowserRouter>
)
