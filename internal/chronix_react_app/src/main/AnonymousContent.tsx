import {Navigate, Route, Routes} from "react-router";
import {VStack} from "@dsherwin/mui-kit";
import {useTheme} from "@mui/material/styles";
import {UserLogin} from "../modules/User/UserLogin.tsx";
import {AdminLogin} from "../modules/Settings/AdminLogin.tsx";
import {InitialSetup} from "../modules/Settings/InitialSetup.tsx";
import {APP_SHELL_PATHS} from "./appShellManifest.ts";

export const AnonymousContent = () => {
    const theme = useTheme();
    return (
        <>
            <VStack
                component="main"
                sx={{
                    flexGrow: 1,
                    marginTop: 0,
                    backgroundColor: theme.palette.background.default,
                    display: 'flex',
                    minHeight: `100vh`,
                }}
            >
                <Routes>
                    <Route index element={<UserLogin/>}/>
                    <Route path={"/settings"} element={<AdminLogin/>}/>
                    <Route path={"/settings/setup"} element={<InitialSetup/>}/>
                    <Route path="*" element={<Navigate to={APP_SHELL_PATHS.dashboard} replace />} />
                </Routes>
            </VStack>
        </>
    );
};
