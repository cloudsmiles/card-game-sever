import React from 'react';
import { Box, Typography } from '@mui/material';
import { FruitCanvas } from './FruitCanvas';
import type { FruitType } from './types';
import { fruitName, fruitColor } from './types';

interface OrderAreaProps {
  orders: Record<string, number>;
}

const fruits: FruitType[] = ['durian', 'banana', 'grape', 'strawberry'];

export const OrderArea: React.FC<OrderAreaProps> = ({ orders }) => {
  const total = fruits.reduce((s, f) => s + (orders[f] || 0), 0);

  return (
    <Box sx={{
      p: 1.5, borderRadius: 2,
      bgcolor: 'rgba(30,30,30,0.8)',
      border: '1px solid #444',
    }}>
      <Typography variant="caption" sx={{ color: '#aaa', mb: 0.5, display: 'block', textAlign: 'center' }}>
        📋 订单区 (共 {total} 个水果)
      </Typography>
      <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center', flexWrap: 'wrap' }}>
        {fruits.map((f) => {
          const count = orders[f] || 0;
          return (
            <Box key={f} sx={{ textAlign: 'center', minWidth: 50 }}>
              {count > 0 ? (
                <FruitCanvas fruit={f} count={Math.min(count, 4)} width={44} height={44} />
              ) : (
                <Box sx={{ width: 44, height: 44, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Typography sx={{ color: '#555', fontSize: 20 }}>-</Typography>
                </Box>
              )}
              <Typography sx={{ fontSize: 14, fontWeight: 'bold', color: count > 0 ? fruitColor[f] : '#555' }}>
                {count}
              </Typography>
              <Typography variant="caption" sx={{ color: '#888', fontSize: 10 }}>
                {fruitName[f]}
              </Typography>
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};
