import React, {createContext, type ReactNode, useContext, useState} from 'react';

interface HelpContextType {
    openHelp: (section?: string) => void;
    closeHelp: () => void;
    isOpen: boolean;
    section: string;
}

const HelpContext = createContext<HelpContextType | undefined>(undefined);

export const HelpProvider: React.FC<{children: ReactNode}> = ({children}) => {
    const [isOpen, setIsOpen] = useState(false);
    const [section, setSection] = useState('');

    const openHelp = (sec?: string) => {
        setSection(sec || '');
        setIsOpen(true);
    };

    const closeHelp = () => {
        setIsOpen(false);
        setSection('');
    };

    return (
        <HelpContext.Provider value={{openHelp, closeHelp, isOpen, section}}>
            {children}
        </HelpContext.Provider>
    );
};

export const useHelp = () => {
    const context = useContext(HelpContext);
    if (context === undefined) {
        throw new Error('useHelp must be used within a HelpProvider');
    }
    return context;
};
