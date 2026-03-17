import { Box } from '@mui/material';
import { DDZCard } from './DDZCard';
import type { CardObj } from './types';

interface DDZHandProps {
  cards: CardObj[];
  selectedCards: Set<number>; // indices
  onCardClick: (index: number) => void;
  disabled?: boolean;
}

export const DDZHand: React.FC<DDZHandProps> = ({
  cards,
  selectedCards,
  onCardClick,
  disabled = false,
}) => {
  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        flexWrap: 'wrap',
        gap: 0.5,
        py: 1,
      }}
    >
      {cards.map((card, index) => (
        <DDZCard
          key={`${card.Value}-${card.Suit}-${index}`}
          value={card.Value}
          suit={card.Suit}
          index={index}
          selected={selectedCards.has(index)}
          onClick={disabled ? undefined : () => onCardClick(index)}
        />
      ))}
    </Box>
  );
};
