import React, { useState } from 'react';
import { wsService } from '../../services/WebSocketService';

// 卡牌值转字符串
const valueToStr = (v: number): string => {
    if (v >= 3 && v <= 10) return String(v);
    if (v === 11) return 'J';
    if (v === 12) return 'Q';
    if (v === 13) return 'K';
    if (v === 14) return 'A';
    if (v === 15) return '2';
    if (v === 16) return '小王';
    if (v === 17) return '大王';
    return '?';
};

interface SelectedCard {
    id: string;
    value: number;
    index: number;
}

interface DDZGameProps {
    gameState: any;
    playerID: string;
    onLog: (message: string) => void;
}

const DDZGame: React.FC<DDZGameProps> = ({ gameState, playerID, onLog }) => {
    const [selectedCards, setSelectedCards] = useState<SelectedCard[]>([]);
    const [isProcessingAction, setIsProcessingAction] = useState(false);

    // 获取阶段显示文字
    const getPhaseText = (phase: string) => {
        if (phase === 'call') return '叫地主';
        if (phase === 'play') return '出牌';
        return '等待中';
    };

    // 获取上家出牌显示
    const getLastPlayDisplay = (lastPlay: any) => {
        if (!lastPlay) return '无';
        const cards = lastPlay.Cards || lastPlay.cards || [];
        const type = lastPlay.Type || lastPlay.type || '';
        const cardsDisplay = cards.length > 0
            ? cards.map((c: any) => valueToStr(c.Value || c.value || c)).join(', ')
            : '';
        return `${type} (${cardsDisplay})`;
    };

    // 选择卡牌
    const selectCard = (card: number, index: number) => {
        const cardId = `${card}_${index}`;
        const selectedIndex = selectedCards.findIndex(c => c.id === cardId);

        if (selectedIndex > -1) {
            setSelectedCards(prev => prev.filter(c => c.id !== cardId));
        } else {
            setSelectedCards(prev => [...prev, { id: cardId, value: card, index }]);
        }
    };

    // 出牌
    const playSelected = () => {
        if (selectedCards.length === 0) {
            alert('请选择至少一张牌');
            return;
        }
        if (isProcessingAction) return;
        setIsProcessingAction(true);

        const cardsToPlay = selectedCards.map(card => ({ value: card.value }));
        wsService.gameAction('play_cards', { card: cardsToPlay });

        onLog(`出牌: ${selectedCards.map(c => valueToStr(c.value)).join(', ')}`);

        setSelectedCards([]);
        setTimeout(() => setIsProcessingAction(false), 300);
    };

    // 不出
    const passTurn = () => {
        if (isProcessingAction) return;
        setIsProcessingAction(true);
        wsService.gameAction('pass');
        onLog('你选择不出');
        setTimeout(() => setIsProcessingAction(false), 300);
    };

    // 叫地主
    const callLandlord = (score: number) => {
        if (isProcessingAction) return;
        setIsProcessingAction(true);
        wsService.gameAction('call_landlord', { card: score });
        onLog(`你叫了 ${score} 分`);
        setTimeout(() => setIsProcessingAction(false), 300);
    };

    // 清空选择
    const clearSelection = () => {
        setSelectedCards([]);
        onLog('已清空选择');
    };

    // 渲染手牌
    const renderHand = (hand: number[]) => {
        const sortedHand = [...hand].sort((a, b) => a - b);
        return sortedHand.map((cardValue, index) => {
            const cardId = `${cardValue}_${index}`;
            const isSelected = selectedCards.some(c => c.id === cardId);
            return (
                <div
                    key={cardId}
                    onClick={() => selectCard(cardValue, index)}
                    className={`card bg-white text-black w-12 h-20 flex items-center justify-center text-2xl font-bold rounded-lg shadow-xl border-2 border-gray-300 cursor-pointer transition-all ${isSelected
                            ? 'border-blue-500 transform -translate-y-3'
                            : 'hover:-translate-y-2 hover:shadow-lg'
                        }`}
                >
                    {valueToStr(cardValue)}
                </div>
            );
        });
    };

    // 渲染操作按钮
    const renderActions = () => {
        const phase = gameState?.phase || gameState?.Phase || '';
        const currentTurn = gameState?.current || gameState?.current_turn || gameState?.CurrentTurn || '';

        if (currentTurn !== playerID) return null;

        if (phase === 'call') {
            return (
                <div className="flex gap-4 mb-4">
                    <button onClick={() => callLandlord(0)} className="bg-gray-600 hover:bg-gray-500 px-6 py-2 rounded-xl">不叫</button>
                    <button onClick={() => callLandlord(1)} className="bg-blue-600 hover:bg-blue-500 px-6 py-2 rounded-xl">叫 1 分</button>
                    <button onClick={() => callLandlord(2)} className="bg-blue-600 hover:bg-blue-500 px-6 py-2 rounded-xl">叫 2 分</button>
                    <button onClick={() => callLandlord(3)} className="bg-blue-600 hover:bg-blue-500 px-6 py-2 rounded-xl">叫 3 分</button>
                    <button onClick={clearSelection} className="bg-gray-500 hover:bg-gray-400 px-6 py-2 rounded-xl">清空选择</button>
                </div>
            );
        }

        if (phase === 'play') {
            return (
                <div className="flex gap-4 mb-4">
                    <button
                        onClick={playSelected}
                        disabled={selectedCards.length === 0}
                        className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed px-6 py-2 rounded-xl"
                    >
                        出牌
                    </button>
                    <button onClick={passTurn} className="bg-red-600 hover:bg-red-700 px-6 py-2 rounded-xl">不出</button>
                    <button onClick={clearSelection} className="bg-gray-500 hover:bg-gray-400 px-6 py-2 rounded-xl">清空选择</button>
                </div>
            );
        }

        return null;
    };

    const currentTurn = gameState?.current || gameState?.current_turn || gameState?.CurrentTurn || '';
    const isMyTurn = currentTurn === playerID;

    return (
        <div className="bg-gray-800 rounded-2xl p-6 shadow-2xl">
            {/* 游戏状态信息 */}
            <div className="flex justify-between items-center mb-6">
                <div>
                    <span className="text-gray-400">当前回合：</span>
                    <span className={`font-bold text-xl ${isMyTurn ? 'text-green-400 animate-pulse' : 'text-yellow-400'}`}>
                        {isMyTurn ? '你的回合' : currentTurn || '等待中'}
                    </span>
                </div>
                <div>
                    <span className="text-gray-400">上家出牌：</span>
                    <span className="font-mono text-3xl font-bold text-yellow-400">
                        {getLastPlayDisplay(gameState.last_play || gameState.lastCard || gameState.LastPlay)}
                    </span>
                </div>
                <div>
                    <span className="text-gray-400">阶段：</span>
                    <span className="font-bold text-xl">{getPhaseText(gameState.phase || gameState.Phase)}</span>
                </div>
                <div>
                    <span className="text-gray-400">地主：</span>
                    <span className="font-bold text-xl">{(gameState.landlord || gameState.Landlord) || '未定'}</span>
                </div>
            </div>

            {/* 手牌区域 */}
            <div className="mb-4">
                <div className="text-sm text-gray-400 mb-2">你的手牌（点击选择）</div>
                <div className="flex gap-1 flex-wrap overflow-x-auto">
                    {renderHand(gameState.my_hand || gameState.MyHand || [])}
                </div>
            </div>

            {/* 操作按钮 */}
            {renderActions()}

            {/* 玩家信息 */}
            <div className="mt-6">
                <div className="text-sm text-gray-400 mb-2">玩家状态</div>
                <div className="grid grid-cols-3 gap-4">
                    {(gameState.players || gameState.Players || []).map((player: string, index: number) => (
                        <div key={index} className={`p-3 rounded-xl text-center ${player === playerID ? 'bg-blue-500/20 border border-blue-500' : 'bg-gray-700'}`}>
                            <div className="font-medium">
                                {player}
                                {player === playerID && ' (你)'}
                            </div>
                            <div className="text-sm text-gray-400">
                                {player === (gameState.landlord || gameState.Landlord) && '🌟 地主'}
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default DDZGame;