import React, { useMemo } from 'react';
import { Box, Typography, Button } from '@mui/material';
import { motion } from 'framer-motion';
import { useGameStore } from '@/stores/gameStore';
import { useRoomStore } from '@/stores/roomStore';
import { useUserStore } from '@/stores/userStore';
import { useWebSocket } from '@/hooks/useWebSocket';
import { DurianCard } from './DurianCard';
import { SettlementView } from './SettlementView';
import { PlayerSeat } from './PlayerSeat';
import { AngerToken } from './AngerToken';
import { TurnCountdown } from '@/components/common/TurnCountdown';
import type { DurianServerState, CardData } from './types';

// 环绕座位: 自己在底部, 其他人沿上方弧线 (2~7人)
const getSeatPositions = (count: number) => {
  const positions: Array<{ top: string; left: string; transform: string }> = [];
  positions.push({ top: '88%', left: '50%', transform: 'translate(-50%, -50%)' });
  const others = count - 1;
  for (let i = 0; i < others; i++) {
    const angle = Math.PI + (Math.PI * (i + 1)) / (others + 1);
    const rx = 42;
    const ry = 38;
    const left = 50 + Math.cos(angle) * rx;
    const top = 48 + Math.sin(angle) * ry;
    positions.push({ top: `${top}%`, left: `${left}%`, transform: 'translate(-50%, -50%)' });
  }
  return positions;
};

export const DurianGame: React.FC = () => {
  const { gameData } = useGameStore();
  const { currentRoomId, players: roomPlayers } = useRoomStore();
  const { playerId } = useUserStore();
  const { sendGameAction } = useWebSocket();

  const state = gameData as DurianServerState | null;
  const nicknameMap = new Map<string, string>();
  roomPlayers.forEach((p) => nicknameMap.set(p.player_id, p.nickname));

  if (!state || !playerId) {
    return (
      <Box sx={{ p: 4, textAlign: 'center', bgcolor: '#1a1a1a', borderRadius: 2 }}>
        <Typography sx={{ color: '#888' }}>等待游戏数据...</Typography>
      </Box>
    );
  }

  const isMyTurn = state.current === playerId;
  const hasDrawnCard = state.current_drawn_card !== null;
  const isWaitingContinue = state.phase === 'waiting_continue';
  const needMyContinue = isWaitingContinue && state.waiting_for_continue?.includes(playerId);
  const isSwapOrder = state.waiting_for_swap_order;
  const hasHistory = state.played_cards_history.length > 0;
  const getName = (pid: string) => nicknameMap.get(pid) || pid;

  const handleTakeOrder = () => { if (currentRoomId) sendGameAction(currentRoomId, 'take_order', {}); };
  const handleChooseSide = (side: 'left' | 'right') => {
    if (currentRoomId) sendGameAction(currentRoomId, 'take_order', { card: { chosen_side: side } });
  };
  const handleRingBell = () => { if (currentRoomId) sendGameAction(currentRoomId, 'ring_bell', {}); };
  const handleContinue = () => { if (currentRoomId) sendGameAction(currentRoomId, 'continue', {}); };
  const handleSwapOrder = (idx: number) => {
    if (currentRoomId) sendGameAction(currentRoomId, 'take_order', { card: { order_index: idx } });
  };

  // 排列: 自己在第一位, 用 useMemo 锁定顺序避免结算时位置跳动
  const myIdx = state.players.findIndex((p) => p.player_id === playerId);
  const currentIds = state.players.map((p) => p.player_id);
  const currentIdsKey = currentIds.join(',');

  // 只在玩家列表真正变化时更新顺序
  const stableOrder = useMemo(() => {
    return myIdx >= 0
      ? [...currentIds.slice(myIdx), ...currentIds.slice(0, myIdx)]
      : currentIds;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIdsKey, playerId]);

  // 用稳定顺序映射到最新的 player 数据
  const playerMap = new Map(state.players.map((p) => [p.player_id, p]));
  const orderedPlayers = stableOrder.map((id) => playerMap.get(id)).filter(Boolean) as typeof state.players;

  const otherCardsMap = new Map<string, CardData>();
  state.others_cards?.forEach((oc) => otherCardsMap.set(oc.player_id, oc.card));
  const seatPositions = useMemo(() => getSeatPositions(orderedPlayers.length), [stableOrder.length]);
  const compact = orderedPlayers.length >= 5;


  // Game over
  if (state.game_over) {
    const winners = state.winner.split(',');
    return (
      <Box sx={{ p: 4, textAlign: 'center', bgcolor: '#1a1a1a', borderRadius: 3, border: '2px solid #FFD54F' }}>
        <Typography variant="h5" sx={{ fontWeight: 'bold', color: '#FFD54F', mb: 2 }}>游戏结束</Typography>
        <Typography variant="h6" sx={{ color: '#4CAF50', mb: 2 }}>
          🏆 获胜者: {winners.map(getName).join(', ')}
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center', flexWrap: 'wrap' }}>
          {state.players.map((p) => (
            <Box key={p.player_id} sx={{ textAlign: 'center' }}>
              <Typography variant="caption" sx={{
                color: winners.includes(p.player_id) ? '#4CAF50' : '#888',
                fontWeight: winners.includes(p.player_id) ? 'bold' : 'normal',
              }}>{getName(p.player_id)}</Typography>
              {p.anger_tokens.map((v, i) => <AngerToken key={i} value={v} size={36} />)}
              <Typography variant="caption" sx={{ color: '#E53935', display: 'block' }}>{p.anger_score}分</Typography>
            </Box>
          ))}
        </Box>
      </Box>
    );
  }

  return (
    <Box component={motion.div} initial={{ opacity: 0 }} animate={{ opacity: 1 }} sx={{ width: '100%' }}>
      {/* 桌游桌面 */}
      <Box sx={{
        position: 'relative', width: '100%', minHeight: 560,
        bgcolor: '#2d4a2d', borderRadius: 4, border: '4px solid #5D4037',
        overflow: 'hidden', boxShadow: 'inset 0 0 60px rgba(0,0,0,0.4)',
      }}>
        {/* 桌面纹理 */}
        <Box sx={{
          position: 'absolute', inset: 0,
          background: 'radial-gradient(ellipse at center, rgba(60,90,60,0.3) 0%, transparent 70%)',
          pointerEvents: 'none',
        }} />

        {/* 顶部信息 */}
        <Box sx={{
          position: 'absolute', top: 16, left: 12, right: 12,
          display: 'flex', justifyContent: 'space-between', alignItems: 'center', zIndex: 10,
        }}>
          <Typography variant="caption" sx={{ color: '#aaa' }}>
            第 {state.round_number} 轮 · 牌堆 {state.deck_remaining}
          </Typography>
          <Box sx={{ display: 'flex', gap: 0.3 }}>
            {state.anger_token_pool.map((v, i) => <AngerToken key={i} value={v} size={28} />)}
          </Box>
        </Box>

        {/* 环绕座位 */}
        {orderedPlayers.map((p, idx) => (
          <Box key={p.player_id} sx={{
            position: 'absolute', ...seatPositions[idx],
            zIndex: p.player_id === playerId ? 5 : 3,
          }}>
            <PlayerSeat
              player={p} nickname={getName(p.player_id)}
              holderCard={p.player_id === playerId ? undefined : otherCardsMap.get(p.player_id)}
              isCurrentTurn={p.player_id === state.current}
              isSelf={p.player_id === playerId} compact={compact}
            />
          </Box>
        ))}


        {/* 中央区域: 牌堆 + 订单 + 翻牌 */}
        <Box sx={{
          position: 'absolute', top: '50%', left: '50%',
          transform: 'translate(-50%, -50%)',
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 1,
          zIndex: 6, maxWidth: '65%',
        }}>
          {/* 已打出的牌堆 (多列布局, 超出一列自动换列) */}
          {state.played_cards_history.length > 0 && !state.last_settlement && (
            <Box sx={{
              display: 'flex', flexDirection: 'row', gap: 0.5,
              maxHeight: 260, alignItems: 'flex-start',
            }}>
              {(() => {
                const cardH = 64; // small card height + gap
                const maxPerCol = Math.floor(260 / cardH);
                const cols: typeof state.played_cards_history[] = [];
                for (let i = 0; i < state.played_cards_history.length; i += maxPerCol) {
                  cols.push(state.played_cards_history.slice(i, i + maxPerCol));
                }
                return cols.map((col, ci) => (
                  <Box key={ci} sx={{ display: 'flex', flexDirection: 'column', gap: 0.3 }}>
                    {col.map((record, idx) => (
                      <Box key={ci * maxPerCol + idx} sx={{ position: 'relative' }}>
                        <DurianCard
                          card={record.card}
                          size="small"
                          chosenSide={record.card.card_type === 'fruit' ? 'right' : null}
                          flipped={record.card.card_type === 'fruit' && record.chosen_side === 'left'}
                        />
                        {record.swapped && (
                          <Box sx={{
                            position: 'absolute', top: -4, right: -4,
                            bgcolor: '#FF6B35', borderRadius: '50%',
                            width: 16, height: 16, display: 'flex',
                            alignItems: 'center', justifyContent: 'center',
                            fontSize: 9, lineHeight: 1, boxShadow: '0 1px 3px rgba(0,0,0,0.5)',
                          }}>🔄</Box>
                        )}
                      </Box>
                    ))}
                  </Box>
                ));
              })()}
            </Box>
          )}

          {/* 翻出的牌 */}
          {hasDrawnCard && state.current_drawn_card && !isSwapOrder && (
            <Box sx={{ textAlign: 'center' }}>
              <Typography variant="caption" sx={{ color: '#ccc', mb: 0.5, display: 'block' }}>
                翻出的牌：
              </Typography>
              <DurianCard
                card={state.current_drawn_card} size="large"
                clickable={isMyTurn && state.current_drawn_card.card_type === 'fruit'}
                onChooseLeft={() => handleChooseSide('left')}
                onChooseRight={() => handleChooseSide('right')}
              />
              {isMyTurn && state.current_drawn_card.card_type === 'fruit' && (
                <Box sx={{ display: 'flex', gap: 1, justifyContent: 'center', mt: 0.5 }}>
                  <Button size="small" variant="contained"
                    sx={{ bgcolor: '#555', '&:hover': { bgcolor: '#666' }, fontSize: 11 }}
                    onClick={() => handleChooseSide('left')}>✕ 左面</Button>
                  <Button size="small" variant="contained"
                    sx={{ bgcolor: '#26C6B0', '&:hover': { bgcolor: '#2DD4BF' }, fontSize: 11 }}
                    onClick={() => handleChooseSide('right')}>✓ 右面</Button>
                </Box>
              )}
            </Box>
          )}

          {/* 猩猩牌交换订单 */}
          {isSwapOrder && (
            <Box sx={{
              textAlign: 'center', p: 1.5, borderRadius: 2,
              bgcolor: 'rgba(30,30,30,0.85)', border: '1px solid #FFD54F',
            }}>
              <Typography variant="caption" sx={{ color: '#FFD54F', mb: 1, display: 'block', fontWeight: 'bold' }}>
                🦍 翻到猩猩牌！选择一张牌翻转左右水果
              </Typography>
              {state.current_drawn_card && (
                <Box sx={{ mb: 1 }}>
                  <DurianCard card={state.current_drawn_card} size="small" />
                </Box>
              )}
              <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', justifyContent: 'center' }}>
                {state.played_cards_history.map((record, idx) => {
                  if (record.card.card_type !== 'fruit') return null;
                  const swapped = record.swapped;
                  return (
                    <Box key={idx} sx={{
                      p: 0.5, borderRadius: 1, cursor: isMyTurn && !swapped ? 'pointer' : 'default',
                      opacity: swapped ? 0.4 : 1,
                      border: swapped ? '1px solid #555' : '1px solid #666',
                      '&:hover': !swapped && isMyTurn ? { border: '1px solid #FFD54F' } : {},
                    }}
                      onClick={() => isMyTurn && !swapped && handleSwapOrder(idx)}
                    >
                      <DurianCard
                        card={record.card} size="small"
                        chosenSide="right"
                        flipped={record.chosen_side === 'left'}
                      />
                      <Typography variant="caption" sx={{
                        display: 'block', textAlign: 'center', mt: 0.3,
                        color: swapped ? '#666' : '#ccc', fontSize: 9,
                      }}>
                        {getName(record.player_name)}
                        {swapped ? ' (已翻转)' : ''}
                      </Typography>
                    </Box>
                  );
                })}
              </Box>
            </Box>
          )}

          {/* 结算结果 */}
          {state.last_settlement && (state.phase === 'waiting_continue' || state.phase === 'round_end') && (
            <SettlementView settlement={state.last_settlement} getName={getName} />
          )}
        </Box>


        {/* 底部操作栏 */}
        <Box sx={{
          position: 'absolute', bottom: 8, left: '50%', transform: 'translateX(-50%)',
          display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 0.5, zIndex: 10,
        }}>
          {state.phase === 'playing' && isMyTurn && !hasDrawnCard && !isSwapOrder && (
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button variant="contained" size="small"
                sx={{ bgcolor: '#FFD54F', color: '#333', fontWeight: 'bold', '&:hover': { bgcolor: '#FFE082' } }}
                onClick={handleTakeOrder}>🃏 接新订单</Button>
              {hasHistory && (
                <Button variant="contained" size="small"
                  sx={{ bgcolor: '#E53935', '&:hover': { bgcolor: '#EF5350' } }}
                  onClick={handleRingBell}>🔔 摇铃结算</Button>
              )}
            </Box>
          )}

          {needMyContinue && (
            <Button variant="contained" size="small"
              sx={{ bgcolor: '#FFD54F', color: '#333', '&:hover': { bgcolor: '#FFE082' } }}
              onClick={handleContinue}>确认继续</Button>
          )}

          {isWaitingContinue && !needMyContinue && (
            <Typography variant="caption" sx={{ color: '#888' }}>等待其他玩家确认继续...</Typography>
          )}

          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            {state.turnDeadline && state.phase === 'playing' && (
              <TurnCountdown deadline={state.turnDeadline} />
            )}
            <Typography variant="caption" sx={{
              color: isMyTurn ? '#FFD54F' : '#888',
              fontWeight: isMyTurn ? 'bold' : 'normal',
            }}>
              {isWaitingContinue ? '结算完成'
                : isMyTurn
                  ? hasDrawnCard ? '点击牌面选择左/右' : '轮到你行动'
                  : `等待 ${getName(state.current)}...`}
            </Typography>
          </Box>
        </Box>
      </Box>
    </Box>
  );
};
