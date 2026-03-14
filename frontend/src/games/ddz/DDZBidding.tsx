import { Box, Button, Typography } from '@mui/material';
import { motion } from 'framer-motion';

interface DDZBiddingProps {
  isMyTurn: boolean;
  onBid: (score: number) => void;
}

export const DDZBidding: React.FC<DDZBiddingProps> = ({ isMyTurn, onBid }) => {
  return (
    <Box
      component={motion.div}
      initial={{ y: 20, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 2,
        py: 3,
      }}
    >
      <Typography variant="h6" fontWeight="bold">
        {isMyTurn ? '请叫地主' : '等待其他玩家叫地主...'}
      </Typography>

      {isMyTurn && (
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', justifyContent: 'center' }}>
          <Button
            variant="outlined"
            color="secondary"
            onClick={() => onBid(0)}
            sx={{ minWidth: 80 }}
          >
            不叫
          </Button>
          <Button
            variant="contained"
            onClick={() => onBid(1)}
            sx={{ minWidth: 80 }}
          >
            1分
          </Button>
          <Button
            variant="contained"
            onClick={() => onBid(2)}
            sx={{ minWidth: 80 }}
          >
            2分
          </Button>
          <Button
            variant="contained"
            color="error"
            onClick={() => onBid(3)}
            sx={{ minWidth: 80 }}
          >
            3分
          </Button>
        </Box>
      )}
    </Box>
  );
};
