import {createContext} from 'react';
import type {ServerStatus, SiteState} from "./types.ts";

export interface AuthContextValue {
    serverStatus: typeof ServerStatus[keyof typeof ServerStatus];
    siteState: typeof SiteState[keyof typeof SiteState];
    loggedIn: boolean;
    setLoggedIn: (loggedIn: boolean) => void;
    logout: (reload?: boolean) => void;
    user?: { id: number; email: string; name: string; phone?: string; admin: boolean; forcePasswordChange: boolean; timeZone?: string; timeFormat?: '12h' | '24h' } | null;
    setUser: (u: AuthContextValue['user'] | ((prev: AuthContextValue['user']) => AuthContextValue['user'])) => void;
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);
