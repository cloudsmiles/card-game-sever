import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Container,
  TextField,
  Button,
  Typography,
  Paper,
  CircularProgress,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useUIStore } from '@/stores/uiStore';

export const ConnectPage: React.FC = () => {
  const [nickname, setNickname] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const { connect } = useWebSocket();
  const { addNotification } = useUIStore();
  const navigate = useNavigate();

  const handleConnect = async () => {
    // Validate nickname
    if (!nickname.trim()) {
      addNotification({
        type: 'error',
        message: '请输入昵称',
      });
      return;
    }

    if (nickname.length > 20) {
      addNotification({
        type: 'error',
        message: '昵称不能超过20个字符',
      });
      return;
    }

    try {
      setIsConnecting(true);
      
      // Generate unique player ID
      const playerId = `player_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      
      await connect(playerId, nickname.trim());
      
      addNotification({
        type: 'success',
        message: '连接成功！',
      });
      
      navigate('/lobby');
    } catch (error: any) {
      addNotification({
        type: 'error',
        message: error.message || '连接失败，请重试',
      });
    } finally {
      setIsConnecting(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !isConnecting) {
      handleConnect();
    }
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Paper
          component={motion.div}
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.3 }}
          elevation={3}
          sx={{
            p: 4,
            width: '100%',
            borderRadius: 3,
          }}
        >
          <Box sx={{ textAlign: 'center', mb: 4 }}>
            <Typography
              variant="h3"
              component={motion.h1}
              initial={{ y: -20 }}
              animate={{ y: 0 }}
              sx={{ fontWeight: 'bold', mb: 1 }}
            >
              QCard
            </Typography>
            <Typography variant="body1" color="text.secondary">
              多人卡牌游戏平台
            </Typography>
          </Box>

          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
            <TextField
              fullWidth
              label="昵称"
              placeholder="请输入您的昵称 (1-20字符)"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              onKeyPress={handleKeyPress}
              disabled={isConnecting}
              inputProps={{ maxLength: 20 }}
              helperText={`${nickname.length}/20`}
            />

            <Button
              component={motion.button}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              fullWidth
              variant="contained"
              size="large"
              onClick={handleConnect}
              disabled={isConnecting || !nickname.trim()}
              sx={{ py: 1.5 }}
            >
              {isConnecting ? (
                <>
                  <CircularProgress size={24} sx={{ mr: 1 }} />
                  连接中...
                </>
              ) : (
                '进入大厅'
              )}
            </Button>
          </Box>
        </Paper>
      </Box>
    </Container>
  );
};
