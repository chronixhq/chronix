import {type ReactNode, useEffect, useState} from 'react';
import {AuthContext} from './context';
import {apiGet, enableAPILog, setAPIBaseURL, setIncludeCredentials} from "@dsherwin/react-api-interface";
import {baseURL} from '../lib/api_config';
import {ServerStatus, SiteState} from "./types.ts";
import {setDefaultTimeZone} from '../lib/dayjs';
import {setGlobalDisplayOptions} from '../lib/datetime';
import {useNavigate} from "react-router";

setAPIBaseURL(baseURL);
enableAPILog(false);
setIncludeCredentials(true);

interface ProviderProps {
    children: ReactNode;
}

const sleep = (ms: number): Promise<void> => (new Promise(resolve => setTimeout(resolve, ms)));

export function AuthContextProvider({children}: ProviderProps) {
    const navigate = useNavigate();
    const [serverStatus, setServerStatus] = useState(ServerStatus.UNKNOWN);
    const [siteState, setSiteState] = useState(SiteState.LOADING);
    const [loggedIn, setLoggedIn] = useState(false);
    const [user, setUser] = useState<{ id: number; email: string; name: string; phone?: string; admin: boolean; forcePasswordChange: boolean; timeZone?: string; timeFormat?: '12h' | '24h' } | null | undefined>(undefined);

    // keep dayjs and global display options in sync when user changes
    useEffect(() => {
        const tz = user?.timeZone;
        const tf = user?.timeFormat;
        try {
            setDefaultTimeZone(tz || undefined);
        } catch {
        }
        try {
            setGlobalDisplayOptions({timeZone: tz, hour12: tf === '12h' ? true : tf === '24h' ? false : undefined});
        } catch {
        }
    }, [user]);

    useEffect(() => {
        (async () => {
            while (siteState != SiteState.READY) {
                try {
                    const res: { status: string } = await apiGet("/server/status");
                    if (res.status != ServerStatus.STARTINGUP) {
                        setServerStatus(res.status);
                        try {
                            await apiGet("/checkauth")
                            setLoggedIn(true);
                        } catch {
                            setLoggedIn(false);
                            setUser(null);
                        }
                        setSiteState(SiteState.READY);
                        break;
                    }
                } catch {
                    setSiteState(SiteState.ERROR);
                }
                await sleep(5000);
            }
            // After site is ready and if logged in, fetch current user
            if (loggedIn) {
                try {
                    const me = await apiGet('/me') as { id: number; email: string; name: string; phone?: string; admin: boolean; forcePasswordChange: boolean; timeZone?: string; timeFormat?: '12h' | '24h' }
                    setUser(me)
                } catch (e) {
                    console.error(e);
                    setUser(null)
                }
            } else {
                setUser(null)
            }
        })();
    }, []); // eslint-disable-line react-hooks/exhaustive-deps -- [deps-intentional] bootstrap on mount; subsequent user fetch handled in separate effect

    useEffect(() => {
        // whenever login state changes, refresh user info
        (async () => {
            if (loggedIn) {
                try {
                    const me = await apiGet('/me') as { id: number; email: string; name: string; phone?: string; admin: boolean; forcePasswordChange: boolean; timeZone?: string; timeFormat?: '12h' | '24h' }
                    setUser(me)
                } catch (e) {
                    console.error(e);
                    setUser(null)
                }
            } else {
                setUser(null)
            }
        })()
    }, [loggedIn])

    const logout = (reload?: boolean) => {
        (async () => {
            try {
                await apiGet('/logout');
            } catch (e) {
                console.error(e);
            }
            setLoggedIn(false);
            setUser(null);
            if (reload) {
                window.location.href = '/';
            } else {
                navigate('/');
            }
        })()
    }

    return <AuthContext.Provider value={{
        serverStatus,
        siteState,
        loggedIn,
        setLoggedIn,
        logout,
        user,
        setUser,
    }}>{children}</AuthContext.Provider>;
}
