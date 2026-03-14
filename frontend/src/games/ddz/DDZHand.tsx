import { Box } from '@mui/material';
import { DDZCard } from './DDZCard';

interface DDZHandProps {
  cards: number[];
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
      {cards.map((value, index) => (
        <DDZCard
          key={`${value}-${index}`}
          value={value}
          index={index}
          selected={selectedCards.has(index)}
          onClick={disabled ? undefined : () => onCardClick(index)}
        />
      ))}
    </Box>
  );
};
