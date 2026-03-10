import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { wsService } from '../services/WebSocketService';
import { useApp } from '../context/AppContext';
import { GameType } from '../types/websocket';
import RoomList from '../components/RoomList';

const LobbyPage: React.FC = () => {
    const navigate = useNavigate();
    const { setPlayerID } = useApp();
    const [playerName, setPlayerName] = useState('');
    const [isConnected, setIsConnected] = useState(false);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [selectedGameType, setSelectedGameType] = useState<GameType>('simple');
    const [joinRoomId, setJoinRoomId] = useState('');

    // 连接WebSocket
    const handleConnect = async () => {
        if (!playerName.trim()) {
            alert('请输入玩家昵称');
            return;
        }

        try {
            await wsService.connect(playerName);
            setIsConnected(true);
            setPlayerID(playerName); // 设置全局玩家ID

            // 监听房间状态变化
            wsService.on('roomStateChanged', (data) => {
                console.log('房间状态更新:', data);
                // 房间列表组件会自动处理更新
            });
        } catch (error) {
            console.error('连接失败:', error);
            alert('连接服务器失败');
        }
    };

    // 创建房间
    const handleCreateRoom = () => {
        wsService.createRoom(selectedGameType);
        setShowCreateModal(false);
        // 等待服务器返回房间ID后导航
        wsService.on('roomStateChanged', (data) => {
            if (data.room_id) {
                navigate(`/room/${data.room_id}?gameType=${selectedGameType}`);
            }
        });
    };

    // 加入房间
    const handleJoinRoom = (roomId: string) => {
        wsService.joinRoom(roomId);
        navigate(`/room/${roomId}`);
    };

    // 直接加入房间（通过输入框）
    const handleDirectJoin = () => {
        if (!joinRoomId.trim()) {
            alert('请输入房间号');
            return;
        }
        wsService.joinRoom(joinRoomId);
        navigate(`/room/${joinRoomId}`);
    };

    // 游戏类型选项
    const gameTypes = [
        { value: 'simple', label: '简易卡牌', icon: '🃏', players: 2 },
        { value: 'ddz', label: '斗地主', icon: '🂠', players: 3 },
        { value: 'mahjong', label: '麻将', icon: '🀀', players: 4 },
        { value: 'durian', label: '榴莲忘返', icon: '🍈', players: 4 }
    ];

    return (
        <div className="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 text-white">
            <div className="container mx-auto px-4 py-8">
                {/* 头部 */}
                <div className="text-center mb-12">
                    <h1 className="text-5xl font-bold mb-4 bg-clip-text text-transparent bg-gradient-to-r from-yellow-400 to-pink-500">
                        🎴 卡牌游戏大厅
                    </h1>
                    <p className="text-xl text-gray-300">选择游戏，创建房间，与朋友一起享受卡牌乐趣！</p>
                </div>

                {/* 连接区域 */}
                {!isConnected && (
                    <div className="max-w-md mx-auto bg-gray-800 rounded-2xl p-8 shadow-2xl backdrop-blur-sm">
                        <h2 className="text-2xl font-bold mb-6 text-center">🎮 连接到游戏</h2>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-2">玩家昵称</label>
                                <input
                                    type="text"
                                    value={playerName}
                                    onChange={(e) => setPlayerName(e.target.value)}
                                    className="w-full px-4 py-3 bg-gray-700 border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
                                    placeholder="输入你的昵称"
                                />
                            </div>
                            <button
                                onClick={handleConnect}
                                className="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white font-bold py-3 px-4 rounded-lg transition-all duration-200 transform hover:scale-105 shadow-lg"
                            >
                                🔌 连接服务器
                            </button>
                        </div>
                    </div>
                )}

                {/* 大厅功能区域 */}
                {isConnected && (
                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                        {/* 左侧：房间列表 */}
                        <div className="lg:col-span-2">
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl backdrop-blur-sm">
                                <div className="flex justify-between items-center mb-6">
                                    <h2 className="text-2xl font-bold">🎲 当前房间</h2>
                                    <button
                                        onClick={() => setShowCreateModal(true)}
                                        className="bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white font-bold py-2 px-4 rounded-lg transition-all duration-200 transform hover:scale-105"
                                    >
                                        ✨ 创建房间
                                    </button>
                                </div>

                                {/* 房间搜索 */}
                                <div className="mb-6">
                                    <div className="flex gap-2">
                                        <input
                                            type="text"
                                            value={joinRoomId}
                                            onChange={(e) => setJoinRoomId(e.target.value)}
                                            className="flex-1 px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
                                            placeholder="输入房间号加入游戏"
                                        />
                                        <button
                                            onClick={handleDirectJoin}
                                            className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-6 rounded-lg transition-colors"
                                        >
                                            加入
                                        </button>
                                    </div>
                                </div>

                                {/* 房间列表组件 */}
                                <RoomList
                                    onJoinRoom={handleJoinRoom}
                                    onCreateRoom={setSelectedGameType}
                                />
                            </div>
                        </div>

                        {/* 右侧：游戏介绍 */}
                        <div className="space-y-6">
                            {/* 玩家信息 */}
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl backdrop-blur-sm">
                                <h3 className="text-xl font-bold mb-4">👤 玩家信息</h3>
                                <div className="space-y-3">
                                    <div className="flex justify-between">
                                        <span className="text-gray-300">昵称:</span>
                                        <span className="font-medium">{playerName}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-gray-300">状态:</span>
                                        <span className="text-green-400 font-medium">🟢 已连接</span>
                                    </div>
                                </div>
                            </div>

                            {/* 游戏类型介绍 */}
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl backdrop-blur-sm">
                                <h3 className="text-xl font-bold mb-4">🎯 游戏类型</h3>
                                <div className="space-y-3">
                                    {gameTypes.map((game) => (
                                        <div key={game.value} className="flex items-center gap-3 p-3 bg-gray-700 rounded-lg">
                                            <span className="text-2xl">{game.icon}</span>
                                            <div className="flex-1">
                                                <div className="font-medium">{game.label}</div>
                                                <div className="text-sm text-gray-400">{game.players} 人游戏</div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* 创建房间模态框 */}
            {showCreateModal && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-50">
                    <div className="bg-gray-800 rounded-2xl p-8 max-w-md w-full shadow-2xl">
                        <h2 className="text-2xl font-bold mb-6 text-center">✨ 创建房间</h2>

                        <div className="space-y-6">
                            <div>
                                <label className="block text-sm font-medium mb-3">选择游戏类型</label>
                                <div className="grid grid-cols-2 gap-3">
                                    {gameTypes.map((game) => (
                                        <button
                                            key={game.value}
                                            onClick={() => setSelectedGameType(game.value as GameType)}
                                            className={`p-4 rounded-xl border-2 transition-all ${selectedGameType === game.value
                                                ? 'border-blue-500 bg-blue-500/20'
                                                : 'border-gray-600 hover:border-gray-500'
                                                }`}
                                        >
                                            <div className="text-3xl mb-2">{game.icon}</div>
                                            <div className="font-medium">{game.label}</div>
                                            <div className="text-sm text-gray-400">{game.players} 人</div>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="flex gap-3 pt-4">
                                <button
                                    onClick={() => setShowCreateModal(false)}
                                    className="flex-1 py-3 px-4 bg-gray-600 hover:bg-gray-700 rounded-lg font-medium transition-colors"
                                >
                                    取消
                                </button>
                                <button
                                    onClick={handleCreateRoom}
                                    className="flex-1 py-3 px-4 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 rounded-lg font-medium transition-all transform hover:scale-105"
                                >
                                    创建房间
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default LobbyPage;