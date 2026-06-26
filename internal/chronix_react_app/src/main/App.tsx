import {Box} from "@mui/material";
import {MuiKit} from "@dsherwin/mui-kit";
import {MainContent} from "./MainContent.tsx";
import {useEffect, useRef, useState} from "react";
import {NavDrawer} from "./NavDrawer.tsx";
import {AnonymousContent} from "./AnonymousContent.tsx";
import {useAuthContext} from "../data/useAuthContext.ts";
import {ServerStatus, SiteState} from "../data/types.ts";
import {AdminLogin} from "../modules/Settings/AdminLogin.tsx";
import {LoadingScreen} from "./LoadingScreen.tsx";
import {ErrorScreen} from "./ErrorScreen.tsx";
import {ServerSuspendedScreen} from "./ServerSuspendedScreen.tsx";
import {NotificationsProvider} from "../data/NotificationsContext.tsx";
import {SseProvider} from "../data/SseContext.tsx";
import {RunsCommandsProvider} from "../data/RunNowContext.tsx";
import {GlobalRunProgressPanel} from "../modules/Runs/GlobalRunProgressPanel";
import {GlobalRunFinishedDialog} from "./GlobalRunFinishedDialog";
import {GlobalRunSnack} from "./GlobalRunSnack";
import {ConnectionStatusBanner} from "./ConnectionStatusBanner";
import {GlobalAgentRegistrationDialog} from "../modules/Agents/GlobalAgentRegistrationDialog";
import {RunsProvider} from "../data/RunsContext.tsx";
import {ConnectionsProvider} from "../data/ConnectionsContext.tsx";
import {ActionsProvider} from "../data/ActionsContext.tsx";
import {JobsProvider} from "../data/JobsContext.tsx";
import {FeatureAvailabilityProvider} from "../data/FeatureAvailabilityContext.tsx";
import {TopAppBar} from "./TopAppBar.tsx";
import {HelpProvider} from "../data/HelpContext.tsx";
import {GlobalHelpDialog} from "./GlobalHelpDialog.tsx";
import {MobileNavDrawer} from "./MobileNavDrawer.tsx";

export const App = () => {
    const appbarRef = useRef<HTMLDivElement>(null);
    const [appbarHeight, setAppbarHeight] = useState(0);
    const [mobileNavOpen, setMobileNavOpen] = useState(false);
    const {serverStatus, siteState, loggedIn} = useAuthContext();
    useEffect(() => {
        if (!appbarRef.current) return;
        const observer = new ResizeObserver(entries => {
            for (const entry of entries) {
                const height = entry.contentRect.height;
                setAppbarHeight(height);
            }
        })
        observer.observe(appbarRef.current);
        return () => {
            observer.disconnect();
        }
    }, [loggedIn]);

    switch (siteState) {
        case SiteState.LOADING:
            return <LoadingScreen/>;
        case SiteState.ERROR:
            return <ErrorScreen/>;
    }
    switch (serverStatus) {
        case ServerStatus.UNKNOWN:
            return <ErrorScreen/>;
        case ServerStatus.SUSPENDED:
            return <ServerSuspendedScreen/>;
        case ServerStatus.UNINITIALIZED:
            return <AdminLogin/>
    }


    return (<>
        <MuiKit>
            <HelpProvider>
                <Box sx={{display: 'flex', minHeight: '100vh'}}>
                    {!loggedIn && (<>
                        <AnonymousContent/>
                    </>)}
                    {loggedIn && (<>
                            <SseProvider>
                                <FeatureAvailabilityProvider>
                                    <ConnectionsProvider>
                                        <ActionsProvider>
                                            <JobsProvider>
                                                <RunsProvider>
                                                    <NotificationsProvider>
                                                        <RunsCommandsProvider>
                                                            <TopAppBar ref={appbarRef} onToggleNavigation={() => setMobileNavOpen((open) => !open)}/>
                                                            <MainContent appbarHeight={appbarHeight}/>
                                                            <GlobalRunProgressPanel/>
                                                            <GlobalRunFinishedDialog/>
                                                            <GlobalRunSnack/>
                                                            <GlobalAgentRegistrationDialog/>
                                                            <ConnectionStatusBanner/>
                                                            <GlobalHelpDialog/>
                                                        </RunsCommandsProvider>
                                                    </NotificationsProvider>
                                                </RunsProvider>
                                            </JobsProvider>
                                        </ActionsProvider>
                                    </ConnectionsProvider>
                                </FeatureAvailabilityProvider>
                            </SseProvider>
                        </>
                    )}
                </Box>
                {loggedIn && <NavDrawer appbarHeight={appbarHeight}/>}
                {loggedIn && <MobileNavDrawer open={mobileNavOpen} onClose={() => setMobileNavOpen(false)} appbarHeight={appbarHeight}/>}
            </HelpProvider>
        </MuiKit>
    </>);
};
