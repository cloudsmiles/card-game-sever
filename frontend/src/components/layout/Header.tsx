import { AppBar, Toolbar, IconButton, Box } from '@mui/material';
import { Brightness4, Brightness7 } from '@mui/icons-material';
import { useUIStore } from '@/stores/uiStore';
import { motion } from 'framer-motion';
import { QCardLogo } from '@/components/common';

export const Header: React.FC = () => {
  const { theme, toggleTheme } = useUIStore();

  return (
    <AppBar position="static" elevation={0} sx={{ borderRadius: 0 }}>
      <Toolbar>
        <Box
          component={motion.div}
          initial={{ x: -20, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          sx={{ flexGrow: 1 }}
        >
          <QCardLogo size={36} variant="light" />
        </Box>
        
        <IconButton
          component={motion.button}
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
          onClick={toggleTheme}
          color="inherit"
        >
          {theme === 'dark' ? <Brightness7 /> : <Brightness4 />}
        </IconButton>
      </Toolbar>
    </AppBar>
  );
};
