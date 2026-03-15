import { Box, Typography, Avatar, Chip } from '@mui/material';
import { motion } from 'framer-motion';
import { MahjongTileComp } from './MahjongTile';
import type { MahjongPlayerInfo, MahjongTile, MeldData } from './types';

interface MahjongBoardProps {
  players: MahjongPlayerInfo[];
  mySeat: number;
  currentSeat: number;
  lastDiscard?: MahjongTile;
  lastDiscardSeat?: number;
  wallRemaining: number;
  nicknameMap: Map<string, string>;
  phase: string;
}

const turnPulse = {
  active: {
    scale: [1, 1.15, 1],
    transition: { duration: 1.2, repeat: Infinity },
  },
};

const windNames = ['东', '南', '西', '北'];

const MeldDisplay: React.FC<{ meld: MeldData }> = ({ meld }) => (
  <Box sx={{ display: 'flex', gap: 0.2 }}>
    {meld.tiles.map((t, i) => (
      <MahjongTileComp
        key={i}
        tile={t}
        size="small"
        faceDown={meld.type === '暗杠' && i > 0 && i < 3}
      />
    ))}
  </Box>
);

const PlayerSlot: React.FC<{
  player: MahjongPlayerInfo;
  nickname: string;
  isCurrentTurn: boolean;
  position: 'top' | 'left' | 'right';
}> = ({ player, nickname, isCurrentTurn, position }) => {
  const isHorizontal = position === 'top';

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 0.5,
        minWidth: isHorizontal ? undefined : 80,
      }}
    >
      <Box sx={{ position: 'relative' }}>
        <Avatar
          sx={{
            width: 40,
            height: 40,
            bgcolor: player.is_dealer ? 'error.main' : 'primary.main',
            border: isCurrentTurn ? '3px solid #4caf50' : 'none',
            fontSize: '0.85rem',
          }}
        >
          {nickname.charAt(0)}
        </Avatar>
        {isCurrentTurn && (
          <Box
            component={motion.div}
            variants={turnPulse}
            animate="active"
            sx={{
              position: 'absolute', top: -3, right: -3,
              width: 12, height: 12, borderRadius: '50%', bgcolor: 'success.main',
            }}
          />
        )}
      </Box>
      <Typography variant="caption" fontWeight="bold" noWrap sx={{ maxWidth: 72 }}>
        {nickname}
      </Typography>
      <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
        <Chip
          label={`${windNames[player.seat]}${player.is_dealer ? '/庄' : ''}`}
          size="small"
          color={player.is_dealer ? 'error' : 'default'}
          sx={{ height: 18, fontSize: '0.6rem' }}
        />
        <Typography variant="caption" color="text.secondary">
          {player.hand_count}张
        </Typography>
      </Box>
      {/* Melds */}
      {player.melds && player.melds.length > 0 && (
        <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', justifyContent: 'center' }}>
          {player.melds.map((m, i) => <MeldDisplay key={i} meld={m} />)}
        </Box>
      )}
      {/* Discards */}
      {player.discards && player.discards.length > 0 && (
        <Box sx={{ display: 'flex', gap: 0.2, flexWrap: 'wrap', justifyContent: 'center', maxWidth: 160 }}>
          {player.discards.map((t, i) => (
            <MahjongTileComp key={i} tile={t} size="small" />
          ))}
        </Box>
      )}
    </Box>
  );
};


export const MahjongBoard: React.FC<MahjongBoardProps> = ({
  players,
  mySeat,
  currentSeat,
  lastDiscard,
  lastDiscardSeat,
  wallRemaining,
  nicknameMap,
  phase,
}) => {
  // Arrange players relative to me: opposite=top, left, right
  const relativeSeats = [
    (mySeat + 2) % 4, // top (opposite)
    (mySeat + 3) % 4, // left
    (mySeat + 1) % 4, // right
  ];

  const getPlayer = (seat: number) => players.find((p) => p.seat === seat);
  const getName = (p: MahjongPlayerInfo) => nicknameMap.get(p.player_id) || p.player_id;

  const topPlayer = getPlayer(relativeSeats[0]);
  const leftPlayer = getPlayer(relativeSeats[1]);
  const rightPlayer = getPlayer(relativeSeats[2]);

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
      {/* Top player */}
      {topPlayer && (
        <Box sx={{ display: 'flex', justifyContent: 'center' }}>
          <PlayerSlot
            player={topPlayer}
            nickname={getName(topPlayer)}
            isCurrentTurn={topPlayer.seat === currentSeat}
            position="top"
          />
        </Box>
      )}

      {/* Middle row: left - center - right */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, minHeight: 120 }}>
        {leftPlayer && (
          <PlayerSlot
            player={leftPlayer}
            nickname={getName(leftPlayer)}
            isCurrentTurn={leftPlayer.seat === currentSeat}
            position="left"
          />
        )}

        {/* Center area */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: 80,
            bgcolor: 'action.hover',
            borderRadius: 2,
            p: 1.5,
            gap: 0.5,
          }}
        >
          <Typography variant="caption" color="text.secondary">
            牌墙: {wallRemaining} 张
          </Typography>
          {lastDiscard != null && lastDiscardSeat != null ? (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Typography variant="caption" color="text.secondary">
                {nicknameMap.get(players.find(p => p.seat === lastDiscardSeat)?.player_id || '') || `座位${lastDiscardSeat}`} 打出:
              </Typography>
              <MahjongTileComp tile={lastDiscard} highlight />
            </Box>
          ) : (
            <Typography variant="caption" color="text.secondary">
              {phase === 'pending' ? '等待响应...' : '等待出牌...'}
            </Typography>
          )}
        </Box>

        {rightPlayer && (
          <PlayerSlot
            player={rightPlayer}
            nickname={getName(rightPlayer)}
            isCurrentTurn={rightPlayer.seat === currentSeat}
            position="right"
          />
        )}
      </Box>
    </Box>
  );
};
