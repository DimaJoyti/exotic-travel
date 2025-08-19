/**
 * Design Tokens System for Exotic Travel Booking Platform
 * 
 * This file defines the comprehensive design system tokens including:
 * - Colors (semantic and brand colors)
 * - Typography (font families, sizes, weights)
 * - Spacing (consistent spacing scale)
 * - Shadows (elevation system)
 * - Border radius (consistent rounded corners)
 * - Animations (durations and easings)
 */

// Color System - HSL values for better manipulation
export const colors = {
  // Brand Colors - Travel-themed palette
  brand: {
    primary: {
      50: 'hsl(210, 100%, 97%)',
      100: 'hsl(210, 100%, 94%)',
      200: 'hsl(210, 100%, 87%)',
      300: 'hsl(210, 100%, 78%)',
      400: 'hsl(210, 100%, 66%)',
      500: 'hsl(210, 100%, 56%)', // Main brand color - Ocean Blue
      600: 'hsl(210, 100%, 47%)',
      700: 'hsl(210, 100%, 39%)',
      800: 'hsl(210, 100%, 31%)',
      900: 'hsl(210, 100%, 24%)',
      950: 'hsl(210, 100%, 15%)',
    },
    secondary: {
      50: 'hsl(45, 100%, 97%)',
      100: 'hsl(45, 100%, 94%)',
      200: 'hsl(45, 100%, 87%)',
      300: 'hsl(45, 100%, 78%)',
      400: 'hsl(45, 100%, 66%)',
      500: 'hsl(45, 100%, 56%)', // Sunset Orange
      600: 'hsl(45, 100%, 47%)',
      700: 'hsl(45, 100%, 39%)',
      800: 'hsl(45, 100%, 31%)',
      900: 'hsl(45, 100%, 24%)',
      950: 'hsl(45, 100%, 15%)',
    },
    accent: {
      50: 'hsl(165, 100%, 97%)',
      100: 'hsl(165, 100%, 94%)',
      200: 'hsl(165, 100%, 87%)',
      300: 'hsl(165, 100%, 78%)',
      400: 'hsl(165, 100%, 66%)',
      500: 'hsl(165, 100%, 56%)', // Tropical Teal
      600: 'hsl(165, 100%, 47%)',
      700: 'hsl(165, 100%, 39%)',
      800: 'hsl(165, 100%, 31%)',
      900: 'hsl(165, 100%, 24%)',
      950: 'hsl(165, 100%, 15%)',
    },
  },

  // Semantic Colors
  semantic: {
    success: {
      50: 'hsl(142, 76%, 96%)',
      100: 'hsl(142, 76%, 91%)',
      200: 'hsl(142, 76%, 81%)',
      300: 'hsl(142, 76%, 69%)',
      400: 'hsl(142, 76%, 55%)',
      500: 'hsl(142, 76%, 45%)', // Success Green
      600: 'hsl(142, 76%, 36%)',
      700: 'hsl(142, 76%, 29%)',
      800: 'hsl(142, 76%, 24%)',
      900: 'hsl(142, 76%, 20%)',
      950: 'hsl(142, 76%, 12%)',
    },
    warning: {
      50: 'hsl(48, 96%, 96%)',
      100: 'hsl(48, 96%, 89%)',
      200: 'hsl(48, 96%, 76%)',
      300: 'hsl(48, 96%, 61%)',
      400: 'hsl(48, 96%, 53%)',
      500: 'hsl(48, 96%, 47%)', // Warning Amber
      600: 'hsl(48, 96%, 37%)',
      700: 'hsl(48, 96%, 27%)',
      800: 'hsl(48, 96%, 23%)',
      900: 'hsl(48, 96%, 20%)',
      950: 'hsl(48, 96%, 11%)',
    },
    error: {
      50: 'hsl(0, 86%, 97%)',
      100: 'hsl(0, 86%, 94%)',
      200: 'hsl(0, 86%, 87%)',
      300: 'hsl(0, 86%, 78%)',
      400: 'hsl(0, 86%, 67%)',
      500: 'hsl(0, 86%, 55%)', // Error Red
      600: 'hsl(0, 86%, 47%)',
      700: 'hsl(0, 86%, 39%)',
      800: 'hsl(0, 86%, 31%)',
      900: 'hsl(0, 86%, 26%)',
      950: 'hsl(0, 86%, 15%)',
    },
    info: {
      50: 'hsl(204, 100%, 97%)',
      100: 'hsl(204, 100%, 94%)',
      200: 'hsl(204, 100%, 87%)',
      300: 'hsl(204, 100%, 78%)',
      400: 'hsl(204, 100%, 67%)',
      500: 'hsl(204, 100%, 56%)', // Info Blue
      600: 'hsl(204, 100%, 47%)',
      700: 'hsl(204, 100%, 39%)',
      800: 'hsl(204, 100%, 31%)',
      900: 'hsl(204, 100%, 24%)',
      950: 'hsl(204, 100%, 15%)',
    },
  },

  // Neutral Colors
  neutral: {
    50: 'hsl(210, 20%, 98%)',
    100: 'hsl(210, 20%, 95%)',
    200: 'hsl(210, 16%, 93%)',
    300: 'hsl(210, 14%, 89%)',
    400: 'hsl(210, 14%, 83%)',
    500: 'hsl(210, 11%, 71%)',
    600: 'hsl(210, 7%, 56%)',
    700: 'hsl(210, 9%, 31%)',
    800: 'hsl(210, 10%, 23%)',
    900: 'hsl(210, 11%, 15%)',
    950: 'hsl(210, 11%, 9%)',
  },
} as const

// Typography System
export const typography = {
  fontFamilies: {
    display: ['Playfair Display', 'Georgia', 'serif'],
    body: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
    mono: ['JetBrains Mono', 'Menlo', 'Monaco', 'monospace'],
  },
  
  fontSizes: {
    xs: '0.75rem',    // 12px
    sm: '0.875rem',   // 14px
    base: '1rem',     // 16px
    lg: '1.125rem',   // 18px
    xl: '1.25rem',    // 20px
    '2xl': '1.5rem',  // 24px
    '3xl': '1.875rem', // 30px
    '4xl': '2.25rem', // 36px
    '5xl': '3rem',    // 48px
    '6xl': '3.75rem', // 60px
    '7xl': '4.5rem',  // 72px
    '8xl': '6rem',    // 96px
    '9xl': '8rem',    // 128px
  },
  
  fontWeights: {
    thin: '100',
    extralight: '200',
    light: '300',
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
    extrabold: '800',
    black: '900',
  },
  
  lineHeights: {
    none: '1',
    tight: '1.25',
    snug: '1.375',
    normal: '1.5',
    relaxed: '1.625',
    loose: '2',
  },
  
  letterSpacing: {
    tighter: '-0.05em',
    tight: '-0.025em',
    normal: '0em',
    wide: '0.025em',
    wider: '0.05em',
    widest: '0.1em',
  },
} as const

// Spacing System - 8px base unit
export const spacing = {
  0: '0px',
  px: '1px',
  0.5: '0.125rem', // 2px
  1: '0.25rem',    // 4px
  1.5: '0.375rem', // 6px
  2: '0.5rem',     // 8px
  2.5: '0.625rem', // 10px
  3: '0.75rem',    // 12px
  3.5: '0.875rem', // 14px
  4: '1rem',       // 16px
  5: '1.25rem',    // 20px
  6: '1.5rem',     // 24px
  7: '1.75rem',    // 28px
  8: '2rem',       // 32px
  9: '2.25rem',    // 36px
  10: '2.5rem',    // 40px
  11: '2.75rem',   // 44px
  12: '3rem',      // 48px
  14: '3.5rem',    // 56px
  16: '4rem',      // 64px
  20: '5rem',      // 80px
  24: '6rem',      // 96px
  28: '7rem',      // 112px
  32: '8rem',      // 128px
  36: '9rem',      // 144px
  40: '10rem',     // 160px
  44: '11rem',     // 176px
  48: '12rem',     // 192px
  52: '13rem',     // 208px
  56: '14rem',     // 224px
  60: '15rem',     // 240px
  64: '16rem',     // 256px
  72: '18rem',     // 288px
  80: '20rem',     // 320px
  96: '24rem',     // 384px
} as const

// Shadow System - Elevation levels
export const shadows = {
  none: 'none',
  sm: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  base: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.25)',
  inner: 'inset 0 2px 4px 0 rgb(0 0 0 / 0.05)',
  
  // Travel-themed shadows
  glow: '0 0 20px rgb(59 130 246 / 0.5)',
  warm: '0 8px 32px rgb(251 146 60 / 0.35)',
  cool: '0 8px 32px rgb(14 165 233 / 0.35)',
} as const

// Border Radius System
export const borderRadius = {
  none: '0px',
  sm: '0.125rem',   // 2px
  base: '0.25rem',  // 4px
  md: '0.375rem',   // 6px
  lg: '0.5rem',     // 8px
  xl: '0.75rem',    // 12px
  '2xl': '1rem',    // 16px
  '3xl': '1.5rem',  // 24px
  full: '9999px',
} as const

// Animation System
export const animations = {
  durations: {
    instant: '0ms',
    fast: '150ms',
    normal: '300ms',
    slow: '500ms',
    slower: '750ms',
    slowest: '1000ms',
  },
  
  easings: {
    linear: 'linear',
    easeIn: 'cubic-bezier(0.4, 0, 1, 1)',
    easeOut: 'cubic-bezier(0, 0, 0.2, 1)',
    easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
    spring: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
    bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
  },
} as const

// Breakpoints for responsive design
export const breakpoints = {
  xs: '475px',
  sm: '640px',
  md: '768px',
  lg: '1024px',
  xl: '1280px',
  '2xl': '1536px',
} as const

// Z-index scale
export const zIndex = {
  hide: -1,
  auto: 'auto',
  base: 0,
  docked: 10,
  dropdown: 1000,
  sticky: 1100,
  banner: 1200,
  overlay: 1300,
  modal: 1400,
  popover: 1500,
  skipLink: 1600,
  toast: 1700,
  tooltip: 1800,
} as const

// Export all tokens as a single object
export const designTokens = {
  colors,
  typography,
  spacing,
  shadows,
  borderRadius,
  animations,
  breakpoints,
  zIndex,
} as const

export type DesignTokens = typeof designTokens
export type ColorScale = typeof colors.brand.primary
export type SpacingScale = typeof spacing
export type FontSize = keyof typeof typography.fontSizes
export type FontWeight = keyof typeof typography.fontWeights
