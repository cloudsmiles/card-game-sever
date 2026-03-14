import { Box, Paper, Typography, Chip } from '@mui/material';
import { useConnectionStore } from '@/stores/connectionStore';

export const ConnectionDebug: React.FC = () => {
  const { status, playerId, nickname, error, reconnectAttempts } = useConnectionStore();

  if (process.env.NODE_ENV === 'production') {
    return null;
  }

  return (
    <Paper
      sx={{
        position: 'fixed',
        bottom: 16,
        left: 16,
        p: 2,
        zIndex: 9999,
        minWidth: 300,
        bgcolor: 'background.paper',
        border: 1,
        borderColor: 'divider',
      }}
    >
      <Typography variant="caption" fontWeight="bold" display="block" mb={1}>
        WebSocket Debug
      </Typography>
      
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="caption">Status:</Typography>
          <Chip
            label={status}
            size="small"
            color={
              status === 'connected' ? 'success' :
              status === 'connecting' ? 'info' :
              status === 'reconnecting' ? 'warning' :
              'error'
            }
          />
        </Box>

        {playerId && (
          <Typography variant="caption">
            Player ID: {playerId}
          </Typography>
        )}

        {nickname && (
          <Typography variant="caption">
            Nickname: {nickname}
          </Typography>
        )}

        {reconnectAttempts > 0 && (
          <Typography variant="caption" color="warning.main">
            Reconnect attempts: {reconnectAttempts}
          </Typography>
        )}

        {error && (
          <Typography variant="caption" color="error.main">
            Error: {error}
          </Typography>
        )}

        <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
          WS URL: {(import.meta as any).env?.VITE_WS_URL || 'ws://localhost:8080/ws'}
        </Typography>
      </Box>
    </Paper>
  );
};
