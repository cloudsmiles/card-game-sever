import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import MagicalLobbyPage from './pages/MagicalLobbyPage';
import GameRoom from './pages/GameRoom';

function App() {
    return (
        <Router>
            <AppProvider>
                <div className="App">
                    <Routes>
                        <Route path="/" element={<MagicalLobbyPage />} />
                        <Route path="/room/:roomId" element={<GameRoom />} />
                    </Routes>
                </div>
            </AppProvider>
        </Router>
    );
}

export default App;