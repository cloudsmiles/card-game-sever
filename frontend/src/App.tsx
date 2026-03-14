import { useMemo } from 'react';
import { RouterProvider } from 'react-router-dom';
import { ThemeProvider, CssBaseline, Box } from '@mui/material';
import { createAppTheme } from './theme';
import { useUIStore } from './stores/uiStore';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { NotificationManager, ConnectionDebug } from './components/common';
import { Header } from './components/layout';
import { ChatPanel } from './components/chat/ChatPanel';
import { router } from './router';

function App() {
  const { theme: themeMode } = useUIStore();
  const theme = useMemo(() => createAppTheme(themeMode), [themeMode]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <WebSocketProvider>
        <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
          <Header />
          <Box component="main" sx={{ flexGrow: 1 }}>
            <RouterProvider router={router} />
          </Box>
        </Box>
        <ChatPanel roomOnly={true} />
        <ConnectionDebug />
        <NotificationManager />
      </WebSocketProvider>
    </ThemeProvider>
  );
}

export default App;
