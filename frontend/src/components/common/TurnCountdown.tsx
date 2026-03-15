import { useState, useEffect, useRef } from 'react';
import { Box, Typography } from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';

interface TurnCountdownProps {
  deadline: number; // Unix毫秒时间戳
  label?: string;
}

export const TurnCountdown: React.FC<TurnCountdownProps> = ({ deadline, label }) => {
  const [seconds, setSeconds] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval>>();

  useEffect(() => {
    const update = () => {
      const remaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
      setSeconds(remaining);
    };
    update();
    intervalRef.current = setInterval(update, 500);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [deadline]);

  if (seconds <= 0) return null;

  const isUrgent = seconds <= 5;
  const color = isUrgent ? '#C62828' : seconds <= 10 ? '#E65100' : '#FF6B35';

  return (
    <Box sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.5 }}>
      {label && (
        <Typography variant="caption" color="text.secondary">
          {label}
        </Typography>
      )}
      <AnimatePresence mode="wait">
        <Box
          component={motion.div}
          key={seconds}
          initial={{ scale: isUrgent ? 1.3 : 1 }}
          animate={{ scale: 1 }}
          transition={{ duration: 0.2 }}
          sx={{
            minWidth: 28,
            height: 28,
            borderRadius: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            bgcolor: `${color}18`,
            border: `2px solid ${color}`,
          }}
        >
          <Typography
            variant="body2"
            sx={{ fontWeight: 'bold', color, fontSize: 13, lineHeight: 1 }}
          >
            {seconds}
          </Typography>
        </Box>
      </AnimatePresence>
    </Box>
  );
};
