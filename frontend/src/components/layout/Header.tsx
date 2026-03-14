import { AppBar, Toolbar, Typography, IconButton, Box } from '@mui/material';
import { Brightness4, Brightness7 } from '@mui/icons-material';
import { useUIStore } from '@/stores/uiStore';
import { motion } from 'framer-motion';

export const Header: React.FC = () => {
  const { theme, toggleTheme } = useUIStore();

  return (
    <AppBar position="static" elevation={0}>
      <Toolbar>
        <Box
          component={motion.div}
          initial={{ x: -20, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          sx={{ flexGrow: 1 }}
        >
          <Typography variant="h5" component="h1" fontWeight="bold">
            QCard
          </Typography>
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
