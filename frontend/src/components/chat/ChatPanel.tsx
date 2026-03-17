import { useState, useRef, useEffect } from 'react';
import {
  Box,
  IconButton,
  TextField,
  Typography,
} from '@mui/material';
import { Chat, Send, Close, EmojiEmotions } from '@mui/icons-material';
import { motion, AnimatePresence } from 'framer-motion';
import { useUIStore } from '@/stores/uiStore';
import { useChatStore } from '@/stores/chatStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';

// 内置快捷消息
const QUICK_MESSAGES = [
  { emoji: '👍', text: '厉害' },
  { emoji: '😂', text: '哈哈哈' },
  { emoji: '🎉', text: '赢了' },
  { emoji: '😭', text: '太难了' },
  { emoji: '🤔', text: '让我想想' },
  { emoji: '⏰', text: '快点啊' },
  { emoji: '🍀', text: '好运' },
  { emoji: '💪', text: '加油' },
  { emoji: '😎', text: '稳了' },
  { emoji: '🙏', text: '求放过' },
  { emoji: '😡', text: '气死了' },
  { emoji: '👋', text: '大家好' },
];

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
  const [emojiOpen, setEmojiOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const chatRef = useRef<HTMLDivElement>(null);
  const emojiRef = useRef<HTMLDivElement>(null);

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
        setEmojiOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [chatOpen, toggleChat]);

  // 点击表情面板外关闭
  useEffect(() => {
    if (!emojiOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (emojiRef.current && !emojiRef.current.contains(e.target as Node)) {
        setEmojiOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [emojiOpen]);

  if (roomOnly && !currentRoomId) {
    return null;
  }

  const handleSend = () => {
    if (!message.trim() || !currentRoomId) return;
    sendChat(currentRoomId, message.trim());
    setMessage('');
  };

  const handleQuickSend = (emoji: string, text: string) => {
    if (!currentRoomId) return;
    sendChat(currentRoomId, `${emoji} ${text}`);
    setEmojiOpen(false);
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
                      {/* 快捷消息：大号 emoji 显示 */}
                      {QUICK_MESSAGES.some((qm) => msg.content === `${qm.emoji} ${qm.text}`) ? (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          <Typography sx={{ fontSize: '1.5rem', lineHeight: 1 }}>
                            {msg.content.split(' ')[0]}
                          </Typography>
                          <Typography variant="body2" sx={{ fontSize: '0.85rem', lineHeight: 1.3 }}>
                            {msg.content.split(' ').slice(1).join(' ')}
                          </Typography>
                        </Box>
                      ) : (
                        <Typography variant="body2" sx={{ fontSize: '0.85rem', lineHeight: 1.3 }}>
                          {msg.content}
                        </Typography>
                      )}
                    </Box>
                  </Box>
                );
              })}
              <div ref={messagesEndRef} />
            </Box>

            {/* 快捷表情面板 */}
            <AnimatePresence>
              {emojiOpen && (
                <Box
                  ref={emojiRef}
                  component={motion.div}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 10 }}
                  transition={{ duration: 0.15 }}
                  sx={{
                    px: 2,
                    py: 1.5,
                    bgcolor: 'rgba(0,0,0,0.75)',
                    backdropFilter: 'blur(8px)',
                    pointerEvents: 'auto',
                    display: 'grid',
                    gridTemplateColumns: 'repeat(4, 1fr)',
                    gap: 0.5,
                  }}
                >
                  {QUICK_MESSAGES.map((qm) => (
                    <Box
                      key={qm.text}
                      component={motion.div}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.95 }}
                      onClick={() => handleQuickSend(qm.emoji, qm.text)}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 0.5,
                        px: 1,
                        py: 0.5,
                        borderRadius: 1.5,
                        cursor: 'pointer',
                        bgcolor: 'rgba(255,255,255,0.08)',
                        '&:hover': { bgcolor: 'rgba(255,255,255,0.15)' },
                        transition: 'background 0.15s',
                      }}
                    >
                      <Typography sx={{ fontSize: '1.1rem', lineHeight: 1 }}>{qm.emoji}</Typography>
                      <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.85)', fontSize: '0.7rem', whiteSpace: 'nowrap' }}>
                        {qm.text}
                      </Typography>
                    </Box>
                  ))}
                </Box>
              )}
            </AnimatePresence>

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
              <IconButton
                size="small"
                onClick={() => setEmojiOpen(!emojiOpen)}
                sx={{ color: emojiOpen ? 'primary.main' : 'rgba(255,255,255,0.6)' }}
              >
                <EmojiEmotions fontSize="small" />
              </IconButton>
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
