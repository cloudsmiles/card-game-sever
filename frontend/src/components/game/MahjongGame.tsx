import React from 'react';
import { wsService } from '../../services/WebSocketService';

interface MahjongGameProps {
    gameState: any;
    playerID: string;
    onLog: (message: string) => void;
}

// 牌值转显示字符
const tileToStr = (tile: number): string => {
    // 万子 1-9
    if (tile >= 1 && tile <= 9) return `${tile}万`;
    // 筒子 11-19
    if (tile >= 11 && tile <= 19) return `${tile - 10}筒`;
    // 索子 21-29
    if (tile >= 21 && tile <= 29) return `${tile - 20}索`;
    // 字牌 31-37
    const ziPai = ['', '东', '南', '西', '北', '中', '发', '白'];
    const ziIndex = tile - 30;
    if (ziIndex >= 1 && ziIndex <= 7) return ziPai[ziIndex];
    return '?';
};

// 获取颜色类
const getTileColor = (tile: number): string => {
    if (tile >= 1 && tile <= 9) return 'bg-green-600';
    if (tile >= 11 && tile <= 19) return 'bg-red-600';
    if (tile >= 21 && tile <= 29) return 'bg-blue-600';
    return 'bg-gray-600';
};

const MahjongGame: React.FC<MahjongGameProps> = ({ gameState, playerID, onLog }) => {
    // 打牌
    const discardTile = (tile: number) => {
        const currentTurn = gameState?.current_turn || gameState?.current || '';
        if (currentTurn !== playerID) {
            onLog('不是你的回合');
            return;
        }
        wsService.gameAction('discard', { tile });
        onLog(`打牌: ${tileToStr(tile)}`);
    };

    // 吃牌
    const chowTile = (tiles: number[]) => {
        wsService.gameAction('chow', { tiles });
        onLog(`吃牌: ${tiles.map(tileToStr).join('')}`);
    };

    // 碰牌
    const pongTile = (tile: number) => {
        wsService.gameAction('pong', { tile });
        onLog(`碰牌: ${tileToStr(tile)}`);
    };

    // 杠牌
    const kongTile = (tile: number) => {
        wsService.gameAction('kong', { tile });
        onLog(`杠牌: ${tileToStr(tile)}`);
    };

    // 和牌
    const huTile = () => {
        wsService.gameAction('hu', {});
        onLog('和牌！');
    };

    const currentTurn = gameState?.current_turn || gameState?.current || '';
    const isMyTurn = currentTurn === playerID;
    const myHand = gameState?.my_hand || gameState?.hands?.[playerID] || [];
    const lastDraw = gameState?.last_draw || gameState?.lastDraw || null;
    const discards = gameState?.discards || {};
    const melds = gameState?.melds || {};
    const myMelds = melds[playerID] || [];
    const myDiscards = discards[playerID] || [];

    // 等待操作
    const pendingAction = gameState?.pending_action || gameState?.pendingAction || null;

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
                    <span className="text-gray-400">剩余牌：</span>
                    <span className="font-bold text-xl">{gameState?.remaining_tiles || gameState?.remainingTiles || 0}</span>
                </div>
                <div>
                    <span className="text-gray-400">阶段：</span>
                    <span className="font-bold text-xl">{gameState?.phase || '打牌'}</span>
                </div>
            </div>

            {/* 摸到的牌 */}
            {lastDraw && isMyTurn && (
                <div className="mb-4 p-4 bg-green-900/50 rounded-xl">
                    <div className="text-sm text-gray-400 mb-2">摸到的牌</div>
                    <button
                        onClick={() => discardTile(lastDraw)}
                        className={`px-4 py-3 ${getTileColor(lastDraw)} rounded-lg font-bold text-lg hover:scale-105 transition-transform`}
                    >
                        {tileToStr(lastDraw)}
                    </button>
                </div>
            )}

            {/* 手牌 */}
            <div className="mb-6">
                <div className="text-lg font-bold mb-3">你的手牌 ({myHand.length}张)</div>
                <div className="flex flex-wrap gap-2">
                    {[...myHand].sort((a, b) => a - b).map((tile, index) => (
                        <button
                            key={`${tile}-${index}`}
                            onClick={() => isMyTurn && discardTile(tile)}
                            disabled={!isMyTurn}
                            className={`px-3 py-4 ${getTileColor(tile)} rounded-lg font-bold transition-all ${isMyTurn ? 'hover:scale-110 hover:-translate-y-1 cursor-pointer' : 'opacity-70 cursor-not-allowed'}`}
                        >
                            {tileToStr(tile)}
                        </button>
                    ))}
                </div>
            </div>

            {/* 副露（吃碰杠） */}
            {myMelds.length > 0 && (
                <div className="mb-6">
                    <div className="text-sm text-gray-400 mb-2">副露</div>
                    <div className="flex flex-wrap gap-2">
                        {myMelds.map((meld: any, index: number) => (
                            <div key={index} className="flex gap-1">
                                {(meld.tiles || []).map((tile: number, i: number) => (
                                    <div key={i} className={`px-2 py-1 ${getTileColor(tile)} rounded text-sm font-bold`}>
                                        {tileToStr(tile)}
                                    </div>
                                ))}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* 打出的牌 */}
            {myDiscards.length > 0 && (
                <div className="mb-6">
                    <div className="text-sm text-gray-400 mb-2">打出的牌</div>
                    <div className="flex flex-wrap gap-1">
                        {myDiscards.map((tile: number, index: number) => (
                            <div key={index} className={`px-2 py-1 ${getTileColor(tile)} rounded text-sm opacity-60`}>
                                {tileToStr(tile)}
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* 等待操作按钮 */}
            {pendingAction && pendingAction !== 'none' && pendingAction !== 'discard' && isMyTurn && (
                <div className="flex gap-4 mb-4">
                    {pendingAction === 'hu' && (
                        <button onClick={huTile} className="bg-yellow-600 hover:bg-yellow-500 px-6 py-2 rounded-xl font-bold">
                            和牌
                        </button>
                    )}
                    {pendingAction === 'pong' && gameState.pending_tile && (
                        <button onClick={() => pongTile(gameState.pending_tile)} className="bg-blue-600 hover:bg-blue-500 px-6 py-2 rounded-xl">
                            碰
                        </button>
                    )}
                    {pendingAction === 'kong' && gameState.pending_tile && (
                        <button onClick={() => kongTile(gameState.pending_tile)} className="bg-purple-600 hover:bg-purple-500 px-6 py-2 rounded-xl">
                            杠
                        </button>
                    )}
                    {pendingAction === 'chow' && gameState.pending_tile && (
                        <button onClick={() => chowTile([gameState.pending_tile])} className="bg-green-600 hover:bg-green-500 px-6 py-2 rounded-xl">
                            吃
                        </button>
                    )}
                </div>
            )}

            {/* 玩家信息 */}
            <div className="mt-6">
                <div className="text-sm text-gray-400 mb-2">玩家状态</div>
                <div className="grid grid-cols-4 gap-4">
                    {(gameState.players || gameState.Players || []).map((player: string, index: number) => (
                        <div key={index} className={`p-3 rounded-xl text-center ${player === playerID ? 'bg-blue-500/20 border border-blue-500' : 'bg-gray-700'}`}>
                            <div className="font-medium">
                                {player}
                                {player === playerID && ' (你)'}
                            </div>
                            <div className="text-sm text-gray-400">
                                {discards[player]?.length || 0} 张
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default MahjongGame;