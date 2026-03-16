import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Container,
  Typography,
  Button,
  Paper,
  IconButton,
  Tooltip,
} from '@mui/material';
import { ContentCopy, ExitToApp, SmartToy } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useRoomStore } from '@/stores/roomStore';
import { useChatStore } from '@/stores/chatStore';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useUIStore } from '@/stores/uiStore';
import { PlayerList, ReadyButton } from '@/components/room';
import { DDZGame } from '@/games/ddz';
import { MahjongGame } from '@/games/mahjong';
import { DurianGame } from '@/games/durian';

const gameTypeNames: Record<string, string> = {
  ddz: '斗地主',
  mahjong: '麻将',
  durian: '榴莲忘返',
};

export const RoomPage: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const { currentRoomId, gameType, roomStatus, players, maxPlayers, leaveRoom } = useRoomStore();
  const { clearMessages } = useChatStore();
  const { leaveRoom: wsLeaveRoom, addBot, kickBot } = useWebSocket();
  const { addNotification } = useUIStore();
  const navigate = useNavigate();

  useEffect(() => {
    // If no room ID in URL, or we haven't actually joined this room, redirect to lobby
    if (!roomId || !currentRoomId || currentRoomId !== roomId) {
      navigate('/lobby');
    }
  }, [roomId, currentRoomId, navigate]);

  const handleLeave = () => {
    if (roomId) {
      wsLeaveRoom(roomId);
    }
    // Clear local state immediately for responsive UI
    leaveRoom();
    clearMessages();
    navigate('/lobby');
  };

  const handleCopyRoomId = () => {
    if (roomId) {
      navigator.clipboard.writeText(roomId);
      addNotification({
        type: 'success',
        message: '房间号已复制',
      });
    }
  };

  const allPlayersReady = players.length > 0 && players.every((p) => p.ready);
  const hasEnoughPlayers = players.length >= 2; // Minimum 2 players
  const canAddBot = players.length < maxPlayers && roomStatus === 'waiting';

  const handleAddBot = () => {
    if (roomId) {
      addBot(roomId);
    }
  };

  const handleKickBot = (botId: string) => {
    if (roomId) {
      kickBot(roomId, botId);
    }
  };

  return (
    <Container maxWidth="lg">
      <Box sx={{ py: 4 }}>
        <Box
          component={motion.div}
          initial={{ y: -20, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2 }}
        >
          <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Typography variant="h4" fontWeight="bold">
                {gameType ? gameTypeNames[gameType] : '房间'}
              </Typography>
              <Tooltip title="复制房间号">
                <IconButton size="small" onClick={handleCopyRoomId}>
                  <ContentCopy fontSize="small" />
                </IconButton>
              </Tooltip>
            </Box>
            <Typography variant="body2" color="text.secondary">
              房间号: {roomId}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              状态: {roomStatus === 'waiting' ? '等待中' : '游戏中'}
            </Typography>
          </Box>

          <Button variant="outlined" startIcon={<ExitToApp />} onClick={handleLeave}>
            离开房间
          </Button>
        </Box>

        {roomStatus === 'waiting' ? (
          <Box sx={{ display: 'flex', gap: 3, flexDirection: { xs: 'column', md: 'row' } }}>
            <Box sx={{ flex: 1 }}>
              <PlayerList players={players} maxPlayers={maxPlayers} roomStatus={roomStatus} onKickBot={handleKickBot} />
            </Box>

            <Box sx={{ width: { xs: '100%', md: 300 } }}>
              <Paper
                component={motion.div}
                initial={{ x: 20, opacity: 0 }}
                animate={{ x: 0, opacity: 1 }}
                sx={{ p: 3, position: 'sticky', top: 80 }}
              >
                <Typography variant="h6" fontWeight="bold" sx={{ mb: 3 }}>
                  准备状态
                </Typography>

                <Box sx={{ mb: 3 }}>
                  <ReadyButton />
                </Box>

                {canAddBot && (
                  <Button
                    fullWidth
                    variant="outlined"
                    startIcon={<SmartToy />}
                    onClick={handleAddBot}
                    sx={{ mb: 3 }}
                  >
                    添加机器人
                  </Button>
                )}

                {!hasEnoughPlayers && (
                  <Typography variant="body2" color="warning.main" sx={{ textAlign: 'center' }}>
                    等待更多玩家加入...
                  </Typography>
                )}

                {hasEnoughPlayers && !allPlayersReady && (
                  <Typography variant="body2" color="info.main" sx={{ textAlign: 'center' }}>
                    等待所有玩家准备...
                  </Typography>
                )}

                {hasEnoughPlayers && allPlayersReady && (
                  <Typography variant="body2" color="success.main" sx={{ textAlign: 'center' }}>
                    游戏即将开始！
                  </Typography>
                )}
              </Paper>
            </Box>
          </Box>
        ) : (
          <Box>
            {gameType === 'ddz' && <DDZGame />}
            {gameType === 'mahjong' && <MahjongGame />}
            {gameType === 'durian' && <DurianGame />}
          </Box>
        )}
      </Box>
    </Container>
  );
};
