import { Box } from '@mui/material';
import { useTheme } from '@mui/material/styles';

interface QCardLogoProps {
  size?: number;
  showText?: boolean;
  variant?: 'default' | 'light';
}

export const QCardLogo: React.FC<QCardLogoProps> = ({ size = 40, showText = true, variant = 'default' }) => {
  const theme = useTheme();
  const primary = theme.palette.primary.main;
  const accent = theme.palette.mode === 'dark' ? '#FFB74D' : '#E65100';

  const isLight = variant === 'light';
  const textColor = isLight ? '#fff' : (theme.palette.mode === 'dark' ? '#fff' : primary);

  return (
    <Box sx={{ display: 'inline-flex', alignItems: 'center', gap: showText ? 1 : 0 }}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 100 100"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Back card - tilted left */}
        <g transform="rotate(-12 50 55)">
          <rect
            x="20" y="12" width="60" height="85" rx="8"
            fill={isLight ? 'rgba(255,255,255,0.25)' : (theme.palette.mode === 'dark' ? '#555' : '#ddd')}
            stroke={isLight ? 'rgba(255,255,255,0.4)' : (theme.palette.mode === 'dark' ? '#777' : '#bbb')}
            strokeWidth="2"
          />
          <rect x="26" y="18" width="48" height="73" rx="4" fill={isLight ? 'rgba(255,255,255,0.15)' : (theme.palette.mode === 'dark' ? '#666' : '#ccc')} />
          <path d="M50 30 L56 42 L50 54 L44 42 Z" fill={isLight ? 'rgba(255,255,255,0.3)' : (theme.palette.mode === 'dark' ? '#888' : '#aaa')} />
        </g>

        {/* Front card */}
        <g transform="rotate(8 50 55)">
          <rect
            x="20" y="12" width="60" height="85" rx="8"
            fill={isLight ? 'rgba(255,255,255,0.95)' : (theme.palette.mode === 'dark' ? '#2a2a2a' : '#fff')}
            stroke={isLight ? '#fff' : primary}
            strokeWidth="2.5"
          />
          <text
            x="50" y="58"
            textAnchor="middle"
            dominantBaseline="central"
            fontFamily="'Georgia', serif"
            fontWeight="bold"
            fontSize="42"
            fill={isLight ? primary : primary}
          >
            Q
          </text>
          <text x="28" y="28" fontSize="14" fill={accent}>&#9829;</text>
          <text x="66" y="88" fontSize="14" fill={isLight ? primary : primary}>&#9824;</text>
        </g>

        {/* Sparkle accent */}
        <circle cx="82" cy="18" r="3" fill={isLight ? '#fff' : accent} opacity="0.9" />
        <circle cx="88" cy="26" r="2" fill={isLight ? '#fff' : accent} opacity="0.6" />
        <circle cx="76" cy="12" r="1.5" fill={isLight ? '#fff' : accent} opacity="0.5" />
      </svg>

      {showText && (
        <svg
          width={size * 1.8}
          height={size * 0.7}
          viewBox="0 0 180 60"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <text
            x="0" y="46"
            fontFamily="'Georgia', serif"
            fontWeight="bold"
            fontSize="52"
            fill={textColor}
          >
            Q
          </text>
          <text
            x="42" y="46"
            fontFamily="'Segoe UI', 'Helvetica', sans-serif"
            fontWeight="600"
            fontSize="44"
            fill={textColor}
            letterSpacing="-1"
          >
            Card
          </text>
        </svg>
      )}
    </Box>
  );
};
