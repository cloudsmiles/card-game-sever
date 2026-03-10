import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { wsService } from '../services/WebSocketService';
import { useApp } from '../context/AppContext';
import { GameType } from '../types/websocket';
import RoomList from '../components/RoomList';

const MagicalLobbyPage: React.FC = () => {
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
            alert('请输入你的名字！');
            return;
        }

        try {
            await wsService.connect(playerName);
            setIsConnected(true);
            setPlayerID(playerName);

            wsService.on('roomStateChanged', (data) => {
                console.log('房间状态更新:', data);
            });
        } catch (error) {
            console.error('连接服务器失败:', error);
            alert('连接服务器失败，请重试');
        }
    };

    // 创建房间
    const handleCreateRoom = () => {
        wsService.createRoom(selectedGameType);
        setShowCreateModal(false);
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

    // 直接加入房间
    const handleDirectJoin = () => {
        if (!joinRoomId.trim()) {
            alert('请输入房间号！');
            return;
        }
        wsService.joinRoom(joinRoomId);
        navigate(`/room/${joinRoomId}`);
    };

    // 游戏类型选项
    const gameTypes = [
        { value: 'simple', label: '简易卡牌', icon: '🃏', players: 2, color: 'blue' },
        { value: 'ddz', label: '斗地主', icon: '🎴', players: 3, color: 'red' },
        { value: 'mahjong', label: '麻将', icon: '🀄', players: 4, color: 'blue' },
        { value: 'durian', label: '榴莲忘返', icon: '🥭', players: 4, color: 'red' }
    ];

    return (
        <div className="min-h-screen bg-gray-50">
            {/* 顶部导航 */}
            <div className="bg-gradient-to-r from-[#0F52BA] to-[#E0115F] text-white py-3 px-6 shadow-lg">
                <div className="max-w-6xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <img src="/images/logo.png" alt="Logo" className="h-10 w-10 object-contain" />
                        <h1 className="text-xl font-bold">卡牌游戏大厅</h1>
                    </div>
                    {isConnected && (
                        <div className="flex items-center gap-4">
                            <span className="text-sm opacity-90">欢迎, {playerName}</span>
                            <div className="w-2 h-2 bg-green-400 rounded-full"></div>
                        </div>
                    )}
                </div>
            </div>

            <div className="max-w-6xl mx-auto px-4 py-6">
                {/* 连接区域 */}
                {!isConnected && (
                    <div className="max-w-md mx-auto">
                        <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-100">
                            <div className="text-center mb-8">
                                <div className="w-20 h-20 mx-auto mb-4 bg-gradient-to-br from-[#0F52BA] to-[#E0115F] rounded-full flex items-center justify-center text-4xl">
                                    🎮
                                </div>
                                <h2 className="text-2xl font-bold text-gray-800">进入游戏</h2>
                                <p className="text-gray-500 mt-2">请输入你的名字开始游戏</p>
                            </div>

                            <div className="space-y-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">你的名字</label>
                                    <input
                                        type="text"
                                        value={playerName}
                                        onChange={(e) => setPlayerName(e.target.value)}
                                        className="w-full px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-[#0F52BA] focus:outline-none transition-colors"
                                        placeholder="输入你的名字"
                                        onKeyPress={(e) => e.key === 'Enter' && handleConnect()}
                                    />
                                </div>
                                <button
                                    onClick={handleConnect}
                                    className="w-full py-4 bg-gradient-to-r from-[#0F52BA] to-[#E0115F] text-white font-bold rounded-xl hover:opacity-90 transition-opacity shadow-lg"
                                >
                                    开始游戏
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* 房间列表区域 */}
                {isConnected && (
                    <div>
                        <div className="bg-white rounded-2xl shadow-lg p-6 border border-gray-100">
                            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
                                <div className="flex items-center gap-3">
                                    <h2 className="text-xl font-bold text-gray-800">🏠 房间列表</h2>
                                    <span className="text-sm text-gray-500">点击房间加入游戏</span>
                                </div>
                                <div className="flex gap-3 w-full sm:w-auto">
                                    <input
                                        type="text"
                                        value={joinRoomId}
                                        onChange={(e) => setJoinRoomId(e.target.value)}
                                        className="flex-1 sm:w-40 px-4 py-2 rounded-lg border-2 border-gray-200 focus:border-[#0F52BA] focus:outline-none transition-colors text-sm"
                                        placeholder="输入房间号"
                                        onKeyPress={(e) => e.key === 'Enter' && handleDirectJoin()}
                                    />
                                    <button
                                        onClick={handleDirectJoin}
                                        className="px-4 py-2 bg-gradient-to-r from-[#E0115F] to-[#FF2D7A] text-white font-semibold rounded-lg hover:opacity-90 transition-opacity text-sm whitespace-nowrap"
                                    >
                                        加入
                                    </button>
                                    <button
                                        onClick={() => setShowCreateModal(true)}
                                        className="px-6 py-2 bg-gradient-to-r from-[#0F52BA] to-[#2E7BD0] text-white font-semibold rounded-lg hover:opacity-90 transition-opacity"
                                    >
                                        + 创建房间
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
                )}
            </div>

            {/* 创建房间模态框 */}
            {showCreateModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-2xl shadow-2xl p-8 max-w-md w-full">
                        <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center">创建房间</h2>

                        <div className="space-y-6">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-3">选择游戏类型</label>
                                <div className="grid grid-cols-2 gap-3">
                                    {gameTypes.map((game) => (
                                        <button
                                            key={game.value}
                                            onClick={() => setSelectedGameType(game.value as GameType)}
                                            className={`p-4 rounded-xl border-2 transition-all ${
                                                selectedGameType === game.value
                                                    ? game.color === 'blue'
                                                        ? 'border-[#0F52BA] bg-blue-50 ring-2 ring-[#0F52BA]/30'
                                                        : 'border-[#E0115F] bg-red-50 ring-2 ring-[#E0115F]/30'
                                                    : 'border-gray-200 hover:border-gray-300'
                                            }`}
                                        >
                                            <div className="text-3xl mb-1">{game.icon}</div>
                                            <div className="font-bold text-gray-800">{game.label}</div>
                                            <div className="text-sm text-gray-500">{game.players} 人</div>
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div className="flex gap-3 pt-4">
                                <button
                                    onClick={() => setShowCreateModal(false)}
                                    className="flex-1 py-3 px-4 bg-gray-400 hover:bg-gray-500 text-white font-semibold rounded-xl transition-colors"
                                >
                                    取消
                                </button>
                                <button
                                    onClick={handleCreateRoom}
                                    className="flex-1 py-3 px-4 bg-gradient-to-r from-[#0F52BA] to-[#E0115F] hover:opacity-90 text-white font-semibold rounded-xl transition-opacity"
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

export default MagicalLobbyPage;