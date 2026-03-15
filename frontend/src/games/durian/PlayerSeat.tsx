import React from 'react';
import { Box, Typography } from '@mui/material';
import { DurianCard } from './DurianCard';
import { AngerToken } from './AngerToken';
import type { CardData, PlayerInfo } from './types';

interface PlayerSeatProps {
  player: PlayerInfo;
  nickname: string;
  holderCard?: CardData;
  isCurrentTurn: boolean;
  isSelf: boolean;
  compact?: boolean;
}

export const PlayerSeat: React.FC<PlayerSeatProps> = ({
  player, nickname, holderCard, isCurrentTurn, isSelf, compact = false,
}) => {
  const tokenSize = compact ? 32 : 40;

  return (
    <Box sx={{
      display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 0.5,
      p: compact ? 0.5 : 1,
      borderRadius: 2,
      border: isCurrentTurn ? '2px solid #FFD54F' : '2px solid transparent',
      bgcolor: isSelf ? 'rgba(255,107,53,0.1)' : 'transparent',
      transition: 'all 0.3s',
      minWidth: compact ? 80 : 100,
    }}>
      {/* 昵称 */}
      <Typography
        variant="caption"
        sx={{
          fontWeight: isCurrentTurn ? 'bold' : 'normal',
          color: isCurrentTurn ? '#FFD54F' : isSelf ? '#FF6B35' : '#ccc',
          fontSize: compact ? 10 : 12,
          maxWidth: compact ? 70 : 90,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}
      >
        {nickname}{isSelf ? '(你)' : ''}{isCurrentTurn ? ' 👈' : ''}
      </Typography>

      {/* 牌架卡 (其他人可见) */}
      {holderCard && (
        <DurianCard card={holderCard} size={compact ? 'small' : 'small'} />
      )}

      {/* 自己的牌架卡 - 背面 */}
      {isSelf && !holderCard && (
        <Box sx={{
          width: compact ? 70 : 90, height: compact ? 46 : 60,
          borderRadius: 1, bgcolor: '#2a2a2a',
          border: '1px dashed #555',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <Typography sx={{ fontSize: 10, color: '#666' }}>?</Typography>
        </Box>
      )}

      {/* 愤怒标记 */}
      {player.anger_tokens.length > 0 && (
        <Box sx={{ display: 'flex', gap: 0.3, flexWrap: 'wrap', justifyContent: 'center' }}>
          {player.anger_tokens.map((v, i) => (
            <AngerToken key={i} value={v} size={tokenSize} />
          ))}
        </Box>
      )}

      {/* 总分 */}
      {player.anger_score > 0 && (
        <Typography variant="caption" sx={{ color: '#E53935', fontWeight: 'bold', fontSize: compact ? 10 : 11 }}>
          {player.anger_score}分
        </Typography>
      )}
    </Box>
  );
};
