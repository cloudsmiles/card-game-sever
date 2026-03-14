import { Box, Typography, Avatar, Chip } from '@mui/material';
import { motion } from 'framer-motion';
import { DDZCard } from './DDZCard';
import type { DDZPlayerInfo, LastPlay } from './types';

interface DDZBoardProps {
  leftPlayer: DDZPlayerInfo | null;
  rightPlayer: DDZPlayerInfo | null;
  bottomCards: number[] | null;
  lastPlay: LastPlay | null;
  currentTurn: string;
  phase: 'call' | 'play';
  myPlayerId: string;
}

const turnPulse = {
  active: {
    scale: [1, 1.15, 1],
    transition: { duration: 1.2, repeat: Infinity },
  },
};

const PlayerSlot: React.FC<{
  player: DDZPlayerInfo | null;
  isCurrentTurn: boolean;
}> = ({ player, isCurrentTurn }) => {
  if (!player) return null;

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 0.5,
        minWidth: 80,
      }}
    >
      <Box sx={{ position: 'relative' }}>
        <Avatar
          sx={{
            width: 48,
            height: 48,
            bgcolor: player.is_landlord ? 'error.main' : 'primary.main',
            border: isCurrentTurn ? '3px solid #4caf50' : 'none',
            fontSize: '1rem',
          }}
        >
          {(player.nickname || player.player_id).charAt(0)}
        </Avatar>
        {isCurrentTurn && (
          <Box
            component={motion.div}
            variants={turnPulse}
            animate="active"
            sx={{
              position: 'absolute',
              top: -4,
              right: -4,
              width: 14,
              height: 14,
              borderRadius: '50%',
              bgcolor: 'success.main',
            }}
          />
        )}
      </Box>

      <Typography variant="caption" fontWeight="bold" noWrap sx={{ maxWidth: 80 }}>
        {player.nickname || player.player_id}
      </Typography>

      {player.is_landlord && (
        <Chip label="地主" size="small" color="error" sx={{ height: 20, fontSize: '0.65rem' }} />
      )}

      <Typography variant="caption" color="text.secondary">
        {player.hand_count} 张
      </Typography>
    </Box>
  );
};

export const DDZBoard: React.FC<DDZBoardProps> = ({
  leftPlayer,
  rightPlayer,
  bottomCards,
  lastPlay,
  currentTurn,
  phase,
}) => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      {/* Top: Bottom cards (底牌) */}
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 1, minHeight: 60 }}>
        <Typography variant="caption" color="text.secondary">
          底牌:
        </Typography>
        {bottomCards && bottomCards.length > 0 ? (
          bottomCards.map((v, i) => <DDZCard key={i} value={v} index={i} size="small" />)
        ) : (
          <>
            <DDZCard value={0} faceDown size="small" />
            <DDZCard value={0} faceDown size="small" />
            <DDZCard value={0} faceDown size="small" />
          </>
        )}
      </Box>

      {/* Middle: Left player - Play area - Right player */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 2,
          minHeight: 140,
        }}
      >
        {/* Left player */}
        <PlayerSlot player={leftPlayer} isCurrentTurn={leftPlayer?.player_id === currentTurn} />

        {/* Center play area */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: 100,
            bgcolor: 'action.hover',
            borderRadius: 2,
            p: 2,
          }}
        >
          {phase === 'call' ? (
            <Typography variant="body2" color="text.secondary">
              叫地主阶段
            </Typography>
          ) : lastPlay && lastPlay.Cards && lastPlay.Cards.length > 0 ? (
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block', textAlign: 'center' }}>
                上次出牌
              </Typography>
              <Box sx={{ display: 'flex', gap: 0.5, justifyContent: 'center', flexWrap: 'wrap' }}>
                {lastPlay.Cards.map((card, i) => (
                  <DDZCard key={i} value={card.Value} index={i} size="small" />
                ))}
              </Box>
            </Box>
          ) : (
            <Typography variant="body2" color="text.secondary">
              等待出牌...
            </Typography>
          )}
        </Box>

        {/* Right player */}
        <PlayerSlot player={rightPlayer} isCurrentTurn={rightPlayer?.player_id === currentTurn} />
      </Box>
    </Box>
  );
};
