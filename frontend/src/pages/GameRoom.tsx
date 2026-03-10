import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { wsService } from '../services/WebSocketService';
import { useApp } from '../context/AppContext';
import { RoomInfo, RoomStateContent, GameType } from '../types/websocket';
import { DDZGame, SimpleGame, MahjongGame } from '../components/game';

const GameRoom: React.FC = () => {
    const { playerID } = useApp();
    const navigate = useNavigate();
    const { roomId } = useParams<{ roomId: string }>();
    const [searchParams] = useSearchParams();
    const gameTypeFromUrl = searchParams.get('gameType') as GameType || 'simple';
    const [roomInfo, setRoomInfo] = useState<RoomInfo | null>(null);
    const [gameState, setGameState] = useState<any>(null);
    const [isReady, setIsReady] = useState(false);
    const [gameLog, setGameLog] = useState<string[]>([]);
    const [chatMessages, setChatMessages] = useState<{ player: string; content: string }[]>([]);
    const [chatInput, setChatInput] = useState('');

    // 游戏类型信息
    const gameTypeInfo = {
        simple: { name: '简易卡牌', icon: '🃏', maxPlayers: 2 },
        ddz: { name: '斗地主', icon: '🂠', maxPlayers: 3 },
        mahjong: { name: '麻将', icon: '🀀', maxPlayers: 4 },
        durian: { name: '榴莲忘返', icon: '🍈', maxPlayers: 4 }
    };

    // 添加日志
    const addToLog = (message: string) => {
        const timestamp = new Date().toLocaleTimeString();
        setGameLog(prev => [...prev.slice(-49), `[${timestamp}] ${message}`]);
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
                wsService.on('chatMessage', handleChatMessage);
                wsService.on('gameStarted', handleGameStarted);
                wsService.on('gameOver', handleGameOver);

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
            wsService.off('chatMessage', handleChatMessage);
            wsService.off('gameStarted', handleGameStarted);
            wsService.off('gameOver', handleGameOver);
        };
    }, [roomId, playerID, navigate]);

    // 处理聊天消息
    const handleChatMessage = (data: any) => {
        const player = data.player_id || data.PlayerID || '未知玩家';
        const content = data.content || data.Content || '';
        setChatMessages(prev => [...prev.slice(-49), { player, content }]);
    };

    // 发送聊天消息
    const handleSendChat = () => {
        if (!chatInput.trim() || !roomId) return;
        wsService.sendChat(chatInput);
        setChatMessages(prev => [...prev.slice(-49), { player: playerID, content: chatInput }]);
        setChatInput('');
    };

    // 处理房间状态变化
    const handleRoomStateChange = (data: RoomStateContent) => {
        console.log('房间状态更新:', data);

        // 从 RoomStateContent 构建 RoomInfo
        // data 直接就是 RoomStateContent 结构（根据 PROTOCOL.md）
        const newRoomInfo: RoomInfo = {
            id: data.room_id,
            game_type: roomInfo?.game_type || gameTypeFromUrl || 'simple',
            players: data.players.map(p => ({
                id: p.player_id,
                name: p.player_id,
                is_ready: p.ready,
                seat_index: p.seat_number
            })),
            max_players: Math.max(data.players.length, roomInfo?.max_players || 2),
            status: data.state,
            created_at: roomInfo?.created_at || new Date().toISOString()
        };

        setRoomInfo(newRoomInfo);
        addToLog(`房间状态更新: ${data.players.length}/${newRoomInfo.max_players} 人`);

        // 更新准备状态
        const myPlayer = data.players.find(p => p.player_id === playerID);
        if (myPlayer) {
            setIsReady(myPlayer.ready);
        }

        if (data.message) {
            addToLog(data.message);
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
        const errorMsg = error?.message || error?.Message || error?.error || '未知错误';
        addToLog(`错误: ${errorMsg}`);
    };

    // 处理游戏开始
    const handleGameStarted = (_data: any) => {
        addToLog('🎮 游戏开始！');
    };

    // 处理游戏结束
    const handleGameOver = (data: any) => {
        const winner = data.winner || data.Winner || '未知';
        addToLog(`🏆 游戏结束！获胜者: ${winner}`);
        alert(`游戏结束！获胜者：${winner}`);
    };

    // 准备/取消准备
    const toggleReady = () => {
        if (!roomInfo) return;

        wsService.roomAction('ready', { ready: !isReady });
        setIsReady(!isReady);
        addToLog(isReady ? '取消准备' : '准备就绪');
    };

    // 开始游戏
    const startGame = () => {
        if (!roomInfo) return;

        const allReady = roomInfo.players.every(p => p.is_ready);
        if (!allReady) {
            addToLog('请等待所有玩家准备就绪');
            return;
        }

        wsService.gameAction('start');
        addToLog('开始游戏');
    };

    // 离开房间
    const leaveRoom = () => {
        wsService.gameAction('leave');
        navigate('/');
    };

    // 渲染游戏组件
    const renderGame = () => {
        if (!gameState || !roomInfo) return null;

        switch (roomInfo.game_type) {
            case 'ddz':
                return <DDZGame gameState={gameState} playerID={playerID} onLog={addToLog} />;
            case 'simple':
                return <SimpleGame gameState={gameState} playerID={playerID} onLog={addToLog} />;
            case 'mahjong':
                return <MahjongGame gameState={gameState} playerID={playerID} onLog={addToLog} />;
            default:
                return <SimpleGame gameState={gameState} playerID={playerID} onLog={addToLog} />;
        }
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
                        {/* 等待开始面板 */}
                        {roomInfo.status === 'waiting' && (
                            <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                                <h2 className="text-2xl font-bold mb-6 text-center">⏳ 等待开始</h2>

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
                        {roomInfo.status === 'playing' && renderGame()}

                        {/* 游戏日志 */}
                        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
                            <h3 className="text-xl font-bold mb-4">游戏日志</h3>
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
                        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl flex flex-col h-80">
                            <h3 className="text-xl font-bold mb-4">💬 聊天</h3>
                            <div className="flex-1 overflow-y-auto space-y-2 text-sm mb-4">
                                {chatMessages.length === 0 ? (
                                    <div className="text-gray-400 text-center py-8">暂无消息</div>
                                ) : (
                                    chatMessages.map((msg, index) => (
                                        <div key={index} className="flex gap-2">
                                            <span className="text-blue-400 font-medium">{msg.player}：</span>
                                            <span>{msg.content}</span>
                                        </div>
                                    ))
                                )}
                            </div>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    value={chatInput}
                                    onChange={(e) => setChatInput(e.target.value)}
                                    onKeyPress={(e) => e.key === 'Enter' && handleSendChat()}
                                    placeholder="输入消息..."
                                    className="flex-1 px-3 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                />
                                <button
                                    onClick={handleSendChat}
                                    className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-medium"
                                >
                                    发送
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default GameRoom;