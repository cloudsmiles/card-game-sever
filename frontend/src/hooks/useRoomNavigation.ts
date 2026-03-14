import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useRoomStore } from '@/stores/roomStore';

/**
 * Hook to handle automatic navigation when joining/leaving rooms
 */
export const useRoomNavigation = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { currentRoomId } = useRoomStore();

  useEffect(() => {
    // If we have a room ID and we're not already on the room page, navigate to it
    if (currentRoomId && !location.pathname.includes(`/room/${currentRoomId}`)) {
      navigate(`/room/${currentRoomId}`);
    }
    
    // If we don't have a room ID and we're on a room page, navigate to lobby
    if (!currentRoomId && location.pathname.startsWith('/room/')) {
      navigate('/lobby');
    }
  }, [currentRoomId, location.pathname, navigate]);
};
