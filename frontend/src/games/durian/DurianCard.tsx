import React, { useRef, useEffect } from 'react';
import { Box } from '@mui/material';
import type { CardData, FruitType } from './types';
import { layoutFruits } from './FruitCanvas';

interface DurianCardProps {
  card: CardData;
  size?: 'small' | 'medium' | 'large';
  chosenSide?: 'left' | 'right' | null;
  flipped?: boolean;
  onChooseLeft?: () => void;
  onChooseRight?: () => void;
  clickable?: boolean;
}

const sizes = {
  small: { w: 90, h: 60 },
  medium: { w: 130, h: 85 },
  large: { w: 170, h: 110 },
};

// 猩猩牌颜色映射
const gorillaColors: Record<string, string> = {
  cancel_banana: '#E53935',  // 红色 - 无限香蕉
  cancel_count3: '#26C6B0',  // 青绿色 - 去掉3
  do_nothing: '#7E57C2',     // 紫色 - 无事发生
};

// 猩猩牌符号
const gorillaSymbols: Record<string, string> = {
  cancel_banana: '∞',
  cancel_count3: '3',
  do_nothing: '0',
};

// 猩猩牌标签
const gorillaLabels: Record<string, string> = {
  cancel_banana: '无限香蕉',
  cancel_count3: '去掉3',
  do_nothing: '无事发生',
};

// 绘制猩猩牌 (简化: 彩色背景 + 大符号 + 标签)
const drawGorillaCard = (
  ctx: CanvasRenderingContext2D,
  w: number, h: number,
  ability: string,
) => {
  const r = 8;
  const bgColor = gorillaColors[ability] || '#666';
  const symbol = gorillaSymbols[ability] || '?';
  const label = gorillaLabels[ability] || '';

  // 彩色圆角背景
  ctx.beginPath();
  ctx.roundRect(0, 0, w, h, r);
  ctx.fillStyle = bgColor;
  ctx.fill();

  // 半透明深色底部
  ctx.beginPath();
  ctx.roundRect(0, h * 0.6, w, h * 0.4, [0, 0, r, r]);
  ctx.fillStyle = 'rgba(0,0,0,0.3)';
  ctx.fill();

  // 大符号 (居中偏上)
  const symbolSize = Math.min(w, h) * 0.4;
  ctx.font = `bold ${symbolSize}px Arial`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillStyle = '#fff';
  ctx.fillText(symbol, w / 2, h * 0.35);

  // 🦍 小图标
  const emojiSize = Math.min(w, h) * 0.15;
  ctx.font = `${emojiSize}px Arial`;
  ctx.fillText('🦍', w / 2, h * 0.68);

  // 标签文字
  if (label) {
    const labelSize = Math.min(w, h) * 0.12;
    ctx.font = `bold ${labelSize}px Arial`;
    ctx.fillStyle = '#fff';
    ctx.fillText(label, w / 2, h * 0.87);
  }
};

// 绘制水果牌 (黑色背景 + 左右水果)
const drawFruitCard = (
  ctx: CanvasRenderingContext2D,
  w: number, h: number,
  leftFruit: FruitType, leftCount: number,
  rightFruit: FruitType, rightCount: number,
  chosenSide: 'left' | 'right' | null,
) => {
  const halfW = w / 2;
  const r = 8;

  // 黑色圆角背景
  ctx.beginPath();
  ctx.roundRect(0, 0, w, h, r);
  ctx.fillStyle = '#1a1a1a';
  ctx.fill();

  // 中间分割线
  ctx.beginPath();
  ctx.moveTo(halfW, h * 0.08);
  ctx.lineTo(halfW, h * 0.92);
  ctx.strokeStyle = '#444';
  ctx.lineWidth = 1.5;
  ctx.stroke();

  // 左半水果
  layoutFruits(ctx, leftFruit, leftCount, 4, 4, halfW - 8, h - 8);
  // 右半水果
  layoutFruits(ctx, rightFruit, rightCount, halfW + 4, 4, halfW - 8, h - 8);

  // 选中高亮
  if (chosenSide === 'left') {
    ctx.beginPath();
    ctx.roundRect(2, 2, halfW - 3, h - 4, [r, 0, 0, r]);
    ctx.strokeStyle = '#FFD54F';
    ctx.lineWidth = 3;
    ctx.stroke();
  } else if (chosenSide === 'right') {
    ctx.beginPath();
    ctx.roundRect(halfW + 1, 2, halfW - 3, h - 4, [0, r, r, 0]);
    ctx.strokeStyle = '#FFD54F';
    ctx.lineWidth = 3;
    ctx.stroke();
  }
};

export const DurianCard: React.FC<DurianCardProps> = ({
  card, size = 'medium', chosenSide = null, flipped = false,
  onChooseLeft, onChooseRight, clickable = false,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const sz = sizes[size];
  const dpr = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    canvas.width = sz.w * dpr;
    canvas.height = sz.h * dpr;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, sz.w, sz.h);

    if (card.card_type === 'gorilla') {
      drawGorillaCard(ctx, sz.w, sz.h, card.ability);
    } else {
      // If flipped, swap left/right fruits so chosen side always appears on right
      const lf = flipped ? card.right_fruit as FruitType : card.left_fruit as FruitType;
      const lc = flipped ? card.right_count : card.left_count;
      const rf = flipped ? card.left_fruit as FruitType : card.right_fruit as FruitType;
      const rc = flipped ? card.left_count : card.right_count;
      // Always highlight right side when flipped (chosen was left, now shown on right)
      drawFruitCard(ctx, sz.w, sz.h, lf, lc, rf, rc, chosenSide);
    }
  }, [card, sz, chosenSide, flipped, dpr]);

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!clickable || card.card_type !== 'fruit') return;
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    if (x < sz.w / 2) onChooseLeft?.();
    else onChooseRight?.();
  };

  return (
    <Box sx={{ display: 'inline-block', cursor: clickable ? 'pointer' : 'default' }}>
      <canvas
        ref={canvasRef}
        onClick={handleClick}
        style={{ width: sz.w, height: sz.h, display: 'block', borderRadius: 8 }}
      />
    </Box>
  );
};
