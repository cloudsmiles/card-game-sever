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
import { QCardLogo } from '@/components/common';

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
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: (theme) =>
          theme.palette.mode === 'dark'
            ? 'linear-gradient(135deg, #1a1a2e 0%, #2a1a0e 25%, #1a1a2e 50%, #0e1a2a 75%, #1a1a2e 100%)'
            : 'linear-gradient(135deg, #f5f5f5 0%, #ffe8d6 25%, #f5f5f5 50%, #fff0e6 75%, #f5f5f5 100%)',
        backgroundSize: '300% 300%',
        animation: 'bgShift 8s ease infinite',
        '@keyframes bgShift': {
          '0%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
          '100%': { backgroundPosition: '0% 50%' },
        },
      }}
    >
      <Container maxWidth="sm">
        <Paper
          component={motion.div}
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ duration: 0.4, ease: 'easeOut' }}
          elevation={0}
          sx={{
            p: 4,
            width: '100%',
            borderRadius: 3,
            background: (theme) =>
              theme.palette.mode === 'dark'
                ? 'rgba(37, 37, 56, 0.6)'
                : 'rgba(255, 255, 255, 0.45)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            border: (theme) =>
              theme.palette.mode === 'dark'
                ? '1px solid rgba(255, 255, 255, 0.06)'
                : '1px solid rgba(255, 255, 255, 0.5)',
            boxShadow: (theme) =>
              theme.palette.mode === 'dark'
                ? '0 4px 24px rgba(0, 0, 0, 0.2)'
                : '0 4px 24px rgba(0, 0, 0, 0.04)',
          }}
        >
          <Box sx={{ textAlign: 'center', mb: 4 }}>
            <Box
              component={motion.div}
              initial={{ y: -20 }}
              animate={{ y: 0 }}
              sx={{ display: 'flex', justifyContent: 'center', mb: 1 }}
            >
              <QCardLogo size={56} />
            </Box>
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
      </Container>
    </Box>
  );
};
