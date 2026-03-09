import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import LobbyPage from './pages/LobbyPage';
import GameRoom from './pages/GameRoom';

function App() {
    return (
        <Router>
            <AppProvider>
                <div className="App">
                    <Routes>
                        <Route path="/" element={<LobbyPage />} />
                        <Route path="/room/:roomId" element={<GameRoom />} />
                    </Routes>
                </div>
            </AppProvider>
        </Router>
    );
}

export default App;