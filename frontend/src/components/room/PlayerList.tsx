import { Box, Paper, Typography, Avatar, Chip } from '@mui/material';
import { CheckCircle, Cancel, SignalWifiOff } from '@mui/icons-material';
import { motion } from 'framer-motion';
import type { SeatPlayerInfo } from '@/types/room';
import { useUserStore } from '@/stores/userStore';

interface PlayerListProps {
  players: SeatPlayerInfo[];
  maxPlayers: number;
}

export const PlayerList: React.FC<PlayerListProps> = ({ players, maxPlayers }) => {
  const { playerId: currentPlayerId } = useUserStore();

  // Create array with all seats
  const seats = Array.from({ length: maxPlayers }, (_, i) => {
    return players.find((p) => p.seat === i) || null;
  });

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <Typography variant="h6" fontWeight="bold">
        玩家列表 ({players.length}/{maxPlayers})
      </Typography>

      {seats.map((player, index) => (
        <Paper
          key={index}
          component={motion.div}
          initial={{ x: -20, opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          transition={{ delay: index * 0.1 }}
          sx={{
            p: 2,
            display: 'flex',
            alignItems: 'center',
            gap: 2,
            bgcolor: player?.player_id === currentPlayerId ? 'action.selected' : 'background.paper',
            border: player?.player_id === currentPlayerId ? 2 : 0,
            borderColor: 'primary.main',
          }}
        >
          <Avatar
            sx={{
              bgcolor: player ? 'primary.main' : 'action.disabled',
              width: 48,
              height: 48,
            }}
          >
            {player ? player.nickname.charAt(0).toUpperCase() : index + 1}
          </Avatar>

          <Box sx={{ flexGrow: 1 }}>
            <Typography variant="body1" fontWeight={player ? 'bold' : 'normal'}>
              {player ? player.nickname : `座位 ${index + 1}`}
              {player?.player_id === currentPlayerId && (
                <Chip label="你" size="small" color="primary" sx={{ ml: 1 }} />
              )}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              座位 {index + 1}
            </Typography>
          </Box>

          {player && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              {player.offline ? (
                <Chip
                  icon={<SignalWifiOff />}
                  label="离线"
                  size="small"
                  color="error"
                  variant="outlined"
                />
              ) : player.ready ? (
                <Chip
                  icon={<CheckCircle />}
                  label="已准备"
                  size="small"
                  color="success"
                  variant="outlined"
                />
              ) : (
                <Chip
                  icon={<Cancel />}
                  label="未准备"
                  size="small"
                  color="default"
                  variant="outlined"
                />
              )}
            </Box>
          )}
        </Paper>
      ))}
    </Box>
  );
};
