import React, { useState } from 'react';
import { wsService } from '../../services/WebSocketService';

// 字符串转卡牌值
const strToValue = (s: string): number => {
    if (s >= '3' && s <= '10') return parseInt(s);
    if (s === 'J') return 11;
    if (s === 'Q') return 12;
    if (s === 'K') return 13;
    if (s === 'A') return 14;
    if (s === '2') return 15;
    if (s === '小王') return 16;
    if (s === '大王') return 17;
    return 0;
};

interface SelectedCard {
    id: string;
    value: number;
    index: number;
}

interface SimpleGameProps {
    gameState: any;
    playerID: string;
    onLog: (message: string) => void;
}

const SimpleGame: React.FC<SimpleGameProps> = ({ gameState, playerID, onLog }) => {
    const [selectedCards, setSelectedCards] = useState<SelectedCard[]>([]);

    // 选择卡牌
    const selectCard = (card: string, index: number) => {
        const cardId = `${card}_${index}`;
        const selectedIndex = selectedCards.findIndex(c => c.id === cardId);

        if (selectedIndex > -1) {
            setSelectedCards(prev => prev.filter(c => c.id !== cardId));
        } else {
            const value = strToValue(card);
            setSelectedCards(prev => [...prev, { id: cardId, value, index }]);
        }
    };

    // 出牌
    const playSelected = () => {
        if (selectedCards.length === 0) {
            alert('请选择至少一张牌');
            return;
        }
        wsService.gameAction('play_selected', { cards: selectedCards.map(c => c.id) });
        onLog(`出牌: ${selectedCards.map(c => c.id).join(', ')}`);
        setSelectedCards([]);
    };

    const currentTurn = gameState?.current_turn || gameState?.current || '';
    const isMyTurn = currentTurn === playerID;
    const hands = gameState?.hands || {};
    const myHand = hands[playerID] || [];

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
                    {myHand.map((card: string, index: number) => (
                        <button
                            key={`${card}-${index}`}
                            onClick={() => isMyTurn && selectCard(card, index)}
                            className={`px-4 py-3 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl font-bold text-lg transition-all transform hover:scale-110 hover:shadow-lg ${isMyTurn
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
                            onClick={playSelected}
                            disabled={selectedCards.length === 0}
                            className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed px-4 py-2 rounded-lg font-medium transition-colors"
                        >
                            出牌 ({selectedCards.length})
                        </button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        {myHand.map((card: string, index: number) => {
                            const cardId = `${card}_${index}`;
                            const isSelected = selectedCards.some(c => c.id === cardId);
                            return (
                                <button
                                    key={`select-${card}-${index}`}
                                    onClick={() => selectCard(card, index)}
                                    className={`px-4 py-3 rounded-xl font-bold text-lg transition-all border-2 ${isSelected
                                            ? 'bg-blue-600 border-blue-400 scale-105'
                                            : 'bg-gray-700 border-gray-600 hover:bg-gray-600'
                                        }`}
                                >
                                    {card}
                                </button>
                            );
                        })}
                    </div>
                </div>
            )}

            {/* 玩家信息 */}
            <div className="mt-6">
                <div className="text-sm text-gray-400 mb-2">玩家状态</div>
                <div className="grid grid-cols-2 gap-4">
                    {Object.keys(hands).map((player) => (
                        <div key={player} className={`p-3 rounded-xl ${player === playerID ? 'bg-blue-500/20 border border-blue-500' : 'bg-gray-700'}`}>
                            <div className="font-medium">
                                {player}
                                {player === playerID && ' (你)'}
                            </div>
                            <div className="text-sm text-gray-400">
                                {hands[player]?.length || 0} 张牌
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default SimpleGame;