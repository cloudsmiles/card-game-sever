import React from 'react';
import { Box, Typography, Chip } from '@mui/material';
import { DurianCard } from './DurianCard';
import { FruitCanvas } from './FruitCanvas';
import { AngerToken } from './AngerToken';
import type { SettlementData, FruitType } from './types';
import { fruitName } from './types';

interface SettlementViewProps {
  settlement: SettlementData;
  getName: (pid: string) => string;
}

const fruits: FruitType[] = ['durian', 'banana', 'grape', 'strawberry'];

export const SettlementView: React.FC<SettlementViewProps> = ({ settlement, getName }) => {
  const s = settlement;
  return (
    <Box sx={{ p: 2, borderRadius: 2, bgcolor: 'rgba(40,30,20,0.9)', border: '1px solid #FFD54F' }}>
      <Typography variant="body2" sx={{ fontWeight: 'bold', mb: 1, color: '#FFD54F', textAlign: 'center' }}>
        🔔 结算结果 (摇铃者: {getName(s.bell_ringer)})
      </Typography>
      <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 1.5, justifyContent: 'center' }}>
        {s.all_holder_cards.map((hc) => (
          <Box key={hc.player_id} sx={{ textAlign: 'center' }}>
            <DurianCard card={hc.card} size="small" />
            <Typography variant="caption" sx={{ display: 'block', color: '#ccc', mt: 0.3 }}>
              {getName(hc.player_id)}
            </Typography>
          </Box>
        ))}
      </Box>
      {s.gorilla_effects.length > 0 && (
        <Box sx={{ mb: 1, textAlign: 'center' }}>
          {s.gorilla_effects.map((e, i) => (
            <Chip key={i} size="small" label={`🦍 ${getName(e.player_id)}: ${e.ability}`}
              sx={{ mr: 0.5, mb: 0.5, bgcolor: '#333', color: '#ccc' }} />
          ))}
        </Box>
      )}
      <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center', mb: 1.5, flexWrap: 'wrap' }}>
        {fruits.map((f) => {
          const order = s.orders_after_cancel[f] || 0;
          const inv = s.inventory[f] || 0;
          return (
            <Box key={f} sx={{ textAlign: 'center' }}>
              <FruitCanvas fruit={f} count={1} width={30} height={30} />
              <Typography variant="caption" sx={{
                color: order > inv ? '#E53935' : '#4CAF50', fontWeight: 'bold', display: 'block',
              }}>需{order} / 有{inv >= 9999 ? '∞' : inv}</Typography>
            </Box>
          );
        })}
      </Box>
      <Box sx={{
        p: 1, borderRadius: 1, mb: 1, textAlign: 'center',
        bgcolor: s.is_shortage ? 'rgba(229,57,53,0.15)' : 'rgba(76,175,80,0.15)',
        border: s.is_shortage ? '1px solid #E53935' : '1px solid #4CAF50',
      }}>
        <Typography variant="body2" sx={{ color: s.is_shortage ? '#EF9A9A' : '#A5D6A7' }}>
          {s.is_shortage
            ? `缺货！缺少: ${s.shortage_fruits.map((f) => fruitName[f as FruitType] || f).join(', ')}`
            : '库存充足，摇铃者判断失误'}
        </Typography>
      </Box>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
        <Typography variant="body2" sx={{ color: '#ccc' }}>
          受罚者: <span style={{ color: '#FFD54F', fontWeight: 'bold' }}>{getName(s.punished_player)}</span>
          {s.punish_reason === 'last_order' ? '（最后接订单者）' : '（摇铃者）'}
        </Typography>
        <AngerToken value={s.anger_token_given} size={36} />
      </Box>
    </Box>
  );
};
