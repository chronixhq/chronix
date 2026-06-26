import {Box, CircularProgress} from "@mui/material";
import {alpha, useTheme} from "@mui/material/styles";
import useMediaQuery from '@mui/material/useMediaQuery';
import {VStack} from "@dsherwin/mui-kit";
import {Dashboard} from "./Dashboard.tsx";
import {Suspense} from "react";
import {Navigate, Route, Routes, useLocation} from "react-router";
import {ModuleSideNav} from "./ModuleSideNav";
import {useFeatureAvailability} from "../data/FeatureAvailabilityContext.tsx";
import {APP_SHELL_PATHS, GLOBAL_RAIL_WIDTH, MODULE_SIDENAV_WIDTH, hasModuleSideNav} from "./appShellManifest.ts";
import {
    ActivityList,
    Actions,
    AddDatabaseConnection,
    AddWebtaskConnection,
    AdminOverview,
    AgentDetail,
    AgentsList,
    AlertsSettingsPage,
    AllConnectionsList,
    BrandingPage,
    BugReportPage,
    BugReportsList,
    CreateAction,
    CreateJob,
    CreateShellAction,
    CreateShellConnection,
    CreateWebtaskAction,
    DatabaseConnectionsList,
    EditAction,
    EditDatabaseConnection,
    EditJob,
    EditShellAction,
    EditShellConnection,
    EditUser,
    EditWebtaskAction,
    EditWebtaskConnection,
    EmailNotifierPage,
    FeatureRequestPage,
    FeatureRequestsList,
    HttpsSettingsPage,
    Jobs,
    NotificationsPage,
    ResetPassword,
    RunDetail,
    RunsList,
    ServerUrlPage,
    ShellConnectionsList,
    SmsNotifierPage,
    UpdatesPage,
    UserProfile,
    UsersAdmin,
    WebhooksPage,
    WebtaskConnectionsList,
} from "./lazyRoutes.tsx";

const RouteLoadingFallback = () => (
    <Box
        sx={{
            alignItems: 'center',
            display: 'flex',
            flexGrow: 1,
            justifyContent: 'center',
            minHeight: 240,
            py: 6,
            width: '100%',
        }}
    >
        <CircularProgress size={28}/>
    </Box>
);

export const MainContent = ({appbarHeight}: { appbarHeight: number }) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const location = useLocation();
    const {data: featureData} = useFeatureAvailability();
    const showModuleSideNav = hasModuleSideNav(location.pathname);
    return (<>
        <ModuleSideNav appbarHeight={appbarHeight}/>
        <VStack
            component="main"
            sx={{
                flexGrow: 1,
                marginTop: `${appbarHeight}px`,
                marginLeft: isMobile ? 0 : `${GLOBAL_RAIL_WIDTH + (showModuleSideNav ? MODULE_SIDENAV_WIDTH : 0)}px`,
                backgroundColor: theme.palette.background.default,
                backgroundImage: theme.palette.mode === 'dark'
                    ? `radial-gradient(circle at top right, ${alpha(theme.palette.primary.main, 0.12)} 0%, transparent 26%), radial-gradient(circle at bottom right, ${alpha('#7dd3fc', 0.08)} 0%, transparent 24%), linear-gradient(180deg, ${alpha('#132034', 0.46)} 0%, ${theme.palette.background.default} 22%, ${theme.palette.background.default} 100%)`
                    : `linear-gradient(180deg, ${alpha(theme.palette.primary.light, 0.08)} 0%, ${theme.palette.background.default} 18%, ${theme.palette.background.default} 100%)`,
                display: 'flex',
                minHeight: `calc(100vh - ${appbarHeight}px)`,
                position: 'relative',
                '&::before': {
                    content: '""',
                    position: 'absolute',
                    inset: 0,
                    pointerEvents: 'none',
                    background: `linear-gradient(180deg, ${alpha(theme.palette.common.white, theme.palette.mode === 'dark' ? 0.03 : 0.16)} 0%, transparent 18%)`,
                },
            }}
        >
            <Suspense fallback={<RouteLoadingFallback/>}>
                <Routes>
                    <Route index element={<Dashboard/>}/>
                    <Route path={"/activity"} element={<ActivityList/>}/>
                    <Route path={"/connections"}>
                        <Route index element={<Navigate to={"all"} replace/>}/>
                        <Route path={"all"} element={<AllConnectionsList/>}/>
                    </Route>
                    <Route path={"/databases"}>
                        <Route index element={<Navigate to={"list"} replace/>}/>
                        <Route path={"list"} element={<DatabaseConnectionsList/>}/>
                        <Route path={"add"} element={<AddDatabaseConnection/>}/>
                        <Route path={"edit/:id"} element={<EditDatabaseConnection/>}/>
                    </Route>
                    <Route path={"/shell"}>
                        <Route index element={<Navigate to={"list"} replace/>}/>
                        <Route path={"list"} element={<ShellConnectionsList/>}/>
                        <Route path={"add"} element={<CreateShellConnection/>}/>
                        <Route path={"edit/:id"} element={<EditShellConnection/>}/>
                    </Route>
                    <Route path={"/webtasks"}>
                        <Route index element={<Navigate to={"list"} replace/>}/>
                        <Route path={"list"} element={<WebtaskConnectionsList/>}/>
                        <Route path={"add"} element={<AddWebtaskConnection/>}/>
                        <Route path={"edit/:id"} element={<EditWebtaskConnection/>}/>
                    </Route>
                    <Route path={"/actions"}>
                        <Route index element={<Navigate to={"list"} replace/>}/>
                        <Route path={"list"} element={<Actions/>}/>
                        <Route path={"create"} element={<CreateAction/>}/>
                        <Route path={"create-shell"} element={<CreateShellAction/>}/>
                        <Route path={"create-webtask"} element={<CreateWebtaskAction/>}/>
                        <Route path={"edit/:id"} element={<EditAction/>}/>
                        <Route path={"edit-shell/:id"} element={<EditShellAction/>}/>
                        <Route path={"edit-webtask/:id"} element={<EditWebtaskAction/>}/>
                    </Route>
                    <Route path={"/jobs"}>
                        <Route index element={<Navigate to={"list"} replace/>}/>
                        <Route path={"list"} element={<Jobs/>}/>
                        <Route path={"create"} element={<CreateJob/>}/>
                        <Route path={"edit/:id"} element={<EditJob/>}/>
                    </Route>
                    <Route path={"/settings"}>
                        <Route index element={<Navigate to={"overview"} replace/>}/>
                        <Route path={"overview"} element={<AdminOverview/>}/>
                        <Route path={"server-url"} element={<ServerUrlPage/>}/>
                        <Route path={"email"} element={<EmailNotifierPage/>}/>
                        <Route path={"https"} element={<HttpsSettingsPage/>}/>
                        <Route path={"sms"} element={<SmsNotifierPage/>}/>
                        <Route path={"alerts"} element={<AlertsSettingsPage/>}/>
                        <Route path={"users"} element={<UsersAdmin/>}/>
                        <Route path={"users/edit"} element={<EditUser/>}/>
                        <Route path={"branding"} element={<BrandingPage/>}/>
                        <Route path={"webhooks"} element={<WebhooksPage/>}/>
                        <Route path={"updates"} element={<UpdatesPage/>}/>
                        {featureData?.feedbackEnabled && (
                            <>
                                <Route path={"bug-reports"} element={<BugReportsList/>}/>
                                <Route path={"feature-requests"} element={<FeatureRequestsList/>}/>
                            </>
                        )}
                    </Route>
                    <Route path={"/user"}>
                        <Route index element={<Navigate to={APP_SHELL_PATHS.userProfile} replace/>}/>
                        <Route path={"reset"} element={<ResetPassword/>}/>
                        <Route path={"profile"} element={<UserProfile/>}/>
                    </Route>
                    <Route path={"/runs"}>
                        <Route index element={<RunsList/>}/>
                        <Route path={":runId"} element={<RunDetail/>}/>
                    </Route>
                    <Route path={"/agents"}>
                        <Route index element={<AgentsList/>}/>
                        <Route path={":uuid"} element={<AgentDetail/>}/>
                    </Route>
                    <Route path={"/notifications"} element={<NotificationsPage/>}/>
                    {featureData?.feedbackEnabled && (
                        <>
                            <Route path={"/bug-report"} element={<BugReportPage/>}/>
                            <Route path={"/feature-request"} element={<FeatureRequestPage/>}/>
                        </>
                    )}
                </Routes>
            </Suspense>
        </VStack>
    </>);
}
