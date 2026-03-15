import React, { useRef, useEffect } from 'react';

interface AngerTokenProps {
  value: number; // 1-7
  size?: number;
}

// 绘制猩猩愤怒标记 - 扇形/半圆形黑色背景，猩猩脸 + 头顶分数
// 分数越大，猩猩颜色越深红
const drawAngerToken = (ctx: CanvasRenderingContext2D, w: number, h: number, value: number) => {
  const cx = w / 2;
  const r = w * 0.45;
  const bottomY = h * 0.85;

  // 背景 - 半圆/拱形
  ctx.beginPath();
  ctx.moveTo(cx - r, bottomY);
  ctx.arc(cx, bottomY, r, Math.PI, 0, false);
  ctx.closePath();
  ctx.fillStyle = '#1a1a1a';
  ctx.fill();

  // 猩猩颜色随分数变化: 1=暗灰, 7=深红
  const t = (value - 1) / 6;
  const bodyR = Math.floor(60 + t * 140); // 60 -> 200
  const bodyG = Math.floor(40 - t * 30);  // 40 -> 10
  const bodyB = Math.floor(50 - t * 30);  // 50 -> 20
  const bodyColor = `rgb(${bodyR},${bodyG},${bodyB})`;

  // 星星装饰 (分数越高越多)
  const starCount = Math.min(value, 5);
  const starColor = `rgba(${180 + t * 75}, ${40 + t * 20}, ${60 + t * 20}, ${0.4 + t * 0.4})`;
  for (let i = 0; i < starCount; i++) {
    const angle = Math.PI + (Math.PI * (i + 0.5)) / (starCount);
    const sr = r * (0.6 + Math.random() * 0.3);
    const sx = cx + Math.cos(angle) * sr;
    const sy = bottomY + Math.sin(angle) * sr;
    drawStar(ctx, sx, sy, w * 0.04, starColor);
  }

  // 猩猩身体 (半圆形)
  const gorillaR = r * 0.5;
  const gorillaCY = bottomY - r * 0.35;
  ctx.beginPath();
  ctx.arc(cx, gorillaCY, gorillaR, 0, Math.PI * 2);
  ctx.fillStyle = bodyColor;
  ctx.fill();

  // 耳朵
  ctx.beginPath();
  ctx.arc(cx - gorillaR * 0.85, gorillaCY - gorillaR * 0.2, gorillaR * 0.3, 0, Math.PI * 2);
  ctx.fillStyle = bodyColor;
  ctx.fill();
  ctx.beginPath();
  ctx.arc(cx + gorillaR * 0.85, gorillaCY - gorillaR * 0.2, gorillaR * 0.3, 0, Math.PI * 2);
  ctx.fillStyle = bodyColor;
  ctx.fill();

  // 眼睛 - 墨镜效果 (分数越高墨镜越暗)
  const glassOpacity = 0.6 + t * 0.4;
  // 左眼
  ctx.beginPath();
  ctx.ellipse(cx - gorillaR * 0.3, gorillaCY - gorillaR * 0.05, gorillaR * 0.25, gorillaR * 0.15, 0, 0, Math.PI * 2);
  ctx.fillStyle = `rgba(0,0,0,${glassOpacity})`;
  ctx.fill();
  // 右眼
  ctx.beginPath();
  ctx.ellipse(cx + gorillaR * 0.3, gorillaCY - gorillaR * 0.05, gorillaR * 0.25, gorillaR * 0.15, 0, 0, Math.PI * 2);
  ctx.fillStyle = `rgba(0,0,0,${glassOpacity})`;
  ctx.fill();
  // 墨镜桥
  ctx.beginPath();
  ctx.moveTo(cx - gorillaR * 0.08, gorillaCY - gorillaR * 0.05);
  ctx.lineTo(cx + gorillaR * 0.08, gorillaCY - gorillaR * 0.05);
  ctx.strokeStyle = `rgba(0,0,0,${glassOpacity})`;
  ctx.lineWidth = gorillaR * 0.08;
  ctx.stroke();

  // 嘴巴
  ctx.beginPath();
  ctx.arc(cx, gorillaCY + gorillaR * 0.35, gorillaR * 0.2, 0.1, Math.PI - 0.1);
  ctx.strokeStyle = `rgba(0,0,0,0.5)`;
  ctx.lineWidth = gorillaR * 0.06;
  ctx.stroke();

  // 分数 - 头顶
  const fontSize = w * 0.3;
  ctx.font = `bold ${fontSize}px Arial`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'bottom';
  ctx.fillStyle = '#fff';
  ctx.fillText(String(value), cx, gorillaCY - gorillaR - w * 0.02);

  // 小分母 /7
  const smallFs = w * 0.12;
  ctx.font = `${smallFs}px Arial`;
  ctx.fillStyle = 'rgba(255,255,255,0.6)';
  ctx.fillText('/7', cx + fontSize * 0.35, gorillaCY - gorillaR - w * 0.02);
};

const drawStar = (ctx: CanvasRenderingContext2D, cx: number, cy: number, r: number, color: string) => {
  ctx.beginPath();
  for (let i = 0; i < 4; i++) {
    const angle = (i / 4) * Math.PI * 2;
    const x = cx + Math.cos(angle) * r;
    const y = cy + Math.sin(angle) * r;
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  }
  ctx.closePath();
  ctx.fillStyle = color;
  ctx.fill();
};

export const AngerToken: React.FC<AngerTokenProps> = ({ value, size = 50 }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const dpr = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1;
  const h = size * 0.9;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    canvas.width = size * dpr;
    canvas.height = h * dpr;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, size, h);
    drawAngerToken(ctx, size, h, value);
  }, [value, size, h, dpr]);

  return (
    <canvas
      ref={canvasRef}
      style={{ width: size, height: h, display: 'block' }}
    />
  );
};
