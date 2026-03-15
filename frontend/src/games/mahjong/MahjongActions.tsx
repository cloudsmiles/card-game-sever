import { Box, Button, Typography, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { useState } from 'react';
import { motion } from 'framer-motion';
import { MahjongTileComp } from './MahjongTile';
import type { MahjongTile, KongOption } from './types';

interface MahjongActionsProps {
  availableActions: string[];
  onDiscard: () => void;
  onWin: () => void;
  onKong: (tile: MahjongTile) => void;
  onPong: () => void;
  onChow: (tiles: MahjongTile[]) => void;
  onPass: () => void;
  hasSelection: boolean;
  kongOptions?: KongOption[];
  chowOptions?: MahjongTile[][];
}

export const MahjongActions: React.FC<MahjongActionsProps> = ({
  availableActions,
  onDiscard,
  onWin,
  onKong,
  onPong,
  onChow,
  onPass,
  hasSelection,
  kongOptions = [],
  chowOptions = [],
}) => {
  const [showKongDialog, setShowKongDialog] = useState(false);
  const [showChowDialog, setShowChowDialog] = useState(false);

  const has = (action: string) => availableActions.includes(action);

  const handleKongClick = () => {
    if (kongOptions.length === 1) {
      onKong(kongOptions[0].tile);
    } else if (kongOptions.length > 1) {
      setShowKongDialog(true);
    }
  };

  const handleChowClick = () => {
    if (chowOptions.length === 1) {
      onChow(chowOptions[0]);
    } else if (chowOptions.length > 1) {
      setShowChowDialog(true);
    }
  };

  if (availableActions.length === 0) {
    return null;
  }

  return (
    <>
      <Box sx={{ display: 'flex', justifyContent: 'center', gap: 1, flexWrap: 'wrap', py: 1 }}>
        {/* Play phase actions */}
        {has('discard') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="contained"
            size="small"
            onClick={onDiscard}
            disabled={!hasSelection}
          >
            出牌
          </Button>
        )}
        {has('win') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="contained"
            color="error"
            size="small"
            onClick={onWin}
          >
            胡
          </Button>
        )}
        {has('kong') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="contained"
            color="secondary"
            size="small"
            onClick={handleKongClick}
          >
            杠
          </Button>
        )}
        {/* Pending phase actions */}
        {has('pong') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="contained"
            color="secondary"
            size="small"
            onClick={onPong}
          >
            碰
          </Button>
        )}
        {has('chow') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="contained"
            color="secondary"
            size="small"
            onClick={handleChowClick}
          >
            吃
          </Button>
        )}
        {has('pass') && (
          <Button
            component={motion.button}
            whileTap={{ scale: 0.95 }}
            variant="outlined"
            size="small"
            onClick={onPass}
          >
            过
          </Button>
        )}
      </Box>

      {/* Kong selection dialog */}
      <Dialog open={showKongDialog} onClose={() => setShowKongDialog(false)}>
        <DialogTitle>选择杠牌</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', gap: 2, p: 1 }}>
            {kongOptions.map((opt, i) => (
              <Box
                key={i}
                onClick={() => { onKong(opt.tile); setShowKongDialog(false); }}
                sx={{ cursor: 'pointer', textAlign: 'center' }}
              >
                <MahjongTileComp tile={opt.tile} />
                <Typography variant="caption">
                  {opt.type === 'concealed' ? '暗杠' : '补杠'}
                </Typography>
              </Box>
            ))}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowKongDialog(false)}>取消</Button>
        </DialogActions>
      </Dialog>

      {/* Chow selection dialog */}
      <Dialog open={showChowDialog} onClose={() => setShowChowDialog(false)}>
        <DialogTitle>选择吃牌组合</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, p: 1 }}>
            {chowOptions.map((opt, i) => (
              <Box
                key={i}
                onClick={() => { onChow(opt); setShowChowDialog(false); }}
                sx={{ display: 'flex', gap: 0.5, cursor: 'pointer', p: 1, borderRadius: 1, '&:hover': { bgcolor: 'action.hover' } }}
              >
                {opt.map((t, j) => (
                  <MahjongTileComp key={j} tile={t} size="small" />
                ))}
              </Box>
            ))}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowChowDialog(false)}>取消</Button>
        </DialogActions>
      </Dialog>
    </>
  );
};
