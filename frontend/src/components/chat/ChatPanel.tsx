import { useState, useRef, useEffect } from 'react';
import {
  Box,
  Paper,
  IconButton,
  TextField,
  Button,
  Typography,
  Drawer,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { Chat, Close, Send, EmojiEmotions } from '@mui/icons-material';
import { motion, AnimatePresence } from 'framer-motion';
import { useUIStore } from '@/stores/uiStore';
import { useChatStore } from '@/stores/chatStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';

interface ChatPanelProps {
  roomOnly?: boolean; // 是否只在房间内显示
}

const QUICK_PHRASES = [
  '你好！',
  '快点啊',
  '等等',
  '好的',
  '谢谢',
  '再见',
];

export const ChatPanel: React.FC<ChatPanelProps> = ({ roomOnly = false }) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { chatOpen, toggleChat } = useUIStore();
  const { messages } = useChatStore();
  const { currentRoomId } = useRoomStore();
  const { playerId } = useUserStore();
  const { sendChat } = useWebSocket();

  const [message, setMessage] = useState('');
  const [showQuickPhrases, setShowQuickPhrases] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 如果设置为只在房间内显示，且当前不在房间内，则不显示
  if (roomOnly && !currentRoomId) {
    return null;
  }

  const handleSend = () => {
    if (!message.trim() || !currentRoomId) return;

    sendChat(currentRoomId, message.trim());
    setMessage('');
  };

  const handleQuickPhrase = (phrase: string) => {
    if (!currentRoomId) return;
    sendChat(currentRoomId, phrase);
    setShowQuickPhrases(false);
  };

  const chatContent = (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
      }}
    >
      <Box
        sx={{
          p: 2,
          borderBottom: 1,
          borderColor: 'divider',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Typography variant="h6" fontWeight="bold">
          聊天
        </Typography>
        <IconButton size="small" onClick={toggleChat}>
          <Close />
        </IconButton>
      </Box>

      <Box
        sx={{
          flexGrow: 1,
          overflowY: 'auto',
          p: 2,
          display: 'flex',
          flexDirection: 'column',
          gap: 1,
        }}
      >
        {messages.length === 0 ? (
          <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', mt: 4 }}>
            暂无消息
          </Typography>
        ) : (
          messages.map((msg) => {
            const isOwn = msg.playerId === playerId;
            const isSystem = msg.playerId === 'system';

            if (isSystem) {
              return (
                <Box
                  key={msg.id}
                  component={motion.div}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  sx={{ display: 'flex', justifyContent: 'center', py: 0.5 }}
                >
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ bgcolor: 'action.hover', px: 1.5, py: 0.5, borderRadius: 2, fontSize: '0.7rem' }}
                  >
                    {msg.content}
                  </Typography>
                </Box>
              );
            }

            return (
              <Box
                key={msg.id}
                component={motion.div}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: isOwn ? 'flex-end' : 'flex-start',
                }}
              >
                <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5 }}>
                  {isOwn ? '我' : msg.nickname}
                </Typography>
                <Paper
                  sx={{
                    p: 1.5,
                    maxWidth: '70%',
                    bgcolor: isOwn ? 'primary.main' : 'background.default',
                    color: isOwn ? 'primary.contrastText' : 'text.primary',
                  }}
                >
                  <Typography variant="body2">{msg.content}</Typography>
                </Paper>
              </Box>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </Box>

      <Box sx={{ p: 2, borderTop: 1, borderColor: 'divider' }}>
        <AnimatePresence>
          {showQuickPhrases && (
            <Box
              component={motion.div}
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              sx={{ mb: 2, display: 'flex', flexWrap: 'wrap', gap: 1 }}
            >
              {QUICK_PHRASES.map((phrase) => (
                <Button
                  key={phrase}
                  size="small"
                  variant="outlined"
                  onClick={() => handleQuickPhrase(phrase)}
                >
                  {phrase}
                </Button>
              ))}
            </Box>
          )}
        </AnimatePresence>

        <Box sx={{ display: 'flex', gap: 1 }}>
          <IconButton
            size="small"
            onClick={() => setShowQuickPhrases(!showQuickPhrases)}
            color={showQuickPhrases ? 'primary' : 'default'}
          >
            <EmojiEmotions />
          </IconButton>

          <TextField
            fullWidth
            size="small"
            placeholder="输入消息..."
            value={message}
            onChange={(e) => setMessage(e.target.value.slice(0, 200))}
            onKeyPress={(e) => e.key === 'Enter' && handleSend()}
            disabled={!currentRoomId}
          />

          <IconButton color="primary" onClick={handleSend} disabled={!message.trim() || !currentRoomId}>
            <Send />
          </IconButton>
        </Box>

        <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
          {message.length}/200
        </Typography>
      </Box>
    </Box>
  );

  if (isMobile) {
    return (
      <>
        <IconButton
          sx={{
            position: 'fixed',
            bottom: 16,
            right: 16,
            bgcolor: 'primary.main',
            color: 'white',
            '&:hover': { bgcolor: 'primary.dark' },
            zIndex: 1000,
          }}
          onClick={toggleChat}
        >
          <Chat />
        </IconButton>

        <Drawer anchor="right" open={chatOpen} onClose={toggleChat}>
          <Box sx={{ width: 320, height: '100%' }}>{chatContent}</Box>
        </Drawer>
      </>
    );
  }

  return (
    <>
      {!chatOpen && (
        <IconButton
          component={motion.button}
          whileHover={{ scale: 1.1 }}
          sx={{
            position: 'fixed',
            top: 80,
            right: 16,
            bgcolor: 'primary.main',
            color: 'white',
            '&:hover': { bgcolor: 'primary.dark' },
            zIndex: 1000,
          }}
          onClick={toggleChat}
        >
          <Chat />
        </IconButton>
      )}

      <AnimatePresence>
        {chatOpen && (
          <Box
            component={motion.div}
            initial={{ x: 320, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
            exit={{ x: 320, opacity: 0 }}
            transition={{ type: 'spring', damping: 25 }}
            sx={{
              position: 'fixed',
              top: 64,
              right: 0,
              width: 320,
              height: 'calc(100vh - 64px)',
              zIndex: 1000,
            }}
          >
            {chatContent}
          </Box>
        )}
      </AnimatePresence>
    </>
  );
};
