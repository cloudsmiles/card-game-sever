import React, { useRef, useEffect } from 'react';
import type { FruitType } from './types';

interface FruitCanvasProps {
  fruit: FruitType;
  count: number;
  width?: number;
  height?: number;
}

// 草莓: 粉红色圆润水滴形 + 小叶子
const drawStrawberry = (ctx: CanvasRenderingContext2D, cx: number, cy: number, s: number) => {
  const r = s * 0.38;
  ctx.beginPath();
  ctx.moveTo(cx, cy + r * 1.1);
  ctx.quadraticCurveTo(cx - r * 1.2, cy + r * 0.3, cx - r * 0.8, cy - r * 0.3);
  ctx.quadraticCurveTo(cx - r * 0.3, cy - r * 1.0, cx, cy - r * 0.7);
  ctx.quadraticCurveTo(cx + r * 0.3, cy - r * 1.0, cx + r * 0.8, cy - r * 0.3);
  ctx.quadraticCurveTo(cx + r * 1.2, cy + r * 0.3, cx, cy + r * 1.1);
  const grad = ctx.createRadialGradient(cx - r * 0.2, cy - r * 0.2, 0, cx, cy, r * 1.2);
  grad.addColorStop(0, '#FF7EB3');
  grad.addColorStop(0.7, '#FF5C8A');
  grad.addColorStop(1, '#E8456B');
  ctx.fillStyle = grad;
  ctx.fill();
  ctx.beginPath();
  ctx.ellipse(cx - r * 0.15, cy - r * 0.75, r * 0.2, r * 0.1, -0.4, 0, Math.PI * 2);
  ctx.fillStyle = '#4CAF50';
  ctx.fill();
  ctx.beginPath();
  ctx.ellipse(cx + r * 0.15, cy - r * 0.75, r * 0.2, r * 0.1, 0.4, 0, Math.PI * 2);
  ctx.fillStyle = '#4CAF50';
  ctx.fill();
};

// 香蕉: 黄色月牙形 (类似🍌 emoji)
const drawBanana = (ctx: CanvasRenderingContext2D, cx: number, cy: number, s: number) => {
  ctx.save();
  ctx.translate(cx, cy);
  ctx.rotate(0.15);
  const h = s * 0.75, w = s * 0.3;
  // 香蕉主体 - 弯曲的月牙
  ctx.beginPath();
  ctx.moveTo(0, -h * 0.45);
  // 左侧外弧 (凸出)
  ctx.bezierCurveTo(-w * 1.8, -h * 0.15, -w * 1.6, h * 0.35, -w * 0.1, h * 0.45);
  // 底部尖端
  ctx.bezierCurveTo(w * 0.2, h * 0.42, w * 0.3, h * 0.35, w * 0.3, h * 0.25);
  // 右侧内弧 (凹进)
  ctx.bezierCurveTo(w * 0.5, -h * 0.05, w * 0.3, -h * 0.3, w * 0.1, -h * 0.42);
  ctx.closePath();
  const grad = ctx.createLinearGradient(-w, 0, w * 0.8, 0);
  grad.addColorStop(0, '#FFD54F');
  grad.addColorStop(0.5, '#FFCA28');
  grad.addColorStop(1, '#F9A825');
  ctx.fillStyle = grad;
  ctx.fill();
  // 茎
  ctx.beginPath();
  ctx.moveTo(w * 0.05, -h * 0.44);
  ctx.lineTo(w * 0.15, -h * 0.52);
  ctx.strokeStyle = '#8D6E63';
  ctx.lineWidth = s * 0.04;
  ctx.lineCap = 'round';
  ctx.stroke();
  ctx.restore();
};

// 榴莲: 绿色锯齿圆形
const drawDurian = (ctx: CanvasRenderingContext2D, cx: number, cy: number, s: number) => {
  const r = s * 0.38;
  const spikes = 14, innerR = r * 0.78, outerR = r * 1.05;
  ctx.beginPath();
  for (let i = 0; i <= spikes; i++) {
    const angle = (i / spikes) * Math.PI * 2 - Math.PI / 2;
    const nextAngle = ((i + 0.5) / spikes) * Math.PI * 2 - Math.PI / 2;
    const ox = cx + Math.cos(angle) * outerR;
    const oy = cy + Math.sin(angle) * outerR;
    const ix = cx + Math.cos(nextAngle) * innerR;
    const iy = cy + Math.sin(nextAngle) * innerR;
    if (i === 0) ctx.moveTo(ox, oy); else ctx.lineTo(ox, oy);
    ctx.lineTo(ix, iy);
  }
  ctx.closePath();
  const grad = ctx.createRadialGradient(cx - r * 0.3, cy - r * 0.3, 0, cx, cy, outerR);
  grad.addColorStop(0, '#66DDAA');
  grad.addColorStop(0.5, '#4DB89A');
  grad.addColorStop(1, '#2E8B6E');
  ctx.fillStyle = grad;
  ctx.fill();
};

// 葡萄: 紫色倒三角串 (类似🍇 emoji)
const drawGrape = (ctx: CanvasRenderingContext2D, cx: number, cy: number, s: number) => {
  const r = s * 0.13;
  // 倒三角形排列: 3顶 + 2中 + 1底, 紧密贴合
  const dy = r * 1.7; // 行间距
  const dx = r * 1.1; // 列间距
  const positions = [
    [-dx, -dy * 0.7], [0, -dy * 0.7], [dx, -dy * 0.7],
    [-dx * 0.5, dy * 0.05], [dx * 0.5, dy * 0.05],
    [0, dy * 0.8],
  ];
  for (const [ox, oy] of positions) {
    const grad = ctx.createRadialGradient(cx + ox - r * 0.3, cy + oy - r * 0.3, 0, cx + ox, cy + oy, r);
    grad.addColorStop(0, '#CE93D8');
    grad.addColorStop(0.6, '#9C27B0');
    grad.addColorStop(1, '#6A1B9A');
    ctx.beginPath();
    ctx.arc(cx + ox, cy + oy, r, 0, Math.PI * 2);
    ctx.fillStyle = grad;
    ctx.fill();
  }
  // 茎
  ctx.beginPath();
  ctx.moveTo(cx, cy - dy * 0.7 - r);
  ctx.lineTo(cx, cy - dy * 0.7 - r - s * 0.12);
  ctx.strokeStyle = '#5D4037';
  ctx.lineWidth = s * 0.035;
  ctx.lineCap = 'round';
  ctx.stroke();
  // 叶子
  ctx.beginPath();
  ctx.ellipse(cx + s * 0.05, cy - dy * 0.7 - r - s * 0.08, s * 0.07, s * 0.03, 0.4, 0, Math.PI * 2);
  ctx.fillStyle = '#4CAF50';
  ctx.fill();
};

const drawFruit = (ctx: CanvasRenderingContext2D, fruit: FruitType, cx: number, cy: number, size: number) => {
  ctx.save();
  switch (fruit) {
    case 'durian': drawDurian(ctx, cx, cy, size); break;
    case 'banana': drawBanana(ctx, cx, cy, size); break;
    case 'grape': drawGrape(ctx, cx, cy, size); break;
    case 'strawberry': drawStrawberry(ctx, cx, cy, size); break;
  }
  ctx.restore();
};

const layoutFruits = (
  ctx: CanvasRenderingContext2D, fruit: FruitType, count: number,
  x: number, y: number, w: number, h: number,
) => {
  if (count === 1) {
    drawFruit(ctx, fruit, x + w / 2, y + h / 2, Math.min(w, h) * 0.85);
  } else if (count === 2) {
    const sz = Math.min(w / 2, h) * 0.75;
    drawFruit(ctx, fruit, x + w * 0.3, y + h / 2, sz);
    drawFruit(ctx, fruit, x + w * 0.7, y + h / 2, sz);
  } else if (count === 3) {
    const sz = Math.min(w / 2, h / 2) * 0.72;
    drawFruit(ctx, fruit, x + w / 2, y + h * 0.3, sz);
    drawFruit(ctx, fruit, x + w * 0.28, y + h * 0.7, sz);
    drawFruit(ctx, fruit, x + w * 0.72, y + h * 0.7, sz);
  } else {
    const sz = Math.min(w / 2, h / 2) * 0.68;
    drawFruit(ctx, fruit, x + w * 0.28, y + h * 0.28, sz);
    drawFruit(ctx, fruit, x + w * 0.72, y + h * 0.28, sz);
    drawFruit(ctx, fruit, x + w * 0.28, y + h * 0.72, sz);
    drawFruit(ctx, fruit, x + w * 0.72, y + h * 0.72, sz);
  }
};

export const FruitCanvas: React.FC<FruitCanvasProps> = ({ fruit, count, width = 60, height = 60 }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const dpr = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1;
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, width, height);
    layoutFruits(ctx, fruit, count, 0, 0, width, height);
  }, [fruit, count, width, height, dpr]);
  return <canvas ref={canvasRef} style={{ width, height, display: 'block' }} />;
};

export { drawFruit, layoutFruits };
