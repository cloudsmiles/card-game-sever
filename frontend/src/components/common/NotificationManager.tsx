import { useEffect } from 'react';
import { Snackbar, Alert } from '@mui/material';
import { useUIStore } from '@/stores/uiStore';

export const NotificationManager: React.FC = () => {
  const { notifications, removeNotification } = useUIStore();

  useEffect(() => {
    // Auto-dismiss notifications after 5 seconds
    notifications.forEach((notification) => {
      const duration = notification.duration || 5000;
      const timer = setTimeout(() => {
        removeNotification(notification.id);
      }, duration);

      return () => clearTimeout(timer);
    });
  }, [notifications, removeNotification]);

  return (
    <>
      {notifications.map((notification, index) => (
        <Snackbar
          key={notification.id}
          open={true}
          anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
          sx={{ mt: index * 7 }}
        >
          <Alert
            severity={notification.type}
            onClose={() => removeNotification(notification.id)}
            variant="filled"
            sx={{ width: '100%' }}
          >
            {notification.message}
          </Alert>
        </Snackbar>
      ))}
    </>
  );
};
