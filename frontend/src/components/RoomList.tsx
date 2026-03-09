import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { RoomInfo, GameType } from '../types/websocket';
import { wsService } from '../services/WebSocketService';

interface RoomListProps {
    onJoinRoom: (roomId: string) => void;
    onCreateRoom: (gameType: GameType) => void;
}

const RoomList: React.FC<RoomListProps> = ({ onJoinRoom, onCreateRoom }) => {
    const [rooms, setRooms] = useState<RoomInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState<'all' | GameType>('all');
    const [sortOption, setSortOption] = useState<'created' | 'players' | 'status'>('created');
    const [searchTerm, setSearchTerm] = useState('');
    const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');

    // 游戏类型映射
    const gameTypeInfo = {
        simple: { icon: '🃏', label: '简易卡牌', players: 2 },
        ddz: { icon: '🂠', label: '斗地主', players: 3 },
        mahjong: { icon: '🀀', label: '麻将', players: 4 },
        durian: { icon: '🍈', label: '榴莲忘返', players: 4 }
    };

    // 防抖搜索
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearchTerm(searchTerm);
        }, 300);

        return () => clearTimeout(timer);
    }, [searchTerm]);

    // 过滤和排序房间
    const filteredAndSortedRooms = useMemo(() => {
        let result = rooms.filter(room => {
            const matchesFilter = filter === 'all' || room.game_type === filter;
            const matchesSearch = debouncedSearchTerm === '' ||
                room.id.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
                room.players.some(p => p.name.toLowerCase().includes(debouncedSearchTerm.toLowerCase())) ||
                gameTypeInfo[room.game_type].label.toLowerCase().includes(debouncedSearchTerm.toLowerCase());
            return matchesFilter && matchesSearch;
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
    }, [rooms, filter, debouncedSearchTerm, gameTypeInfo, sortOption]);

    // 获取房间列表
    const fetchRooms = async () => {
        try {
            setLoading(true);
            // TODO: 实现获取房间列表的API调用
            // 这里模拟一些房间数据
            const mockRooms: RoomInfo[] = [
                {
                    id: 'ROOM001',
                    game_type: 'simple',
                    players: [
                        { id: 'player1', name: '玩家1', is_ready: true },
                        { id: 'player2', name: '玩家2', is_ready: false }
                    ],
                    max_players: 2,
                    status: 'waiting',
                    created_at: new Date().toISOString()
                },
                {
                    id: 'ROOM002',
                    game_type: 'ddz',
                    players: [
                        { id: 'player3', name: '玩家3', is_ready: true },
                        { id: 'player4', name: '玩家4', is_ready: true },
                        { id: 'player5', name: '玩家5', is_ready: false }
                    ],
                    max_players: 3,
                    status: 'waiting',
                    created_at: new Date(Date.now() - 300000).toISOString()
                },
                {
                    id: 'ROOM003',
                    game_type: 'mahjong',
                    players: [
                        { id: 'player6', name: '玩家6', is_ready: true },
                        { id: 'player7', name: '玩家7', is_ready: true }
                    ],
                    max_players: 4,
                    status: 'playing',
                    created_at: new Date(Date.now() - 600000).toISOString()
                }
            ];
            setRooms(mockRooms);
        } catch (error) {
            console.error('获取房间列表失败:', error);
        } finally {
            setLoading(false);
        }
    };

    // 获取状态标签样式
    const getStatusStyle = (status: string) => {
        switch (status) {
            case 'waiting':
                return 'bg-yellow-500/20 text-yellow-300';
            case 'playing':
                return 'bg-green-500/20 text-green-300';
            case 'gameover':
                return 'bg-gray-500/20 text-gray-300';
            default:
                return 'bg-gray-500/20 text-gray-300';
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

        // 设置定时刷新
        const interval = setInterval(fetchRooms, 5000);

        // 监听房间状态变化
        const handleRoomStateChange = (data: any) => {
            console.log('房间状态变化:', data);
            fetchRooms(); // 重新获取房间列表
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
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 搜索和过滤 */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="flex-1">
                    <input
                        type="text"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        placeholder="搜索房间号或玩家名..."
                        className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
                    />
                </div>
                <select
                    value={filter}
                    onChange={(e) => setFilter(e.target.value as any)}
                    className="px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
                >
                    <option value="all">全部游戏</option>
                    {Object.entries(gameTypeInfo).map(([key, info]) => (
                        <option key={key} value={key}>
                            {info.icon} {info.label}
                        </option>
                    ))}
                </select>
                <select
                    value={sortOption}
                    onChange={(e) => setSortOption(e.target.value as any)}
                    className="px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
                >
                    <option value="created">按创建时间</option>
                    <option value="players">按玩家数量</option>
                    <option value="status">按房间状态</option>
                </select>
            </div>

            {/* 房间统计 */}
            <div className="flex gap-4 text-sm text-gray-400">
                <span>总房间数: {rooms.length}</span>
                <span>•</span>
                <span>可加入: {rooms.filter(r => r.status === 'waiting').length}</span>
                <span>•</span>
                <span>进行中: {rooms.filter(r => r.status === 'playing').length}</span>
            </div>

            {/* 房间列表 */}
            {filteredAndSortedRooms.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                    <div className="text-6xl mb-4">📭</div>
                    <p className="text-xl">
                        {searchTerm || filter !== 'all' ? '没有找到符合条件的房间' : '暂无房间'}
                    </p>
                    <p className="mt-2">
                        {searchTerm || filter !== 'all'
                            ? '请尝试调整搜索条件'
                            : '点击"创建房间"开始游戏吧！'}
                    </p>
                </div>
            ) : (
                <div className="grid gap-4">
                    {filteredAndSortedRooms.map((room: RoomInfo) => {
                        const gameInfo = gameTypeInfo[room.game_type];
                        const isFull = room.players.length >= room.max_players;
                        const canJoin = room.status === 'waiting' && !isFull;

                        return (
                            <div
                                key={room.id}
                                className="bg-gray-700 rounded-xl p-4 hover:bg-gray-600 transition-all duration-200 border border-gray-600 hover:border-gray-500"
                            >
                                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                                    {/* 房间基本信息 */}
                                    <div className="flex items-center gap-4 flex-1">
                                        <div className="text-3xl">{gameInfo.icon}</div>
                                        <div className="flex-1">
                                            <div className="flex items-center gap-2">
                                                <h3 className="font-bold text-lg">房间 {room.id}</h3>
                                                <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusStyle(room.status)}`}>
                                                    {getStatusText(room.status)}
                                                </span>
                                            </div>
                                            <p className="text-gray-300 mt-1">
                                                {gameInfo.label} · {room.players.length}/{room.max_players} 人
                                            </p>
                                            <p className="text-sm text-gray-400 mt-1">
                                                创建时间: {new Date(room.created_at).toLocaleString('zh-CN')}
                                            </p>
                                        </div>
                                    </div>

                                    {/* 玩家列表 */}
                                    <div className="flex-1 min-w-0">
                                        <div className="text-sm text-gray-400 mb-2">玩家列表:</div>
                                        <div className="flex flex-wrap gap-2">
                                            {room.players.map((player: any, index: number) => (
                                                <div
                                                    key={player.id}
                                                    className="flex items-center gap-2 bg-gray-600 px-3 py-1 rounded-full"
                                                >
                                                    <span className="text-xs text-gray-400">#{index + 1}</span>
                                                    <span className="font-medium">{player.name}</span>
                                                    {player.is_ready && (
                                                        <span className="text-green-400 text-xs">✓ 已准备</span>
                                                    )}
                                                </div>
                                            ))}
                                            {Array.from({ length: room.max_players - room.players.length }).map((_, index) => (
                                                <div
                                                    key={`empty-${index}`}
                                                    className="bg-gray-800 px-3 py-1 rounded-full text-gray-500 text-sm border border-dashed border-gray-600"
                                                >
                                                    等待玩家
                                                </div>
                                            ))}
                                        </div>
                                    </div>

                                    {/* 操作按钮 */}
                                    <div className="flex flex-col sm:flex-row gap-2 min-w-fit">
                                        <button
                                            onClick={() => onJoinRoom(room.id)}
                                            disabled={!canJoin}
                                            className={`px-4 py-2 rounded-lg font-medium transition-all whitespace-nowrap ${!canJoin
                                                ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                                                : 'bg-blue-600 hover:bg-blue-700 text-white hover:scale-105'
                                                }`}
                                        >
                                            {!canJoin ? (
                                                isFull ? '房间已满' :
                                                    room.status === 'playing' ? '游戏中' :
                                                        room.status === 'gameover' ? '已结束' : '无法加入'
                                            ) : (
                                                '加入游戏'
                                            )}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default RoomList;