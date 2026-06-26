import {IconButton, Tooltip} from "@mui/material";
import HelpOutlinedIcon from "@mui/icons-material/HelpOutlined";
import {useHelp} from "../data/HelpContext.tsx";

export const SectionHelp = ({section}: {section?: string}) => {
    const {openHelp} = useHelp();
    return (
        <Tooltip title="Help">
            <IconButton size="small" color="warning" onClick={() => openHelp(section)} sx={{ml: 0.5}}>
                <HelpOutlinedIcon fontSize="small" />
            </IconButton>
        </Tooltip>
    );
};
