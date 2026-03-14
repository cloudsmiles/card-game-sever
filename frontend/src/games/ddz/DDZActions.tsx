import { Box, Button } from '@mui/material';
import { motion } from 'framer-motion';

interface DDZActionsProps {
  onPlay: () => void;
  onPass: () => void;
  onClearSelection: () => void;
  canPlay: boolean;
  canPass: boolean;
  hasSelection: boolean;
  disabled?: boolean;
}

export const DDZActions: React.FC<DDZActionsProps> = ({
  onPlay,
  onPass,
  onClearSelection,
  canPlay,
  canPass,
  hasSelection,
  disabled = false,
}) => {
  return (
    <Box
      component={motion.div}
      initial={{ y: 20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      sx={{
        display: 'flex',
        justifyContent: 'center',
        gap: 2,
        py: 2,
      }}
    >
      <Button
        variant="contained"
        color="primary"
        onClick={onPlay}
        disabled={disabled || !canPlay || !hasSelection}
        sx={{
          minWidth: 100,
          fontSize: '1rem',
          fontWeight: 'bold',
        }}
      >
        出牌
      </Button>

      <Button
        variant="outlined"
        color="secondary"
        onClick={onPass}
        disabled={disabled || !canPass}
        sx={{
          minWidth: 100,
          fontSize: '1rem',
          fontWeight: 'bold',
        }}
      >
        不出
      </Button>

      <Button
        variant="text"
        onClick={onClearSelection}
        disabled={disabled || !hasSelection}
        sx={{
          minWidth: 100,
          fontSize: '1rem',
        }}
      >
        清空选择
      </Button>
    </Box>
  );
};
