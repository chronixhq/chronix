import {Box, CircularProgress, Stack, Typography} from "@mui/material";
import DarkModeCorner from "../site/themes/DarkModeCorner";

export const LoadingScreen = () => {
    return (
        <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', p: 3}}>
            <DarkModeCorner/>
            <Stack spacing={2} sx={{
                alignItems: "center"
            }}>
                <CircularProgress/>
                <Typography variant="h6" sx={{
                    color: "text.secondary"
                }}>Loading Chronix…</Typography>
            </Stack>
        </Box>
    );
};
