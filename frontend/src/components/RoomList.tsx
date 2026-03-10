import React, { useState, useEffect, useMemo } from 'react';
import { RoomInfo, GameType } from '../types/websocket';
import { wsService } from '../services/WebSocketService';
import { apiService } from '../services/ApiService';

interface RoomListProps {
    onJoinRoom: (roomId: string) => void;
    onCreateRoom: (gameType: GameType) => void;
}

const RoomList: React.FC<RoomListProps> = ({ onJoinRoom, onCreateRoom: _onCreateRoom }) => {
    const [rooms, setRooms] = useState<RoomInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState<'all' | GameType>('all');
    const [sortOption, setSortOption] = useState<'created' | 'players' | 'status'>('created');

    // 游戏类型映射
    const gameTypeInfo = {
        simple: { icon: '🃏', label: '简易卡牌', players: 2, color: 'blue' },
        ddz: { icon: '🂠', label: '斗地主', players: 3, color: 'red' },
        mahjong: { icon: '🀀', label: '麻将', players: 4, color: 'blue' },
        durian: { icon: '🍈', label: '榴莲忘返', players: 4, color: 'red' }
    };

    // 过滤和排序房间
    const filteredAndSortedRooms = useMemo(() => {
        let result = rooms.filter(room => {
            const matchesFilter = filter === 'all' || room.game_type === filter;
            return matchesFilter;
        });

        // 排序
        result.sort((a, b) => {
            switch (sortOption) {
                case 'created':
                    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
                case 'players':
                    return b.players.length - a.players.length;
                case 'status':
                    const statusOrder = { waiting: 0, playing: 1, gameover: 2 };
                    return (statusOrder[a.status] || 0) - (statusOrder[b.status] || 0);
                default:
                    return 0;
            }
        });

        return result;
    }, [rooms, filter, sortOption]);

    // 获取房间列表
    const fetchRooms = async () => {
        try {
            setLoading(true);
            // 调用后端API获取房间列表
            const backendRooms = await apiService.getRoomList();

            // 适配后端数据与前端类型
            const roomList: RoomInfo[] = backendRooms.map(room => ({
                id: room.id,
                game_type: room.game_type as GameType,
                // 后端只返回玩家数量，需要构造空数组
                players: Array.from({ length: room.players }, (_, i) => ({
                    id: `player-${room.id}-${i}`,
                    name: `玩家${i + 1}`,
                    is_ready: false
                })),
                max_players: room.max_players,
                // 后端使用 state 字段，前端使用 status
                status: room.state as 'waiting' | 'playing' | 'gameover',
                created_at: new Date().toISOString()
            }));

            setRooms(roomList);
        } catch (error) {
            console.error('获取房间列表失败:', error);
            // 失败时使用空数组
            setRooms([]);
        } finally {
            setLoading(false);
        }
    };

    // 获取状态样式
    const getStatusStyle = (status: string, color: string) => {
        if (color === 'blue') {
            switch (status) {
                case 'waiting':
                    return 'bg-blue-100 text-[#0F52BA]';
                case 'playing':
                    return 'bg-green-100 text-green-600';
                case 'gameover':
                    return 'bg-gray-100 text-gray-500';
                default:
                    return 'bg-gray-100 text-gray-500';
            }
        } else {
            switch (status) {
                case 'waiting':
                    return 'bg-red-100 text-[#E0115F]';
                case 'playing':
                    return 'bg-green-100 text-green-600';
                case 'gameover':
                    return 'bg-gray-100 text-gray-500';
                default:
                    return 'bg-gray-100 text-gray-500';
            }
        }
    };

    // 获取状态文本
    const getStatusText = (status: string) => {
        switch (status) {
            case 'waiting':
                return '等待中';
            case 'playing':
                return '游戏中';
            case 'gameover':
                return '已结束';
            default:
                return '未知';
        }
    };

    // 组件挂载时获取房间列表
    useEffect(() => {
        fetchRooms();
        const interval = setInterval(fetchRooms, 5000);

        const handleRoomStateChange = (data: any) => {
            console.log('房间状态变化:', data);
            fetchRooms();
        };

        wsService.on('roomStateChanged', handleRoomStateChange);

        return () => {
            clearInterval(interval);
            wsService.off('roomStateChanged', handleRoomStateChange);
        };
    }, []);

    if (loading) {
        return (
            <div className="flex justify-center items-center py-12">
                <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-[#0F52BA]"></div>
            </div>
        );
    }

    return (
        <div>
            {/* 筛选和排序 */}
            <div className="flex flex-wrap items-center gap-3 mb-4">
                {/* 游戏类型筛选 */}
                <div className="flex gap-2">
                    <button
                        onClick={() => setFilter('all')}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                            filter === 'all'
                                ? 'bg-gradient-to-r from-[#0F52BA] to-[#E0115F] text-white'
                                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        }`}
                    >
                        全部
                    </button>
                    {Object.entries(gameTypeInfo).map(([key, info]) => (
                        <button
                            key={key}
                            onClick={() => setFilter(key as GameType)}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-1 ${
                                filter === key
                                    ? info.color === 'blue'
                                        ? 'bg-[#0F52BA] text-white'
                                        : 'bg-[#E0115F] text-white'
                                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                            }`}
                        >
                            <span>{info.icon}</span>
                            <span>{info.label}</span>
                        </button>
                    ))}
                </div>

                {/* 排序下拉框 */}
                <select
                    value={sortOption}
                    onChange={(e) => setSortOption(e.target.value as any)}
                    className="px-4 py-2 bg-gray-100 border-0 rounded-lg text-sm text-gray-600 focus:outline-none focus:ring-2 focus:ring-[#0F52BA]/30 cursor-pointer"
                >
                    <option value="created">按创建时间</option>
                    <option value="players">按玩家数量</option>
                    <option value="status">按房间状态</option>
                </select>

                {/* 统计 */}
                <div className="ml-auto text-sm text-gray-500">
                    共 {filteredAndSortedRooms.length} 个房间
                </div>
            </div>

            {/* 房间列表 */}
            {filteredAndSortedRooms.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                    <div className="text-5xl mb-3">📭</div>
                    <p className="text-lg text-gray-500">
                        {filter !== 'all' ? '没有找到符合条件的房间' : '暂无房间'}
                    </p>
                    <p className="text-sm text-gray-400 mt-1">
                        {filter !== 'all' ? '请尝试调整筛选条件' : '点击"创建房间"开始游戏吧！'}
                    </p>
                </div>
            ) : (
                <div className="grid gap-3">
                    {filteredAndSortedRooms.map((room: RoomInfo) => {
                        const gameInfo = gameTypeInfo[room.game_type];
                        const isFull = room.players.length >= room.max_players;
                        const canJoin = room.status === 'waiting' && !isFull;

                        return (
                            <div
                                key={room.id}
                                className={`flex items-center gap-4 p-4 rounded-xl border-2 transition-all hover:shadow-md cursor-pointer ${
                                    gameInfo.color === 'blue'
                                        ? 'border-gray-100 hover:border-[#0F52BA]/30 bg-gray-50 hover:bg-blue-50/50'
                                        : 'border-gray-100 hover:border-[#E0115F]/30 bg-gray-50 hover:bg-red-50/50'
                                }`}
                                onClick={() => canJoin && onJoinRoom(room.id)}
                            >
                                {/* 游戏图标 */}
                                <div className={`w-14 h-14 rounded-xl flex items-center justify-center text-2xl ${
                                    gameInfo.color === 'blue'
                                        ? 'bg-[#0F52BA]/10'
                                        : 'bg-[#E0115F]/10'
                                }`}>
                                    {gameInfo.icon}
                                </div>

                                {/* 房间信息 */}
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2">
                                        <h3 className="font-bold text-gray-800">房间 {room.id}</h3>
                                        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${getStatusStyle(room.status, gameInfo.color)}`}>
                                            {getStatusText(room.status)}
                                        </span>
                                    </div>
                                    <p className="text-sm text-gray-500 mt-0.5">
                                        {gameInfo.label} · {room.players.length}/{room.max_players} 人
                                    </p>
                                </div>

                                {/* 加入按钮 */}
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        canJoin && onJoinRoom(room.id);
                                    }}
                                    disabled={!canJoin}
                                    className={`px-5 py-2 rounded-lg font-medium text-sm transition-all whitespace-nowrap ${
                                        !canJoin
                                            ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                                            : gameInfo.color === 'blue'
                                                ? 'bg-[#0F52BA] text-white hover:bg-[#0F52BA]/90 hover:shadow-lg'
                                                : 'bg-[#E0115F] text-white hover:bg-[#E0115F]/90 hover:shadow-lg'
                                    }`}
                                >
                                    {!canJoin ? (
                                        isFull ? '已满' :
                                            room.status === 'playing' ? '进行中' :
                                                room.status === 'gameover' ? '已结束' : '加入'
                                    ) : (
                                        '加入'
                                    )}
                                </button>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default RoomList;