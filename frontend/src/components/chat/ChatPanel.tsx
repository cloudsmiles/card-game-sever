import { useState, useRef, useEffect } from 'react';
import {
  Box,
  IconButton,
  TextField,
  Typography,
} from '@mui/material';
import { Chat, Send, Close } from '@mui/icons-material';
import { motion, AnimatePresence } from 'framer-motion';
import { useUIStore } from '@/stores/uiStore';
import { useChatStore } from '@/stores/chatStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';

interface ChatPanelProps {
  roomOnly?: boolean;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ roomOnly = false }) => {
  const { chatOpen, toggleChat } = useUIStore();
  const { messages } = useChatStore();
  const { currentRoomId } = useRoomStore();
  const { playerId } = useUserStore();
  const { sendChat } = useWebSocket();

  const [message, setMessage] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const chatRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // 点击聊天区域外自动隐藏
  useEffect(() => {
    if (!chatOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (chatRef.current && !chatRef.current.contains(e.target as Node)) {
        toggleChat();
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [chatOpen, toggleChat]);

  if (roomOnly && !currentRoomId) {
    return null;
  }

  const handleSend = () => {
    if (!message.trim() || !currentRoomId) return;
    sendChat(currentRoomId, message.trim());
    setMessage('');
  };

  // 只显示最近的消息（浮层模式下不需要太多历史）
  const recentMessages = messages.slice(-20);

  return (
    <>
      {/* 聊天开关按钮 */}
      {!chatOpen && (
        <IconButton
          component={motion.button}
          whileHover={{ scale: 1.1 }}
          sx={{
            position: 'fixed',
            bottom: 16,
            right: 16,
            bgcolor: 'primary.main',
            color: 'white',
            '&:hover': { bgcolor: 'primary.dark' },
            zIndex: 1200,
            width: 48,
            height: 48,
          }}
          onClick={toggleChat}
        >
          <Chat />
        </IconButton>
      )}

      {/* 透明浮层聊天窗口 */}
      <AnimatePresence>
        {chatOpen && (
          <Box
            ref={chatRef}
            component={motion.div}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 20 }}
            transition={{ duration: 0.2 }}
            sx={{
              position: 'fixed',
              bottom: 0,
              right: 0,
              width: { xs: '100%', sm: 360 },
              maxHeight: 400,
              zIndex: 1200,
              display: 'flex',
              flexDirection: 'column',
              pointerEvents: 'none',
            }}
          >
            {/* 消息区域 - 透明背景，消息自底向上 */}
            <Box
              sx={{
                flex: 1,
                overflowY: 'auto',
                px: 2,
                py: 1,
                display: 'flex',
                flexDirection: 'column',
                gap: 0.5,
                pointerEvents: 'auto',
                // 隐藏滚动条
                '&::-webkit-scrollbar': { display: 'none' },
                scrollbarWidth: 'none',
              }}
            >
              {recentMessages.map((msg) => {
                const isOwn = msg.playerId === playerId;
                const isSystem = msg.playerId === 'system';

                if (isSystem) {
                  return (
                    <Box
                      key={msg.id}
                      component={motion.div}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      sx={{ display: 'flex', justifyContent: 'center', py: 0.25 }}
                    >
                      <Typography
                        variant="caption"
                        sx={{
                          color: 'rgba(255,255,255,0.7)',
                          bgcolor: 'rgba(0,0,0,0.3)',
                          px: 1.5,
                          py: 0.25,
                          borderRadius: 2,
                          fontSize: '0.7rem',
                        }}
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
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    sx={{
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: isOwn ? 'flex-end' : 'flex-start',
                    }}
                  >
                    <Box
                      sx={{
                        maxWidth: '75%',
                        bgcolor: isOwn ? 'rgba(255,107,53,0.85)' : 'rgba(0,0,0,0.5)',
                        color: 'white',
                        px: 1.5,
                        py: 0.5,
                        borderRadius: 2,
                        backdropFilter: 'blur(4px)',
                      }}
                    >
                      {!isOwn && (
                        <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.7)', fontSize: '0.65rem' }}>
                          {msg.nickname}
                        </Typography>
                      )}
                      <Typography variant="body2" sx={{ fontSize: '0.85rem', lineHeight: 1.3 }}>
                        {msg.content}
                      </Typography>
                    </Box>
                  </Box>
                );
              })}
              <div ref={messagesEndRef} />
            </Box>

            {/* 底部输入框 */}
            <Box
              sx={{
                px: 2,
                py: 1,
                bgcolor: 'rgba(0,0,0,0.6)',
                backdropFilter: 'blur(8px)',
                display: 'flex',
                alignItems: 'center',
                gap: 1,
                pointerEvents: 'auto',
              }}
            >
              <TextField
                fullWidth
                size="small"
                placeholder="输入消息..."
                value={message}
                onChange={(e) => setMessage(e.target.value.slice(0, 200))}
                onKeyPress={(e) => e.key === 'Enter' && handleSend()}
                disabled={!currentRoomId}
                sx={{
                  '& .MuiOutlinedInput-root': {
                    bgcolor: 'rgba(255,255,255,0.1)',
                    color: 'white',
                    fontSize: '0.85rem',
                    '& fieldset': { borderColor: 'rgba(255,255,255,0.2)' },
                    '&:hover fieldset': { borderColor: 'rgba(255,255,255,0.4)' },
                    '&.Mui-focused fieldset': { borderColor: 'primary.main' },
                  },
                  '& .MuiOutlinedInput-input::placeholder': {
                    color: 'rgba(255,255,255,0.5)',
                  },
                }}
              />
              <IconButton
                size="small"
                onClick={handleSend}
                disabled={!message.trim() || !currentRoomId}
                sx={{ color: 'white' }}
              >
                <Send fontSize="small" />
              </IconButton>
              <IconButton
                size="small"
                onClick={toggleChat}
                sx={{ color: 'rgba(255,255,255,0.6)' }}
              >
                <Close fontSize="small" />
              </IconButton>
            </Box>
          </Box>
        )}
      </AnimatePresence>
    </>
  );
};
