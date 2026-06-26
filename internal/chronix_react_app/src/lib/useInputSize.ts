import {useTheme} from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';

export function useInputSize(): 'small' | 'medium' {
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up('md'));
  return isDesktop ? 'small' : 'medium';
}
