import { useRef, useEffect, memo } from 'react';
import { Box } from '@mui/material';
import { motion } from 'framer-motion';
import type { MahjongTile as TileType } from './types';

interface MahjongTileProps {
  tile: TileType;
  selected?: boolean;
  onClick?: () => void;
  size?: 'small' | 'medium';
  faceDown?: boolean;
  highlight?: boolean;
}

const DPR = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1;
const sizes = {
  small: { w: 36, h: 48, r: 5 },
  medium: { w: 50, h: 66, r: 6 },
};

const cnNum = ['一','二','三','四','五','六','七','八','九'];

// ── Drawing helpers ──
function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}

function drawCircle(ctx: CanvasRenderingContext2D, cx: number, cy: number, r: number, color: string) {
  ctx.beginPath();
  ctx.arc(cx, cy, r, 0, Math.PI * 2);
  ctx.strokeStyle = color;
  ctx.lineWidth = 1.5;
  ctx.stroke();
  ctx.beginPath();
  ctx.arc(cx, cy, r * 0.35, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
}

function drawStick(ctx: CanvasRenderingContext2D, cx: number, cy: number, sw: number, sh: number, color: string) {
  const x = cx - sw / 2, y = cy - sh / 2;
  ctx.beginPath();
  ctx.moveTo(x + sw / 2, y);
  ctx.arcTo(x + sw, y, x + sw, y + sh, sw / 2);
  ctx.arcTo(x + sw, y + sh, x, y + sh, sw / 2);
  ctx.arcTo(x, y + sh, x, y, sw / 2);
  ctx.arcTo(x, y, x + sw, y, sw / 2);
  ctx.closePath();
  ctx.strokeStyle = color;
  ctx.lineWidth = 1.3;
  ctx.stroke();
}

// Draw a bamboo stick segment between two points with knobs at ends
function drawBambooStickSegment(
  ctx: CanvasRenderingContext2D,
  x1: number, y1: number, x2: number, y2: number,
  thickness: number, color: string,
) {
  ctx.strokeStyle = color;
  ctx.fillStyle = color;
  const knob = thickness * 0.7;
  // Stick body
  ctx.lineWidth = thickness;
  ctx.lineCap = 'round';
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.lineTo(x2, y2);
  ctx.stroke();
  // Knobs at both ends
  ctx.beginPath();
  ctx.arc(x1, y1, knob, 0, Math.PI * 2);
  ctx.fill();
  ctx.beginPath();
  ctx.arc(x2, y2, knob, 0, Math.PI * 2);
  ctx.fill();
}

// Draw two M shapes (inverted M top, normal M bottom) using bamboo stick segments
function drawBambooM(ctx: CanvasRenderingContext2D, w: number, h: number, color: string) {
  const t = Math.max(1.2, w * 0.05); // stick thickness
  const lx = w * 0.1, rx = w * 0.9, mx = w * 0.5;
  const lmx = w * 0.3, rmx = w * 0.7;

  // Top half: inverted M (looks like W) — sticks go from top corners down to valley, up to peak, down to valley, up to top corner
  const tTop = h * 0.06;
  const tBot = h * 0.44;
  const tMid = h * 0.2;
  // Left vertical
  drawBambooStickSegment(ctx, lx, tTop, lx, tBot, t, color);
  // Left diagonal up
  drawBambooStickSegment(ctx, lx, tBot, lmx, tMid, t, color);
  // Center-left diagonal down
  drawBambooStickSegment(ctx, lmx, tMid, mx, tBot, t, color);
  // Center-right diagonal up
  drawBambooStickSegment(ctx, mx, tBot, rmx, tMid, t, color);
  // Right diagonal down
  drawBambooStickSegment(ctx, rmx, tMid, rx, tBot, t, color);
  // Right vertical
  drawBambooStickSegment(ctx, rx, tTop, rx, tBot, t, color);

  // Bottom half: normal M — sticks go from bottom corners up to peak, down to valley, up to peak, down to bottom corner
  const bBot = h * 0.94;
  const bTop = h * 0.56;
  const bMid = h * 0.8;
  // Left vertical
  drawBambooStickSegment(ctx, lx, bBot, lx, bTop, t, color);
  // Left diagonal down
  drawBambooStickSegment(ctx, lx, bTop, lmx, bMid, t, color);
  // Center-left diagonal up
  drawBambooStickSegment(ctx, lmx, bMid, mx, bTop, t, color);
  // Center-right diagonal down
  drawBambooStickSegment(ctx, mx, bTop, rmx, bMid, t, color);
  // Right diagonal up
  drawBambooStickSegment(ctx, rmx, bMid, rx, bTop, t, color);
  // Right vertical
  drawBambooStickSegment(ctx, rx, bBot, rx, bTop, t, color);
}

// ── Content drawers ──
function drawWan(ctx: CanvasRenderingContext2D, rank: number, w: number, h: number) {
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  // Number in black
  ctx.fillStyle = '#212121';
  ctx.font = `bold ${Math.round(h * 0.32)}px serif`;
  ctx.fillText(cnNum[rank - 1], w / 2, h * 0.38);
  // 萬 in red — larger
  ctx.fillStyle = '#C62828';
  ctx.font = `bold ${Math.round(h * 0.28)}px serif`;
  ctx.fillText('萬', w / 2, h * 0.72);
}

function drawTiao(ctx: CanvasRenderingContext2D, rank: number, w: number, h: number) {
  const G = '#2E7D32', R = '#C62828';
  const sw = w * 0.12, sh = h * 0.22;

  if (rank === 1) {
    // Cute bird — centered, round body + head + wing + tail
    const cx = w / 2, cy = h / 2;
    // Body
    ctx.fillStyle = '#43A047';
    ctx.beginPath();
    ctx.ellipse(cx, cy + h * 0.06, w * 0.28, h * 0.16, 0, 0, Math.PI * 2);
    ctx.fill();
    // Head
    ctx.fillStyle = '#4CAF50';
    ctx.beginPath();
    ctx.arc(cx + w * 0.08, cy - h * 0.12, w * 0.15, 0, Math.PI * 2);
    ctx.fill();
    // Eye
    ctx.fillStyle = '#fff';
    ctx.beginPath();
    ctx.arc(cx + w * 0.14, cy - h * 0.14, w * 0.05, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = '#1B5E20';
    ctx.beginPath();
    ctx.arc(cx + w * 0.15, cy - h * 0.14, w * 0.025, 0, Math.PI * 2);
    ctx.fill();
    // Beak
    ctx.fillStyle = '#FF8F00';
    ctx.beginPath();
    ctx.moveTo(cx + w * 0.22, cy - h * 0.12);
    ctx.lineTo(cx + w * 0.34, cy - h * 0.14);
    ctx.lineTo(cx + w * 0.22, cy - h * 0.08);
    ctx.closePath();
    ctx.fill();
    // Wing
    ctx.fillStyle = '#2E7D32';
    ctx.beginPath();
    ctx.ellipse(cx - w * 0.08, cy + h * 0.02, w * 0.16, h * 0.1, -0.3, 0, Math.PI * 2);
    ctx.fill();
    // Tail feathers
    ctx.strokeStyle = '#2E7D32';
    ctx.lineWidth = 1.5;
    ctx.lineCap = 'round';
    ctx.beginPath();
    ctx.moveTo(cx - w * 0.22, cy + h * 0.04);
    ctx.quadraticCurveTo(cx - w * 0.38, cy - h * 0.05, cx - w * 0.35, cy + h * 0.18);
    ctx.stroke();
    ctx.beginPath();
    ctx.moveTo(cx - w * 0.2, cy + h * 0.08);
    ctx.quadraticCurveTo(cx - w * 0.36, cy + h * 0.02, cx - w * 0.3, cy + h * 0.22);
    ctx.stroke();
    return;
  }

  if (rank === 8) {
    // Two M shapes: top inverted M, bottom normal M, hollow green
    ctx.strokeStyle = G;
    ctx.lineWidth = 1.8;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';
    const mx = w * 0.1, ex = w * 0.9;
    // Top: inverted M (W shape)
    ctx.beginPath();
    ctx.moveTo(mx, h * 0.08);
    ctx.lineTo(mx, h * 0.44);
    ctx.lineTo(w * 0.3, h * 0.22);
    ctx.lineTo(w / 2, h * 0.44);
    ctx.lineTo(w * 0.7, h * 0.22);
    ctx.lineTo(ex, h * 0.44);
    ctx.lineTo(ex, h * 0.08);
    ctx.stroke();
    // Bottom: normal M
    ctx.beginPath();
    ctx.moveTo(mx, h * 0.92);
    ctx.lineTo(mx, h * 0.56);
    ctx.lineTo(w * 0.3, h * 0.78);
    ctx.lineTo(w / 2, h * 0.56);
    ctx.lineTo(w * 0.7, h * 0.78);
    ctx.lineTo(ex, h * 0.56);
    ctx.lineTo(ex, h * 0.92);
    ctx.stroke();
    return;
  }

  if (rank === 8) {
    // Two M shapes made of bamboo sticks with knobs at joints
    // Top: inverted M (W shape), Bottom: normal M
    drawBambooM(ctx, w, h, G);
    return;
  }

  // Stick layouts: [cx%, cy%, isRed]
  type SP = [number, number, boolean];
  const layouts: Record<number, SP[]> = {
    2: [[0.5,0.3,false],[0.5,0.7,false]],
    3: [[0.5,0.2,false],[0.3,0.65,false],[0.7,0.65,false]],
    4: [[0.33,0.3,false],[0.67,0.3,false],[0.33,0.7,false],[0.67,0.7,false]],
    5: [[0.33,0.22,false],[0.67,0.22,false],[0.5,0.5,true],[0.33,0.78,false],[0.67,0.78,false]],
    6: [[0.33,0.2,false],[0.67,0.2,false],[0.33,0.5,false],[0.67,0.5,false],[0.33,0.8,false],[0.67,0.8,false]],
    7: [[0.5,0.12,true],[0.35,0.34,false],[0.65,0.34,false],[0.35,0.54,false],[0.65,0.54,false],[0.35,0.74,false],[0.65,0.74,false]],
    9: [[0.22,0.18,false],[0.5,0.18,false],[0.78,0.18,false],[0.22,0.5,false],[0.5,0.5,true],[0.78,0.5,false],[0.22,0.82,false],[0.5,0.82,false],[0.78,0.82,false]],
  };
  const pts = layouts[rank] || [];
  // 7条 sticks are smaller since there are 7 packed in
  const actualSw = rank === 7 ? sw * 0.75 : sw;
  const actualSh = rank === 7 ? sh * 0.75 : sh;
  for (const [px, py, red] of pts) {
    drawStick(ctx, w * px, h * py, actualSw, actualSh, red ? R : G);
  }
}

function drawTong(ctx: CanvasRenderingContext2D, rank: number, w: number, h: number) {
  const G = '#2E7D32', R = '#C62828', BK = '#333';

  if (rank === 1) {
    // Large ornate circle: outer black, inner red, center red
    drawCircle(ctx, w / 2, h / 2, h * 0.28, BK);
    ctx.beginPath();
    ctx.arc(w / 2, h / 2, h * 0.18, 0, Math.PI * 2);
    ctx.strokeStyle = R;
    ctx.lineWidth = 2;
    ctx.stroke();
    ctx.beginPath();
    ctx.arc(w / 2, h / 2, h * 0.07, 0, Math.PI * 2);
    ctx.fillStyle = R;
    ctx.fill();
    return;
  }

  const cr = rank <= 4 ? h * 0.1 : rank <= 6 ? h * 0.09 : h * 0.075;

  // [cx%, cy%, color] — all layouts centered around (0.5, 0.5)
  type CP = [number, number, string];
  const layouts: Record<number, CP[]> = {
    2: [[0.5,0.3,G],[0.5,0.7,R]],
    3: [[0.65,0.25,BK],[0.5,0.5,R],[0.35,0.75,G]],
    4: [[0.3,0.28,BK],[0.7,0.28,G],[0.3,0.72,G],[0.7,0.72,BK]],
    5: [[0.3,0.25,G],[0.7,0.25,BK],[0.5,0.5,R],[0.3,0.75,BK],[0.7,0.75,G]],
    6: [[0.35,0.2,G],[0.65,0.2,BK],[0.35,0.5,R],[0.65,0.5,G],[0.35,0.8,BK],[0.65,0.8,R]],
    7: [[0.7,0.13,G],[0.5,0.25,G],[0.3,0.37,G],[0.35,0.58,R],[0.65,0.58,R],[0.35,0.78,R],[0.65,0.78,R]],
    8: [[0.35,0.14,BK],[0.65,0.14,BK],[0.35,0.38,BK],[0.65,0.38,BK],[0.35,0.62,BK],[0.65,0.62,BK],[0.35,0.86,BK],[0.65,0.86,BK]],
    9: [[0.25,0.18,G],[0.5,0.18,R],[0.75,0.18,BK],[0.25,0.5,BK],[0.5,0.5,G],[0.75,0.5,R],[0.25,0.82,R],[0.5,0.82,BK],[0.75,0.82,G]],
  };
  const pts = layouts[rank] || [];
  for (const [px, py, color] of pts) {
    drawCircle(ctx, w * px, h * py, cr, color);
  }
}

function drawZi(ctx: CanvasRenderingContext2D, rank: number, w: number, h: number) {
  const zi: Record<number, { char: string; color: string }> = {
    1: { char: '東', color: '#1565C0' },
    2: { char: '南', color: '#2E7D32' },
    3: { char: '西', color: '#6A1B9A' },
    4: { char: '北', color: '#37474F' },
    5: { char: '中', color: '#C62828' },
    6: { char: '發', color: '#2E7D32' },
    7: { char: '白', color: '#9E9E9E' },
  };
  const cfg = zi[rank];
  if (!cfg) return;

  if (rank === 7) {
    // 白板: empty rectangle
    const rw = w * 0.5, rh = h * 0.55;
    ctx.strokeStyle = '#BDBDBD';
    ctx.lineWidth = 2;
    ctx.strokeRect((w - rw) / 2, (h - rh) / 2, rw, rh);
    return;
  }

  ctx.fillStyle = cfg.color;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.font = `bold ${Math.round(h * 0.48)}px serif`;
  ctx.fillText(cfg.char, w / 2, h / 2);
}

// ── Main draw function ──
function drawTileCanvas(
  canvas: HTMLCanvasElement,
  tile: TileType,
  w: number, h: number, r: number,
  faceDown: boolean,
  selected: boolean,
  highlight: boolean,
) {
  const cw = w * DPR, ch = h * DPR;
  canvas.width = cw;
  canvas.height = ch;
  canvas.style.width = w + 'px';
  canvas.style.height = h + 'px';

  const ctx = canvas.getContext('2d');
  if (!ctx) return;
  ctx.scale(DPR, DPR);
  ctx.clearRect(0, 0, w, h);

  const pad = selected || highlight ? 1.5 : 0.5;

  // Shadow
  ctx.save();
  ctx.shadowColor = selected ? 'rgba(255,107,53,0.35)' : highlight ? 'rgba(76,175,80,0.3)' : 'rgba(255,107,53,0.06)';
  ctx.shadowBlur = selected ? 6 : 2;
  ctx.shadowOffsetY = 1;
  roundRect(ctx, pad, pad, w - pad * 2, h - pad * 2, r);
  ctx.fillStyle = faceDown ? '#FF8F5E' : '#FFFDF8';
  ctx.fill();
  ctx.restore();

  // Border — warm tint
  roundRect(ctx, pad, pad, w - pad * 2, h - pad * 2, r);
  ctx.strokeStyle = selected ? '#FF6B35' : highlight ? '#66BB6A' : '#E8D5C0';
  ctx.lineWidth = selected || highlight ? 2 : 1;
  ctx.stroke();

  if (faceDown) {
    ctx.strokeStyle = 'rgba(255,255,255,0.25)';
    ctx.lineWidth = 1;
    const iw = w - 10, ih = h - 10;
    roundRect(ctx, 5, 5, iw, ih, r - 1);
    ctx.stroke();
    return;
  }

  // Clip to tile area for content — uniform centered content area
  ctx.save();
  const inset = Math.round(Math.min(w, h) * 0.12);
  roundRect(ctx, inset, inset, w - inset * 2, h - inset * 2, r - 1);
  ctx.clip();

  const cw2 = w - inset * 2, ch2 = h - inset * 2;
  ctx.translate(inset, inset);

  if (tile.suit === 'wan') drawWan(ctx, tile.rank, cw2, ch2);
  else if (tile.suit === 'tiao') drawTiao(ctx, tile.rank, cw2, ch2);
  else if (tile.suit === 'tong') drawTong(ctx, tile.rank, cw2, ch2);
  else if (tile.suit === 'zi') drawZi(ctx, tile.rank, cw2, ch2);

  ctx.restore();
}

// ── React components ──
const TileCanvas = memo<{
  tile: TileType;
  w: number; h: number; r: number;
  faceDown: boolean; selected: boolean; highlight: boolean;
}>(({ tile, w, h, r, faceDown, selected, highlight }) => {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (ref.current) {
      drawTileCanvas(ref.current, tile, w, h, r, faceDown, selected, highlight);
    }
  }, [tile, w, h, r, faceDown, selected, highlight]);

  return <canvas ref={ref} style={{ display: 'block' }} />;
});

export const MahjongTileComp: React.FC<MahjongTileProps> = ({
  tile,
  selected = false,
  onClick,
  size = 'medium',
  faceDown = false,
  highlight = false,
}) => {
  const sz = sizes[size];

  return (
    <Box
      component={motion.div}
      initial={false}
      animate={{ y: selected ? -10 : 0 }}
      whileHover={onClick ? { y: selected ? -10 : -3 } : undefined}
      transition={{ duration: 0.1 }}
      onClick={onClick}
      sx={{
        width: sz.w,
        height: sz.h,
        cursor: onClick ? 'pointer' : 'default',
        userSelect: 'none',
        flexShrink: 0,
      }}
    >
      <TileCanvas
        tile={tile}
        w={sz.w} h={sz.h} r={sz.r}
        faceDown={faceDown}
        selected={selected} highlight={highlight}
      />
    </Box>
  );
};
