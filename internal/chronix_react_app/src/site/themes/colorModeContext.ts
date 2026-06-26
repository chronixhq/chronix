// Define the shape of our theme context
import {createContext, useContext} from "react";

interface ColorModeContextType {
    toggleColorMode: () => void;
    mode: 'light' | 'dark';
}

export const ColorModeContext = createContext<ColorModeContextType>({
    toggleColorMode: () => {
    },
    mode: 'light',
});
// Custom hook for convenience
export const useThemeMode = () => useContext(ColorModeContext);