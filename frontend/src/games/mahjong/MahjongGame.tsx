import { useState, useEffect } from 'react';
import { Box, Typography, Paper, Button, useMediaQuery, useTheme } from '@mui/material';
import { ScreenRotation } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { MahjongBoard } from './MahjongBoard';
import { MahjongHand } from './MahjongHand';
import { MahjongActions } from './MahjongActions';
import { useGameStore } from '@/stores/gameStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';
import { TurnCountdown } from '@/components/common/TurnCountdown';
import type { MahjongServerState, MahjongTile, MeldData } from './types';
import { MahjongTileComp } from './MahjongTile';

export const MahjongGame: React.FC = () => {
  const { gameData } = useGameStore();
  const { currentRoomId, players: roomPlayers } = useRoomStore();
  const { playerId } = useUserStore();
  const { sendGameAction } = useWebSocket();
  const [selectedIndex, setSelectedIndex] = useState<Set<number>>(new Set());
  const [landscapeDismissed, setLandscapeDismissed] = useState(false);

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const isPortrait = useMediaQuery('(orientation: portrait)');

  const state = gameData as MahjongServerState | null;

  const nicknameMap = new Map<string, string>();
  roomPlayers.forEach((p) => {
    nicknameMap.set(p.player_id, p.nickname);
  });

  // Clear selection on turn change
  useEffect(() => {
    setSelectedIndex(new Set());
  }, [state?.current_seat, state?.phase]);

  if (!state || !playerId) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography color="text.secondary">等待游戏数据...</Typography>
      </Paper>
    );
  }

  // Game over display
  if (state.game_over) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography variant="h5" fontWeight="bold" sx={{ mb: 2 }}>
          游戏结束
        </Typography>
        <Typography variant="h6" color="primary.main" sx={{ mb: 2 }}>
          {state.winner}
        </Typography>
        {state.win_result && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="body2" sx={{ mb: 1 }}>
              总番数: {state.win_result.total_fan}
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center', flexWrap: 'wrap' }}>
              {state.win_result.fan_list?.map((f, i) => (
                <Typography key={i} variant="caption" sx={{ px: 1, py: 0.5, bgcolor: 'action.hover', borderRadius: 1 }}>
                  {f.name} ({f.fan}番)
                </Typography>
              ))}
            </Box>
          </Box>
        )}
      </Paper>
    );
  }

  const myHand = state.my_hand || [];
  const mySeat = state.my_seat;
  const isMyTurn = state.current === playerId;
  const availableActions = state.available_actions || [];
  const myPlayer = state.players.find((p) => p.seat === mySeat);
  const myMelds = myPlayer?.melds || [];
  const myDiscards = myPlayer?.discards || [];

  const handleTileClick = (index: number) => {
    // In play phase, only allow single selection for discard
    if (state.phase === 'play' && isMyTurn) {
      setSelectedIndex((prev) => {
        const next = new Set<number>();
        if (!prev.has(index)) next.add(index);
        return next;
      });
    }
  };

  const handleDiscard = () => {
    if (!currentRoomId || selectedIndex.size !== 1) return;
    const idx = Array.from(selectedIndex)[0];
    const tile = myHand[idx];
    sendGameAction(currentRoomId, 'discard', { card: tile });
    setSelectedIndex(new Set());
  };

  const handleWin = () => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'win');
  };

  const handleKong = (tile: MahjongTile) => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'kong', { card: tile });
  };

  const handlePong = () => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'pong');
  };

  const handleChow = (tiles: MahjongTile[]) => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'chow', { card: tiles });
  };

  const handlePass = () => {
    if (!currentRoomId) return;
    sendGameAction(currentRoomId, 'pass');
    setSelectedIndex(new Set());
  };

  // Mobile portrait prompt
  if (isMobile && isPortrait && !landscapeDismissed) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '60vh', gap: 3, p: 3 }}>
        <ScreenRotation sx={{ fontSize: 64, color: 'primary.main' }} />
        <Typography variant="h6" textAlign="center">为了更好的游戏体验，请横屏使用</Typography>
        <Button variant="outlined" onClick={() => setLandscapeDismissed(true)}>继续竖屏</Button>
      </Box>
    );
  }

  return (
    <Box component={motion.div} initial={{ opacity: 0 }} animate={{ opacity: 1 }} sx={{ width: '100%' }}>
      <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 1 }}>
        <MahjongBoard
          players={state.players}
          mySeat={mySeat}
          currentSeat={state.current_seat}
          lastDiscard={state.last_discard}
          lastDiscardSeat={state.last_discard_seat}
          wallRemaining={state.wall_remaining}
          nicknameMap={nicknameMap}
          phase={state.phase}
        />

        {/* My hand */}
        <Box>
          <Typography variant="body2" sx={{ textAlign: 'center', mb: 0.5 }}>
            我的手牌 ({myHand.length} 张)
            {state.dealer_seat === mySeat && (
              <Typography component="span" color="error.main" fontWeight="bold" sx={{ ml: 1 }}>
                [庄]
              </Typography>
            )}
          </Typography>
          <MahjongHand
            tiles={myHand}
            selectedIndices={selectedIndex}
            onTileClick={handleTileClick}
            disabled={!isMyTurn && state.phase !== 'pending'}
          />
        </Box>

        {/* My melds */}
        {myMelds.length > 0 && (
          <Box sx={{ display: 'flex', gap: 1.5, justifyContent: 'center', flexWrap: 'wrap' }}>
            {myMelds.map((meld: MeldData, i: number) => (
              <Box key={i} sx={{ display: 'flex', gap: 0.3, alignItems: 'center' }}>
                {meld.tiles.map((t, j) => (
                  <MahjongTileComp
                    key={j}
                    tile={t}
                    size="small"
                    faceDown={meld.type === '暗杠' && j > 0 && j < 3}
                  />
                ))}
                <Typography variant="caption" color="text.secondary" sx={{ ml: 0.5 }}>
                  {meld.type}
                </Typography>
              </Box>
            ))}
          </Box>
        )}

        {/* My discards - 每行6张 */}
        {myDiscards.length > 0 && (
          <Box>
            <Typography variant="caption" color="text.secondary" sx={{ textAlign: 'center', display: 'block', mb: 0.5 }}>
              我的出牌 ({myDiscards.length} 张)
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 0.5 }}>
              {Array.from({ length: Math.ceil(myDiscards.length / 6) }).map((_, row) => (
                <Box key={row} sx={{ display: 'flex', gap: 0.3 }}>
                  {myDiscards.slice(row * 6, row * 6 + 6).map((t, i) => (
                    <MahjongTileComp key={row * 6 + i} tile={t} size="small" />
                  ))}
                </Box>
              ))}
            </Box>
          </Box>
        )}

        {/* Actions */}
        <MahjongActions
          availableActions={availableActions}
          onDiscard={handleDiscard}
          onWin={handleWin}
          onKong={handleKong}
          onPong={handlePong}
          onChow={handleChow}
          onPass={handlePass}
          hasSelection={selectedIndex.size > 0}
          kongOptions={state.kong_options}
          chowOptions={state.chow_options}
        />

        {/* Turn indicator */}
        <Typography
          variant="body2"
          sx={{
            textAlign: 'center',
            color: isMyTurn ? 'success.main' : (state.phase === 'pending' && availableActions.length > 0) ? 'warning.main' : 'text.secondary',
            fontWeight: isMyTurn || availableActions.length > 0 ? 'bold' : 'normal',
          }}
        >
          {isMyTurn
            ? '轮到你出牌'
            : availableActions.length > 0
              ? '你可以响应'
              : '等待其他玩家...'}
        </Typography>
        <Box sx={{ display: 'flex', justifyContent: 'center' }}>
          {state.phase === 'pending' && state.pending_deadline > 0 ? (
            <TurnCountdown deadline={state.pending_deadline} label={availableActions.length > 0 ? '响应倒计时' : ''} />
          ) : (
            <TurnCountdown deadline={state.turn_deadline} label={isMyTurn ? '倒计时' : ''} />
          )}
        </Box>
      </Paper>
    </Box>
  );
};
