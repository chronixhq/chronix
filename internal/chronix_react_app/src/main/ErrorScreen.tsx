import {Box, Button, Stack, Typography} from "@mui/material";
import DarkModeCorner from "../site/themes/DarkModeCorner";

interface Props {
    onRetry?: () => void;
}

export const ErrorScreen = ({onRetry}: Props) => {
    return (
        <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', p: 3}}>
            <DarkModeCorner/>
            <Stack spacing={2} sx={{
                alignItems: "center"
            }}>
                <Typography variant="h5" color="error" sx={{
                    fontWeight: 600
                }}>Something went wrong</Typography>
                <Typography
                    variant="body1"
                    sx={{
                        color: "text.secondary",
                        textAlign: "center",
                        maxWidth: 520
                    }}>
                    We couldn’t connect to the Chronix server. This could be a temporary issue or a network problem.
                </Typography>
                {onRetry && <Button variant="contained" onClick={onRetry}>Try again</Button>}
            </Stack>
        </Box>
    );
};
