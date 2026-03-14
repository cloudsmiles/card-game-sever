import { useState, useEffect } from 'react';
import {
  Box,
  Container,
  Typography,
  Button,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  ToggleButtonGroup,
  ToggleButton,
} from '@mui/material';
import { Add, Login, Logout } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useUserStore } from '@/stores/userStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUIStore } from '@/stores/uiStore';
import { RoomList } from '@/components/room';
import type { GameType } from '@/types/room';

export const LobbyPage: React.FC = () => {
  const { nickname } = useUserStore();
  const { currentRoomId } = useRoomStore();
  const { disconnect, createRoom, joinRoom } = useWebSocket();
  const { addNotification } = useUIStore();
  const navigate = useNavigate();

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [joinDialogOpen, setJoinDialogOpen] = useState(false);
  const [selectedGameType, setSelectedGameType] = useState<GameType>('ddz');
  const [roomIdInput, setRoomIdInput] = useState('');

  // Auto-navigate to room when joined
  useEffect(() => {
    if (currentRoomId) {
      navigate(`/room/${currentRoomId}`);
    }
  }, [currentRoomId, navigate]);

  const handleDisconnect = () => {
    disconnect();
    navigate('/');
  };

  const handleCreateRoom = () => {
    createRoom(selectedGameType);
    setCreateDialogOpen(false);
    addNotification({
      type: 'info',
      message: '正在创建房间...',
    });
  };

  const handleJoinRoom = (roomId: string) => {
    joinRoom(roomId);
    addNotification({
      type: 'info',
      message: '正在加入房间...',
    });
  };

  const handleJoinByRoomId = () => {
    if (!roomIdInput.trim()) {
      addNotification({
        type: 'error',
        message: '请输入房间号',
      });
      return;
    }
    handleJoinRoom(roomIdInput.trim());
    setJoinDialogOpen(false);
    setRoomIdInput('');
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
            <Typography variant="h4" fontWeight="bold">
              游戏大厅
            </Typography>
            <Typography variant="body2" color="text.secondary">
              欢迎, {nickname}
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', gap: 2 }}>
            <Button
              variant="outlined"
              startIcon={<Logout />}
              onClick={handleDisconnect}
            >
              断开连接
            </Button>
          </Box>
        </Box>

        <Paper
          component={motion.div}
          initial={{ y: 20, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          sx={{ p: 3, mb: 4 }}
        >
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            <Button
              variant="contained"
              size="large"
              startIcon={<Add />}
              onClick={() => setCreateDialogOpen(true)}
            >
              创建房间
            </Button>

            <Button
              variant="outlined"
              size="large"
              startIcon={<Login />}
              onClick={() => setJoinDialogOpen(true)}
            >
              加入房间
            </Button>
          </Box>
        </Paper>

        <RoomList onJoinRoom={handleJoinRoom} />
      </Box>

      {/* Create Room Dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>创建房间</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              选择游戏类型
            </Typography>
            <ToggleButtonGroup
              value={selectedGameType}
              exclusive
              onChange={(_, value) => value && setSelectedGameType(value)}
              fullWidth
              orientation="vertical"
            >
              <ToggleButton value="ddz">斗地主</ToggleButton>
              <ToggleButton value="mahjong">麻将</ToggleButton>
              <ToggleButton value="durian">榴莲忘返</ToggleButton>
            </ToggleButtonGroup>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>取消</Button>
          <Button onClick={handleCreateRoom} variant="contained">
            创建
          </Button>
        </DialogActions>
      </Dialog>

      {/* Join Room Dialog */}
      <Dialog open={joinDialogOpen} onClose={() => setJoinDialogOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>加入房间</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            fullWidth
            label="房间号"
            placeholder="请输入6位数字房间号"
            value={roomIdInput}
            onChange={(e) => setRoomIdInput(e.target.value.replace(/\D/g, '').slice(0, 6))}
            onKeyPress={(e) => e.key === 'Enter' && handleJoinByRoomId()}
            inputProps={{ inputMode: 'numeric', pattern: '[0-9]*' }}
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setJoinDialogOpen(false)}>取消</Button>
          <Button onClick={handleJoinByRoomId} variant="contained">
            加入
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};
