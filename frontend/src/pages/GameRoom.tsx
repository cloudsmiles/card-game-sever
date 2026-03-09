import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { wsService } from '../services/WebSocketService';
import { useApp } from '../context/AppContext';
import { RoomInfo, GameType } from '../types/websocket';
const GameRoom: React.FC = () => {
    const { playerID } = useApp();
    const navigate = useNavigate();
    const { roomId } = useParams<{ roomId: string }>();
    const [roomInfo, setRoomInfo] = useState<RoomInfo | null>(null);
    const [gameState, setGameState] = useState<any>(null);
    const [isReady, setIsReady] = useState(false);
    const [gameLog, setGameLog] = useState<string[]>([]);
    const [selectedCards, setSelectedCards] = useState<string[]>([]);

    // 游戏类型信息
    const gameTypeInfo = {
        simple: { name: '简易卡牌', icon: '🃏', maxPlayers: 2 },
        ddz: { name: '斗地主', icon: '🂠', maxPlayers: 3 },
        mahjong: { name: '麻将', icon: '🀀', maxPlayers: 4 },
        durian: { name: '榴莲忘返', icon: '🍈', maxPlayers: 4 }
    };

    // 加入房间
    useEffect(() => {
        if (roomId && playerID) {
            try {
                wsService.joinRoom(roomId);
                wsService.setCurrentRoomID(roomId);

                // 监听房间状态变化
                wsService.on('roomStateChanged', handleRoomStateChange);
                wsService.on('gameStateUpdate', handleGameStateUpdate);
                wsService.on('error', handleError);

                addToLog(`加入房间 ${roomId}`);
            } catch (error) {
                console.error('加入房间失败:', error);
                navigate('/');
            }
        }

        return () => {
            wsService.off('roomStateChanged', handleRoomStateChange);
            wsService.off('gameStateUpdate', handleGameStateUpdate);
            wsService.off('error', handleError);
        };
    }, [roomId, playerID, navigate]);

    // 处理房间状态变化
    const handleRoomStateChange = (data: any) => {
        console.log('房间状态更新:', data);
        if (data.room) {
            setRoomInfo(data.room);
            addToLog(`房间状态更新: ${data.room.players.length}/${data.room.max_players} 人`);
        }
    };

    // 处理游戏状态更新
    const handleGameStateUpdate = (data: any) => {
        console.log('游戏状态更新:', data);
        setGameState(data);
        addToLog('游戏状态已更新');
    };

    // 处理错误
    const handleError = (error: any) => {
        console.error('游戏错误:', error);
        addToLog(`错误: ${error.message || '未知错误'}`);
    };

    // 添加日志
    const addToLog = (message: string) => {
        const timestamp = new Date().toLocaleTimeString();
        setGameLog(prev => [...prev.slice(-49), `[${timestamp}] ${message}`]);
    };

    // 准备/取消准备
    const toggleReady = () => {
        if (!roomInfo) return;

        const action = isReady ? 'unready' : 'ready';
        wsService.roomAction(action);
        setIsReady(!isReady);
        addToLog(isReady ? '取消准备' : '准备就绪');
    };

    // 开始游戏
    const startGame = () => {
        if (!roomInfo) return;

        // 检查是否所有玩家都已准备
        const allReady = roomInfo.players.every(p => p.is_ready);
        if (!allReady) {
            addToLog('请等待所有玩家准备就绪');
            return;
        }

        wsService.gameAction('start');
        addToLog('开始游戏');
    };

    // 出牌
    const playCard = (card: string) => {
        if (!gameState?.current_turn || gameState.current_turn !== playerID) {
            addToLog('不是你的回合');
            return;
        }

        wsService.gameAction('play', { card });
        setSelectedCards([]);
        addToLog(`出牌: ${card}`);
    };

    // 选择卡牌
    const selectCard = (card: string) => {
        setSelectedCards(prev => {
            if (prev.includes(card)) {
                return prev.filter(c => c !== card);
            } else {
                return [...prev, card];
            }
        });
    };

    // 离开房间
    const leaveRoom = () => {
        wsService.gameAction('leave');
        navigate('/');
    };

    if (!roomInfo) {
        return (
            <div className="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto mb-4"></div>
                    <p className="text-white text-xl">正在加入房间...</p>
                </div>
            </div>
        );
    }

    const gameInfo = gameTypeInfo[roomInfo.game_type as GameType] || gameTypeInfo.simple;

    return (
        <div className="min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900 text-white">
            <div className="container mx-auto px-4 py-6">
                {/* 头部信息 */}
                <div className="bg-gray-800 rounded-2xl p-6 mb-6 shadow-2xl">
                    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                        <div>
                            <h1 className="text-3xl font-bold mb-2">
                                {gameInfo.icon} {gameInfo.name}房间
                            </h1>
                            <div className="flex items-center gap-4 text-gray-300">
                                <span>房间号: <span className="font-mono text-green-400">{roomInfo.id}</span></span>
                                <span>玩家: {roomInfo.players.length}/{roomInfo.max_players}</span>
                                <span className={`px-3 py-1 rounded-full text-sm ${roomInfo.status === 'waiting' ? 'bg-yellow-500/20 text-yellow-300' :
                                        roomInfo.status === 'playing' ? 'bg-green-500/20 text-green-300' :
                                            'bg-gray-500/20 text-gray-300'
                                    }`}>
                                    {roomInfo.status === 'waiting' ? '等待中' :
                                        roomInfo.status === 'playing' ? '游戏中' : '已结束'}
                                </span>
                            </div>
                        </div>
                        <button
                            onClick={leaveRoom}
                            className="bg-red-600 hover:bg-red-700 px-6 py-3 rounded-xl font-bold transition-colors"
                        >
                            离开房间
                        </button>
                    </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {/* 主游戏区域 */}
                    <div className="lg:col-span-2 space-y-6">
                        {/* 游戏状态面板 */}
                        {roomInfo.status === 'waiting' && (
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                                <h2 className="text-2xl font-bold mb-6 text-center">⏳ 等待开始</h2>

                                {/* 准备按钮 */}
                                <div className="text-center mb-6">
                                    <button
                                        onClick={toggleReady}
                                        className={`px-8 py-4 rounded-xl font-bold text-lg transition-all transform hover:scale-105 ${isReady
                                                ? 'bg-green-600 hover:bg-green-700'
                                                : 'bg-blue-600 hover:bg-blue-700'
                                            }`}
                                    >
                                        {isReady ? '✅ 已准备' : '🎮 准备游戏'}
                                    </button>
                                </div>

                                {/* 开始游戏按钮 */}
                                {roomInfo.players.every(p => p.is_ready) && roomInfo.players.length >= 2 && (
                                    <div className="text-center">
                                        <button
                                            onClick={startGame}
                                            className="bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 px-8 py-4 rounded-xl font-bold text-lg transition-all transform hover:scale-105 shadow-lg"
                                        >
                                            🚀 开始游戏
                                        </button>
                                    </div>
                                )}
                            </div>
                        )}

                        {/* 游戏进行中 */}
                        {roomInfo.status === 'playing' && gameState && (
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                                <div className="flex justify-between items-center mb-6">
                                    <div>
                                        <span className="text-gray-400">当前回合：</span>
                                        <span className="font-bold text-xl text-yellow-400">
                                            {gameState.current_turn === playerID ? '你的回合' : gameState.current_turn}
                                        </span>
                                    </div>
                                    {gameState.last_card && (
                                        <div>
                                            <span className="text-gray-400">上家出牌：</span>
                                            <span className="font-mono text-3xl font-bold text-yellow-400">
                                                {gameState.last_card}
                                            </span>
                                        </div>
                                    )}
                                </div>

                                {/* 手牌区域 */}
                                <div className="mb-6">
                                    <div className="text-lg font-bold mb-3">你的手牌</div>
                                    <div className="flex flex-wrap gap-2">
                                        {(gameState.hands?.[playerID] || []).map((card: string, index: number) => (
                                            <button
                                                key={`${card}-${index}`}
                                                onClick={() => playCard(card)}
                                                className={`px-4 py-3 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl font-bold text-lg transition-all transform hover:scale-110 hover:shadow-lg ${gameState.current_turn === playerID
                                                        ? 'hover:-translate-y-1 cursor-pointer'
                                                        : 'opacity-50 cursor-not-allowed'
                                                    }`}
                                            >
                                                {card}
                                            </button>
                                        ))}
                                    </div>
                                </div>

                                {/* 选牌区域（多选模式） */}
                                {gameState.multi_select && (
                                    <div className="mb-6">
                                        <div className="flex justify-between items-center mb-3">
                                            <div className="text-lg font-bold">选择卡牌</div>
                                            <button
                                                onClick={() => wsService.gameAction('play_selected', { cards: selectedCards })}
                                                disabled={selectedCards.length === 0}
                                                className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed px-4 py-2 rounded-lg font-medium transition-colors"
                                            >
                                                出牌 ({selectedCards.length})
                                            </button>
                                        </div>
                                        <div className="flex flex-wrap gap-2">
                                            {(gameState.hands?.[playerID] || []).map((card: string, index: number) => (
                                                <button
                                                    key={`select-${card}-${index}`}
                                                    onClick={() => selectCard(card)}
                                                    className={`px-4 py-3 rounded-xl font-bold text-lg transition-all border-2 ${selectedCards.includes(card)
                                                            ? 'bg-blue-600 border-blue-400 scale-105'
                                                            : 'bg-gray-700 border-gray-600 hover:bg-gray-600'
                                                        }`}
                                                >
                                                    {card}
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}

                        {/* 游戏日志 */}
                        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                            <h3 className="text-xl font-bold mb-4">游戏副本日志</h3>
                            <div className="bg-black/50 h-48 overflow-y-auto rounded-xl p-4 font-mono text-sm">
                                {gameLog.map((log, index) => (
                                    <div key={index} className="py-1 text-gray-300">
                                        {log}
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* 右侧面板 */}
                    <div className="space-y-6">
                        {/* 玩家列表 */}
                        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                            <h3 className="text-xl font-bold mb-4 flex items-center gap-2">
                                👥 玩家列表 ({roomInfo.players.length}/{roomInfo.max_players})
                            </h3>
                            <div className="space-y-3">
                                {roomInfo.players.map((player) => (
                                    <div
                                        key={player.id}
                                        className={`p-3 rounded-xl flex items-center justify-between ${player.id === playerID
                                                ? 'bg-blue-500/20 border border-blue-500'
                                                : 'bg-gray-700'
                                            }`}
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className={`w-3 h-3 rounded-full ${player.is_ready ? 'bg-green-500' : 'bg-gray-500'
                                                }`}></div>
                                            <span className="font-medium">
                                                {player.name} {player.id === playerID && '(你)'}
                                            </span>
                                        </div>
                                        {player.is_ready && (
                                            <span className="text-green-400 text-sm">✅ 已准备</span>
                                        )}
                                    </div>
                                ))}
                                {Array.from({ length: roomInfo.max_players - roomInfo.players.length }).map((_, index) => (
                                    <div
                                        key={`empty-${index}`}
                                        className="p-3 bg-gray-700 rounded-xl text-gray-500 border border-dashed border-gray-600 flex items-center gap-3"
                                    >
                                        <div className="w-3 h-3 rounded-full bg-gray-600"></div>
                                        等待玩家加入
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* 聊天区域 */}
                        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                            <h3 className="text-xl font-bold mb-4">💬 聊天</h3>
                            <div className="space-y-3">
                                <div className="bg-black/30 h-32 overflow-y-auto rounded-xl p-3 text-sm">
                                    {/* 聊天消息显示区域 */}
                                    <div className="text-gray-400 text-center py-8">暂无消息</div>
                                </div>
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        placeholder="输入消息..."
                                        className="flex-1 px-3 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    />
                                    <button className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-medium">
                                        发送
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default GameRoom;