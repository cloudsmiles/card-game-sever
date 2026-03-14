import { Button } from '@mui/material';
import { CheckCircle, Cancel } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';

export const ReadyButton: React.FC = () => {
  const { currentRoomId, players, roomStatus } = useRoomStore();
  const { playerId } = useUserStore();
  const { setReady } = useWebSocket();

  const currentPlayer = players.find((p) => p.player_id === playerId);
  const isReady = currentPlayer?.ready || false;
  const isDisabled = roomStatus === 'playing';

  const handleToggleReady = () => {
    if (currentRoomId) {
      setReady(currentRoomId, !isReady);
    }
  };

  return (
    <Button
      component={motion.button}
      whileHover={{ scale: isDisabled ? 1 : 1.02 }}
      whileTap={{ scale: isDisabled ? 1 : 0.98 }}
      fullWidth
      variant={isReady ? 'outlined' : 'contained'}
      color={isReady ? 'warning' : 'success'}
      size="large"
      startIcon={isReady ? <Cancel /> : <CheckCircle />}
      onClick={handleToggleReady}
      disabled={isDisabled}
      sx={{ py: 1.5 }}
    >
      {isReady ? '取消准备' : '准备'}
    </Button>
  );
};
