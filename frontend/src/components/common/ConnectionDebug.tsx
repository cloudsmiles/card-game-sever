import { Box, Paper, Typography, Chip, IconButton } from '@mui/material';
import { BugReport, Close } from '@mui/icons-material';
import { useConnectionStore } from '@/stores/connectionStore';
import { useUIStore } from '@/stores/uiStore';

export const ConnectionDebug: React.FC = () => {
  const { status, playerId, nickname, error, reconnectAttempts } = useConnectionStore();
  const { debugOpen, toggleDebug } = useUIStore();

  if (process.env.NODE_ENV === 'production') {
    return null;
  }

  // 收起状态：只显示一个小图标
  if (!debugOpen) {
    return (
      <IconButton
        onClick={toggleDebug}
        size="small"
        sx={{
          position: 'fixed',
          bottom: 16,
          left: 16,
          zIndex: 9999,
          bgcolor: 'rgba(0,0,0,0.3)',
          color: status === 'connected' ? 'success.main' : 'error.main',
          '&:hover': { bgcolor: 'rgba(0,0,0,0.5)' },
          width: 32,
          height: 32,
        }}
      >
        <BugReport fontSize="small" />
      </IconButton>
    );
  }

  return (
    <Paper
      sx={{
        position: 'fixed',
        bottom: 16,
        left: 16,
        p: 2,
        zIndex: 9999,
        minWidth: 280,
        bgcolor: 'background.paper',
        border: 1,
        borderColor: 'divider',
      }}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
        <Typography variant="caption" fontWeight="bold">
          WebSocket Debug
        </Typography>
        <IconButton size="small" onClick={toggleDebug}>
          <Close fontSize="small" />
        </IconButton>
      </Box>

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
          <Typography variant="caption">Player ID: {playerId}</Typography>
        )}

        {nickname && (
          <Typography variant="caption">Nickname: {nickname}</Typography>
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
