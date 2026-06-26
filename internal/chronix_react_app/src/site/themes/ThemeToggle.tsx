import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';
import Brightness4Icon from '@mui/icons-material/Brightness4'; // Moon icon
import Brightness7Icon from '@mui/icons-material/Brightness7';
import {useThemeMode} from "./colorModeContext.ts"; // Sun icon

const ThemeToggle = () => {
//    const theme = useTheme(); // Get the current active MUI theme
    const {toggleColorMode, mode} = useThemeMode(); // Get our custom context values

    return (
        <Box
            sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'text.primary', // Adjust text color as needed
            }}
        >
            <IconButton sx={{ml: 1}} onClick={toggleColorMode} color="inherit">
                {mode === 'dark' ? <Brightness7Icon/> : <Brightness4Icon/>}
            </IconButton>
        </Box>
    );
};

export default ThemeToggle;