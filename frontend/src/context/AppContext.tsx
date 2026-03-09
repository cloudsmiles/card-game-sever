import React, { createContext, useContext, useState, useEffect } from 'react';
import { wsService } from '../services/WebSocketService';

interface AppContextType {
    playerID: string;
    setPlayerID: (id: string) => void;
    isConnected: boolean;
    currentRoomID: string;
    setCurrentRoomID: (id: string) => void;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export const useApp = () => {
    const context = useContext(AppContext);
    if (!context) {
        throw new Error('useApp must be used within AppProvider');
    }
    return context;
};

interface AppProviderProps {
    children: React.ReactNode;
}

export const AppProvider: React.FC<AppProviderProps> = ({ children }) => {
    const [playerID, setPlayerID] = useState('');
    const [isConnected, setIsConnected] = useState(false);
    const [currentRoomID, setCurrentRoomID] = useState('');

    // 监听WebSocket连接状态
    useEffect(() => {
        const handleConnect = () => {
            setIsConnected(true);
            console.log('WebSocket已连接');
        };

        const handleDisconnect = () => {
            setIsConnected(false);
            console.log('WebSocket已断开');
        };

        const handleError = (error: any) => {
            console.error('WebSocket错误:', error);
        };

        wsService.on('message', handleConnect);
        wsService.on('disconnected', handleDisconnect);
        wsService.on('error', handleError);

        return () => {
            wsService.off('message', handleConnect);
            wsService.off('disconnected', handleDisconnect);
            wsService.off('error', handleError);
        };
    }, []);

    const value = {
        playerID,
        setPlayerID,
        isConnected,
        currentRoomID,
        setCurrentRoomID
    };

    return (
        <AppContext.Provider value={value}>
            {children}
        </AppContext.Provider>
    );
};