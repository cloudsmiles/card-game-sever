import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ConnectPage, LobbyPage, RoomPage } from './pages';
import { useConnectionStore } from './stores/connectionStore';

// Protected route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { status } = useConnectionStore();
  
  if (status !== 'connected') {
    return <Navigate to="/" replace />;
  }
  
  return <>{children}</>;
};

export const router = createBrowserRouter([
  {
    path: '/',
    element: <ConnectPage />,
  },
  {
    path: '/lobby',
    element: (
      <ProtectedRoute>
        <LobbyPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/room/:roomId',
    element: (
      <ProtectedRoute>
        <RoomPage />
      </ProtectedRoute>
    ),
  },
], {
  basename: '/qcard',
});
