import { createTheme, ThemeOptions } from '@mui/material/styles';

// 橙红暖调色彩方案 (Option A)
const lightPalette = {
  primary: {
    main: '#FF6B35', // 活力橙
    light: '#FF8F5E',
    dark: '#E64A1A',
  },
  secondary: {
    main: '#F7931E', // 金橙
    light: '#FFB04D',
    dark: '#D67A00',
  },
  error: {
    main: '#E63946', // 热情红
    light: '#FF5A66',
    dark: '#C41E2A',
  },
  background: {
    default: '#F5F5F5',
    paper: '#FFFFFF',
  },
  text: {
    primary: '#1A1A2E',
    secondary: '#6B6B7B',
  },
};

const darkPalette = {
  primary: {
    main: '#FF6B35',
    light: '#FF8F5E',
    dark: '#E64A1A',
  },
  secondary: {
    main: '#F7931E',
    light: '#FFB04D',
    dark: '#D67A00',
  },
  error: {
    main: '#E63946',
    light: '#FF5A66',
    dark: '#C41E2A',
  },
  background: {
    default: '#1A1A2E', // 深蓝灰
    paper: '#252538',
  },
  text: {
    primary: '#FFFFFF',
    secondary: '#B0B0C0',
  },
};

const typography = {
  fontFamily: [
    '-apple-system',
    'BlinkMacSystemFont',
    'PingFang SC',
    'Microsoft YaHei',
    'Inter',
    'Segoe UI',
    'Roboto',
    'sans-serif',
  ].join(','),
  h1: { fontSize: '2.5rem', fontWeight: 700 },
  h2: { fontSize: '2rem', fontWeight: 600 },
  h3: { fontSize: '1.75rem', fontWeight: 600 },
  h4: { fontSize: '1.5rem', fontWeight: 500 },
  h5: { fontSize: '1.25rem', fontWeight: 500 },
  h6: { fontSize: '1rem', fontWeight: 500 },
  body1: { fontSize: '1rem', fontWeight: 400 },
  body2: { fontSize: '0.875rem', fontWeight: 400 },
  button: { fontSize: '0.875rem', fontWeight: 600, textTransform: 'none' as const },
};

const shadows = {
  small: '0 2px 8px rgba(0, 0, 0, 0.1)',
  medium: '0 4px 16px rgba(0, 0, 0, 0.15)',
  large: '0 8px 32px rgba(0, 0, 0, 0.2)',
  card: '0 2px 12px rgba(255, 107, 53, 0.15)',
};

export const createAppTheme = (mode: 'light' | 'dark') => {
  const palette = mode === 'light' ? lightPalette : darkPalette;

  const themeOptions: ThemeOptions = {
    palette: {
      mode,
      ...palette,
    },
    typography,
    spacing: 8,
    shape: {
      borderRadius: 16, // 大圆角
    },
    components: {
      MuiButton: {
        styleOverrides: {
          root: {
            borderRadius: '16px',
            padding: '12px 24px',
            fontSize: '1rem',
            fontWeight: 600,
            textTransform: 'none',
            boxShadow: 'none',
            '&:hover': {
              boxShadow: shadows.small,
            },
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            borderRadius: '16px',
            boxShadow: shadows.card,
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            borderRadius: '16px',
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            borderRadius: 0,
          },
        },
      },
      MuiTextField: {
        styleOverrides: {
          root: {
            '& .MuiOutlinedInput-root': {
              borderRadius: '12px',
            },
          },
        },
      },
      MuiDialog: {
        styleOverrides: {
          paper: {
            borderRadius: '24px',
          },
        },
      },
    },
  };

  return createTheme(themeOptions);
};

export default createAppTheme;
