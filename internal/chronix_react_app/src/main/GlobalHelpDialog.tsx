import {HelpDialog} from "./HelpDialog.tsx";
import {useHelp} from "../data/HelpContext.tsx";

export const GlobalHelpDialog = () => {
    const {isOpen, closeHelp, section} = useHelp();
    return <HelpDialog open={isOpen} onClose={closeHelp} section={section}/>;
};
