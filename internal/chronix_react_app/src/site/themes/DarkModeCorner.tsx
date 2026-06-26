import Box from '@mui/material/Box';
import ThemeToggle from './ThemeToggle';

const DarkModeCorner = () => {
  return (
    <Box sx={{ position: 'fixed', top: 12, right: 12, zIndex: (theme) => theme.zIndex.tooltip + 1 }}>
      <ThemeToggle />
    </Box>
  );
};

export default DarkModeCorner;
