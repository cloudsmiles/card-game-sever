import { Box } from '@mui/material';
import { motion } from 'framer-motion';
import { valueToRank } from './types';

interface DDZCardProps {
  value: number;
  /** Index in hand, used to assign a visual suit */
  index?: number;
  selected?: boolean;
  onClick?: () => void;
  faceDown?: boolean;
  size?: 'small' | 'medium';
}

const cardSizes = {
  small: { width: 36, height: 50, rankSize: '0.6rem', suitSize: '0.9rem' },
  medium: { width: 52, height: 72, rankSize: '0.75rem', suitSize: '1.2rem' },
};

const suits = ['♠', '♥', '♣', '♦'] as const;
const suitColors: Record<string, string> = {
  '♠': '#1A1A2E',
  '♥': '#E63946',
  '♣': '#1A1A2E',
  '♦': '#E63946',
};

/** Deterministically pick a suit symbol based on value + index */
const getSuit = (value: number, index: number): string => {
  if (value >= 16) return ''; // Jokers have no suit
  return suits[(value + index) % 4];
};

export const DDZCard: React.FC<DDZCardProps> = ({
  value,
  index = 0,
  selected = false,
  onClick,
  faceDown = false,
  size = 'medium',
}) => {
  const { width, height, rankSize, suitSize } = cardSizes[size];
  const rank = valueToRank(value);
  const isJoker = value >= 16;
  const suit = getSuit(value, index);
  const color = isJoker
    ? value === 17 ? '#E63946' : '#1A1A2E'
    : suitColors[suit] || '#1A1A2E';

  return (
    <Box
      component={motion.div}
      initial={false}
      animate={{ y: selected ? -16 : 0 }}
      whileHover={onClick ? { y: selected ? -16 : -6 } : undefined}
      transition={{ duration: 0.15 }}
      onClick={onClick}
      sx={{
        width,
        height,
        background: faceDown
          ? 'linear-gradient(145deg, #FF6B35, #E64A1A)'
          : '#FFFFFF',
        borderRadius: '6px',
        border: selected ? '2px solid #FF6B35' : '1.5px solid #ccc',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: onClick ? 'pointer' : 'default',
        userSelect: 'none',
        boxShadow: selected
          ? '0 6px 12px rgba(255, 107, 53, 0.4)'
          : '0 1px 4px rgba(0, 0, 0, 0.12)',
        flexShrink: 0,
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {faceDown ? (
        <Box sx={{ fontSize: '1.2rem' }}>🎴</Box>
      ) : isJoker ? (
        <Box sx={{ textAlign: 'center', color, fontWeight: 'bold', lineHeight: 1.2 }}>
          <Box sx={{ fontSize: suitSize }}>{value === 17 ? '🃏' : '🃏'}</Box>
          <Box sx={{ fontSize: rankSize }}>{rank}</Box>
        </Box>
      ) : (
        <>
          {/* Top-left rank + suit */}
          <Box
            sx={{
              position: 'absolute',
              top: 2,
              left: 3,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              lineHeight: 1,
              color,
            }}
          >
            <Box sx={{ fontSize: rankSize, fontWeight: 'bold' }}>{rank}</Box>
            <Box sx={{ fontSize: `calc(${rankSize} * 0.9)` }}>{suit}</Box>
          </Box>
          {/* Center suit large */}
          <Box sx={{ fontSize: suitSize, color, mt: 0.5 }}>
            {suit}
          </Box>
        </>
      )}
    </Box>
  );
};
