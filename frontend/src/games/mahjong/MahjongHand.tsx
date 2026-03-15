import { Box } from '@mui/material';
import { MahjongTileComp } from './MahjongTile';
import type { MahjongTile } from './types';

interface MahjongHandProps {
  tiles: MahjongTile[];
  selectedIndices: Set<number>;
  onTileClick?: (index: number) => void;
  disabled?: boolean;
}

export const MahjongHand: React.FC<MahjongHandProps> = ({
  tiles,
  selectedIndices,
  onTileClick,
  disabled = false,
}) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexWrap: 'wrap',
        justifyContent: 'center',
        gap: 0.3,
        py: 1,
      }}
    >
      {tiles.map((tile, i) => (
        <MahjongTileComp
          key={i}
          tile={tile}
          selected={selectedIndices.has(i)}
          onClick={disabled ? undefined : () => onTileClick?.(i)}
        />
      ))}
    </Box>
  );
};
