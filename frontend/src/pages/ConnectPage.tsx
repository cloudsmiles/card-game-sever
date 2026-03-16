import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Container,
  TextField,
  Button,
  Typography,
  Paper,
  CircularProgress,
  Divider,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useUIStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import { authService } from '@/services/authService';
import { QCardLogo } from '@/components/common';

export const ConnectPage: React.FC = () => {
  const [nickname, setNickname] = useState(() => {
    // 优先从 authStore 读取已登录用户的昵称
    const authStored = JSON.parse(localStorage.getItem('auth-storage') || '{}');
    if (authStored?.state?.user?.nickname) {
      return authStored.state.user.nickname;
    }
    // 兼容旧的 user-storage
    const stored = JSON.parse(localStorage.getItem('user-storage') || '{}');
    return stored?.state?.nickname || '';
  });
  const [isConnecting, setIsConnecting] = useState(false);
  const { connect } = useWebSocket();
  const { addNotification } = useUIStore();
  const { setAuth, isLoggedIn, token, user } = useAuthStore();
  const navigate = useNavigate();

  // 处理微信回调（URL 中带 code 参数）
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state'); // "h5" 或 "open"
    if (code) {
      handleWeChatCallback(code, (state as 'h5' | 'open') || 'h5');
      // 清除 URL 中的 code 参数
      window.history.replaceState({}, '', window.location.pathname);
    }
  }, []);

  // 微信回调处理
  const handleWeChatCallback = async (code: string, platform: 'h5' | 'open') => {
    try {
      setIsConnecting(true);
      const loginResp = await authService.wechatLogin(code, platform);
      setAuth(loginResp.token, loginResp.user);

      const playerId = `user_${loginResp.user.id}`;
      await connect(playerId, loginResp.user.nickname, loginResp.token);

      addNotification({ type: 'success', message: '微信登录成功！' });
      navigate('/lobby');
    } catch (error: any) {
      addNotification({ type: 'error', message: error.message || '微信登录失败' });
    } finally {
      setIsConnecting(false);
    }
  };

  // 游客登录（通过 API）
  const handleGuestLogin = async () => {
    if (!nickname.trim()) {
      addNotification({ type: 'error', message: '请输入昵称' });
      return;
    }
    if (nickname.length > 20) {
      addNotification({ type: 'error', message: '昵称不能超过20个字符' });
      return;
    }

    try {
      setIsConnecting(true);

      // 尝试通过 API 登录（如果后端有数据库）
      try {
        const loginResp = await authService.guestLogin(nickname.trim());
        setAuth(loginResp.token, loginResp.user);
        const playerId = `user_${loginResp.user.id}`;
        await connect(playerId, loginResp.user.nickname, loginResp.token);
      } catch {
        // API 不可用时，回退到旧的直连模式
        const stored = JSON.parse(localStorage.getItem('user-storage') || '{}');
        const savedPlayerId = stored?.state?.playerId;
        const savedNickname = stored?.state?.nickname;
        const playerId = (savedPlayerId && savedNickname === nickname.trim())
          ? savedPlayerId
          : `player_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        await connect(playerId, nickname.trim());
      }

      addNotification({ type: 'success', message: '连接成功！' });
      navigate('/lobby');
    } catch (error: any) {
      addNotification({ type: 'error', message: error.message || '连接失败，请重试' });
    } finally {
      setIsConnecting(false);
    }
  };

  // 已登录用户快速进入
  const handleQuickConnect = async () => {
    if (!isLoggedIn || !token || !user) return;
    try {
      setIsConnecting(true);
      const playerId = `user_${user.id}`;
      await connect(playerId, user.nickname, token);
      addNotification({ type: 'success', message: '连接成功！' });
      navigate('/lobby');
    } catch (error: any) {
      addNotification({ type: 'error', message: error.message || '连接失败' });
    } finally {
      setIsConnecting(false);
    }
  };

  // 微信 H5 授权跳转
  const handleWeChatH5 = () => {
    const appId = (import.meta as any).env?.VITE_WECHAT_H5_APPID;
    if (!appId) {
      addNotification({ type: 'error', message: '微信 H5 登录未配置' });
      return;
    }
    const redirectUri = encodeURIComponent(window.location.origin + window.location.pathname);
    const url = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${appId}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_userinfo&state=h5#wechat_redirect`;
    window.location.href = url;
  };

  // 微信 PC 扫码登录跳转
  const handleWeChatOpen = () => {
    const appId = (import.meta as any).env?.VITE_WECHAT_OPEN_APPID;
    if (!appId) {
      addNotification({ type: 'error', message: '微信扫码登录未配置' });
      return;
    }
    const redirectUri = encodeURIComponent(window.location.origin + window.location.pathname);
    const url = `https://open.weixin.qq.com/connect/qrconnect?appid=${appId}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_login&state=open#wechat_redirect`;
    window.location.href = url;
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !isConnecting) {
      handleGuestLogin();
    }
  };

  const hasWeChatConfig = !!(
    (import.meta as any).env?.VITE_WECHAT_H5_APPID ||
    (import.meta as any).env?.VITE_WECHAT_OPEN_APPID
  );

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

          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {/* 已登录用户快速进入 */}
            {isLoggedIn && user && (
              <>
                <Button
                  component={motion.button}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  fullWidth
                  variant="contained"
                  size="large"
                  onClick={handleQuickConnect}
                  disabled={isConnecting}
                  sx={{ py: 1.5 }}
                >
                  {isConnecting ? (
                    <>
                      <CircularProgress size={24} sx={{ mr: 1 }} />
                      连接中...
                    </>
                  ) : (
                    `${user.nickname} - 快速进入`
                  )}
                </Button>
                <Divider sx={{ my: 1 }}>
                  <Typography variant="caption" color="text.secondary">或</Typography>
                </Divider>
              </>
            )}

            {/* 游客登录 */}
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
              onClick={handleGuestLogin}
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

            {/* 微信登录 */}
            {hasWeChatConfig && (
              <>
                <Divider sx={{ my: 1 }}>
                  <Typography variant="caption" color="text.secondary">第三方登录</Typography>
                </Divider>
                <Box sx={{ display: 'flex', gap: 2 }}>
                  {(import.meta as any).env?.VITE_WECHAT_H5_APPID && (
                    <Button
                      fullWidth
                      variant="outlined"
                      onClick={handleWeChatH5}
                      disabled={isConnecting}
                      sx={{ color: '#07C160', borderColor: '#07C160', '&:hover': { borderColor: '#06AD56', bgcolor: 'rgba(7,193,96,0.04)' } }}
                    >
                      微信登录
                    </Button>
                  )}
                  {(import.meta as any).env?.VITE_WECHAT_OPEN_APPID && (
                    <Button
                      fullWidth
                      variant="outlined"
                      onClick={handleWeChatOpen}
                      disabled={isConnecting}
                      sx={{ color: '#07C160', borderColor: '#07C160', '&:hover': { borderColor: '#06AD56', bgcolor: 'rgba(7,193,96,0.04)' } }}
                    >
                      微信扫码登录
                    </Button>
                  )}
                </Box>
              </>
            )}
          </Box>
        </Paper>
      </Container>
    </Box>
  );
};
