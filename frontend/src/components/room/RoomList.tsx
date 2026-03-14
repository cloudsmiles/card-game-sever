import { useEffect, useState } from 'react';
import {
  Box,
  Grid,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Pagination,
} from '@mui/material';
import { Refresh } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { RoomCard } from './RoomCard';
import { Loading } from '../common';
import { roomApi } from '@/services/roomApi';
import { useUIStore } from '@/stores/uiStore';
import type { RoomInfo, GameType } from '@/types/room';

interface RoomListProps {
  onJoinRoom: (roomId: string) => void;
}

export const RoomList: React.FC<RoomListProps> = ({ onJoinRoom }) => {
  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [gameTypeFilter, setGameTypeFilter] = useState<GameType | 'all'>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | 'waiting' | 'playing'>('all');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const { addNotification } = useUIStore();

  const limit = 12;

  const fetchRooms = async () => {
    try {
      setLoading(true);
      const params: any = {
        page,
        limit,
      };

      if (gameTypeFilter !== 'all') {
        params.game_type = gameTypeFilter;
      }
      if (statusFilter !== 'all') {
        params.status = statusFilter;
      }

      const response = await roomApi.getRooms(params);
      setRooms(response.rooms);
      setTotalPages(Math.ceil(response.total / limit));
    } catch (error: any) {
      console.error('Failed to fetch rooms:', error);
      addNotification({
        type: 'error',
        message: '获取房间列表失败',
      });
      // Set empty rooms on error
      setRooms([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRooms();
    
    // Poll every 5 seconds
    const interval = setInterval(fetchRooms, 5000);
    
    return () => clearInterval(interval);
  }, [page, gameTypeFilter, statusFilter]);

  const handleRefresh = () => {
    fetchRooms();
  };

  if (loading && rooms.length === 0) {
    return <Loading message="加载房间列表..." />;
  }

  return (
    <Box>
      <Box
        component={motion.div}
        initial={{ y: -10, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        sx={{ mb: 3, display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap' }}
      >
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>游戏类型</InputLabel>
          <Select
            value={gameTypeFilter}
            label="游戏类型"
            onChange={(e) => {
              setGameTypeFilter(e.target.value as any);
              setPage(1);
            }}
          >
            <MenuItem value="all">全部</MenuItem>
            <MenuItem value="ddz">斗地主</MenuItem>
            <MenuItem value="mahjong">麻将</MenuItem>
            <MenuItem value="durian">榴莲忘返</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>状态</InputLabel>
          <Select
            value={statusFilter}
            label="状态"
            onChange={(e) => {
              setStatusFilter(e.target.value as any);
              setPage(1);
            }}
          >
            <MenuItem value="all">全部</MenuItem>
            <MenuItem value="waiting">等待中</MenuItem>
            <MenuItem value="playing">游戏中</MenuItem>
          </Select>
        </FormControl>

        <Button
          variant="outlined"
          startIcon={<Refresh />}
          onClick={handleRefresh}
          disabled={loading}
        >
          刷新
        </Button>

        <Box sx={{ flexGrow: 1 }} />

        <Typography variant="body2" color="text.secondary">
          共 {rooms.length} 个房间
        </Typography>
      </Box>

      {rooms.length === 0 ? (
        <Box
          sx={{
            textAlign: 'center',
            py: 8,
          }}
        >
          <Typography variant="h6" color="text.secondary">
            暂无房间
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            创建一个新房间开始游戏吧！
          </Typography>
        </Box>
      ) : (
        <>
          <Grid container spacing={2}>
            {rooms.map((room, index) => (
              <Grid item xs={12} sm={6} md={4} lg={3} key={room.room_id}>
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.05 }}
                >
                  <RoomCard room={room} onJoin={onJoinRoom} />
                </motion.div>
              </Grid>
            ))}
          </Grid>

          {totalPages > 1 && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
              <Pagination
                count={totalPages}
                page={page}
                onChange={(_, value) => setPage(value)}
                color="primary"
              />
            </Box>
          )}
        </>
      )}
    </Box>
  );
};
