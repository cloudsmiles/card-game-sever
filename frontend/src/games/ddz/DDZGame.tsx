import { useState, useEffect } from 'react';
import { Box, Typography, useMediaQuery, useTheme, Paper, Button } from '@mui/material';
import { ScreenRotation } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { DDZBoard } from './DDZBoard';
import { DDZHand } from './DDZHand';
import { DDZActions } from './DDZActions';
import { DDZBidding } from './DDZBidding';
import { useGameStore } from '@/stores/gameStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { DDZServerState, DDZPlayerInfo } from './types';

export const DDZGame: React.FC = () => {
  const { gameData } = useGameStore();
  const { currentRoomId, players: roomPlayers } = useRoomStore();
  const { playerId } = useUserStore();
  const { sendGameAction } = useWebSocket();
  const [selectedCards, setSelectedCards] = useState<Set<number>>(new Set());
  const [landscapeDismissed, setLandscapeDismissed] = useState(false);

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const isPortrait = useMediaQuery('(orientation: portrait)');

  // Parse server state
  const state = gameData as DDZServerState | null;

  // Build nickname map from room players
  const nicknameMap = new Map<string, string>();
  roomPlayers.forEach((p) => {
    nicknameMap.set(p.player_id, p.nickname);
  });

  // Clear selection when turn changes
  useEffect(() => {
    setSelectedCards(new Set());
  }, [state?.current]);

  if (!state || !playerId) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography color="text.secondary">等待游戏数据...</Typography>
      </Paper>
    );
  }

  // Determine other players relative to me
  const playerInfos = (state.player_infos || []).map((p) => ({
    ...p,
    nickname: nicknameMap.get(p.player_id) || p.player_id,
  }));
  const otherPlayers = playerInfos.filter((p) => p && p.player_id !== playerId);
  const leftPlayer: DDZPlayerInfo | null = otherPlayers[0] || null;
  const rightPlayer: DDZPlayerInfo | null = otherPlayers[1] || null;

  const myHand = state.my_hand || [];
  const isMyTurn = state.current === playerId;
  const phase = state.phase;

  const handleCardClick = (index: number) => {
    setSelectedCards((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  const handleBid = (score: number) => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'call_landlord', { card: score });
  };

  const handlePlay = () => {
    if (!currentRoomId || selectedCards.size === 0) return;
    const cardsToPlay = Array.from(selectedCards).map((idx) => ({ value: myHand[idx] }));
    sendGameAction(currentRoomId, 'play_cards', { card: cardsToPlay });
    setSelectedCards(new Set());
  };

  const handlePass = () => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'pass');
    setSelectedCards(new Set());
  };

  const handleClearSelection = () => {
    setSelectedCards(new Set());
  };

  const canPass = isMyTurn && state.last_play && state.last_play.Type !== '';

  // Mobile portrait landscape prompt
  if (isMobile && isPortrait && !landscapeDismissed) {
    return (
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '60vh',
          gap: 3,
          p: 3,
        }}
      >
        <ScreenRotation sx={{ fontSize: 64, color: 'primary.main' }} />
        <Typography variant="h6" textAlign="center">
          为了更好的游戏体验，请横屏使用
        </Typography>
        <Button variant="outlined" onClick={() => setLandscapeDismissed(true)}>
          继续竖屏
        </Button>
      </Box>
    );
  }

  return (
    <Box
      component={motion.div}
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      sx={{ width: '100%' }}
    >
      <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 1 }}>
        {/* Game Board: bottom cards at top, left/right players */}
        <DDZBoard
          leftPlayer={leftPlayer}
          rightPlayer={rightPlayer}
          bottomCards={state.bottom}
          lastPlay={state.last_play}
          currentTurn={state.current}
          phase={phase}
          myPlayerId={playerId}
        />

        {/* My Hand */}
        <Box>
          <Typography variant="body2" sx={{ textAlign: 'center', mb: 1 }}>
            我的手牌 ({myHand.length} 张)
            {state.landlord === playerId && (
              <Typography component="span" color="error.main" fontWeight="bold" sx={{ ml: 1 }}>
                [地主]
              </Typography>
            )}
          </Typography>

          <DDZHand
            cards={myHand}
            selectedCards={selectedCards}
            onCardClick={handleCardClick}
            disabled={!isMyTurn || phase === 'call'}
          />
        </Box>

        {/* Actions */}
        {phase === 'call' ? (
          <DDZBidding isMyTurn={isMyTurn} onBid={handleBid} />
        ) : (
          <DDZActions
            onPlay={handlePlay}
            onPass={handlePass}
            onClearSelection={handleClearSelection}
            canPlay={isMyTurn && selectedCards.size > 0}
            canPass={!!canPass}
            hasSelection={selectedCards.size > 0}
            disabled={!isMyTurn}
          />
        )}

        {/* Turn indicator */}
        <Typography
          variant="body2"
          sx={{ textAlign: 'center', color: isMyTurn ? 'success.main' : 'text.secondary', fontWeight: isMyTurn ? 'bold' : 'normal' }}
        >
          {isMyTurn
            ? phase === 'call' ? '轮到你叫地主' : '轮到你出牌'
            : '等待其他玩家...'}
        </Typography>
      </Paper>
    </Box>
  );
};
