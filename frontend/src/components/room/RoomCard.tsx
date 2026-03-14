import { Card, CardContent, CardActions, Typography, Button, Box, Chip } from '@mui/material';
import { motion } from 'framer-motion';
import { People, SportsEsports } from '@mui/icons-material';
import type { RoomInfo } from '@/types/room';

interface RoomCardProps {
  room: RoomInfo;
  onJoin: (roomId: string) => void;
}

const gameTypeNames: Record<string, string> = {
  ddz: '斗地主',
  mahjong: '麻将',
  durian: '榴莲忘返',
};

export const RoomCard: React.FC<RoomCardProps> = ({ room, onJoin }) => {
  const isFull = room.players.length >= room.max_players;
  const isPlaying = room.status === 'playing';
  const canJoin = !isFull && !isPlaying;

  return (
    <Card
      component={motion.div}
      whileHover={canJoin ? { scale: 1.02, y: -4 } : {}}
      transition={{ duration: 0.2 }}
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        opacity: canJoin ? 1 : 0.7,
      }}
    >
      <CardContent sx={{ flexGrow: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" component="div" fontWeight="bold">
            {gameTypeNames[room.game_type] || room.game_type}
          </Typography>
          <Chip
            label={room.status === 'waiting' ? '等待中' : '游戏中'}
            color={room.status === 'waiting' ? 'success' : 'warning'}
            size="small"
          />
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <SportsEsports fontSize="small" color="action" />
          <Typography variant="body2" color="text.secondary">
            房间号: {room.room_id}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <People fontSize="small" color="action" />
          <Typography variant="body2" color="text.secondary">
            {room.players.length}/{room.max_players} 人
          </Typography>
        </Box>
      </CardContent>

      <CardActions>
        <Button
          fullWidth
          variant={canJoin ? 'contained' : 'outlined'}
          disabled={!canJoin}
          onClick={() => onJoin(room.room_id)}
        >
          {isFull ? '房间已满' : isPlaying ? '游戏中' : '加入房间'}
        </Button>
      </CardActions>
    </Card>
  );
};
