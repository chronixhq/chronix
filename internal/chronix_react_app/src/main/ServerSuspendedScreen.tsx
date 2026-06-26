import {Box, Stack, Typography} from "@mui/material";
import DarkModeCorner from "../site/themes/DarkModeCorner";

export const ServerSuspendedScreen = () => {
    return (
        <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', p: 3}}>
            <DarkModeCorner/>
            <Stack spacing={2} sx={{
                alignItems: "center"
            }}>
                <Typography variant="h5" sx={{
                    fontWeight: 600
                }}>Server Temporarily Suspended</Typography>
                <Typography
                    variant="body1"
                    sx={{
                        color: "text.secondary",
                        textAlign: "center",
                        maxWidth: 560
                    }}>
                    Chronix is currently in a suspended state. Administrators can resume the server using the CLI
                    interface. Please try again later.
                </Typography>
            </Stack>
        </Box>
    );
};
