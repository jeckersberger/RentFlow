import { Theme } from '@radix-ui/themes'

/**
 * Radix UI Theme Configuration
 * Maps CrateDesk design tokens to Radix UI theme colors
 */
export const radixThemeConfig: Partial<React.ComponentProps<typeof Theme>> = {
  appearance: 'inherit',
  accentColor: 'blue',
  grayColor: 'slate',
  panelBackground: 'translucent',
  scaling: '100%',
}

/**
 * Color palette mapping from CrateDesk SCSS tokens to Radix UI colors
 * Used for consistent styling across the application
 */
export const colorPalette = {
  // Primary Colors (Blue)
  primary: {
    50: 'var(--color-primary-50)',
    100: 'var(--color-primary-100)',
    200: 'var(--color-primary-200)',
    300: 'var(--color-primary-300)',
    400: 'var(--color-primary-400)',
    500: 'var(--color-primary-500)',
    600: 'var(--color-primary-600)',
    700: 'var(--color-primary-700)',
    800: 'var(--color-primary-800)',
    900: 'var(--color-primary-900)',
  },

  // Accent Colors
  success: {
    light: 'var(--color-success-light)',
    base: 'var(--color-success)',
    dark: 'var(--color-success-dark)',
  },
  warning: {
    light: 'var(--color-warning-light)',
    base: 'var(--color-warning)',
    dark: 'var(--color-warning-dark)',
  },
  danger: {
    light: 'var(--color-danger-light)',
    base: 'var(--color-danger)',
    dark: 'var(--color-danger-dark)',
  },
  info: {
    light: 'var(--color-info-light)',
    base: 'var(--color-info)',
    dark: 'var(--color-info-dark)',
  },

  // Neutral Colors (Grays)
  gray: {
    50: 'var(--color-gray-50)',
    100: 'var(--color-gray-100)',
    200: 'var(--color-gray-200)',
    300: 'var(--color-gray-300)',
    400: 'var(--color-gray-400)',
    500: 'var(--color-gray-500)',
    600: 'var(--color-gray-600)',
    700: 'var(--color-gray-700)',
    800: 'var(--color-gray-800)',
    900: 'var(--color-gray-900)',
    950: 'var(--color-gray-950)',
  },

  // Semantic Colors
  background: {
    primary: 'var(--color-bg-primary)',
    secondary: 'var(--color-bg-secondary)',
    tertiary: 'var(--color-bg-tertiary)',
  },
  text: {
    primary: 'var(--color-text-primary)',
    secondary: 'var(--color-text-secondary)',
  },
  border: 'var(--color-border)',
}

/**
 * Size and spacing tokens
 */
export const spacing = {
  xs: 'var(--spacing-2)',
  sm: 'var(--spacing-3)',
  md: 'var(--spacing-4)',
  lg: 'var(--spacing-6)',
  xl: 'var(--spacing-8)',
}

export const borderRadius = {
  none: 'var(--radius-none)',
  sm: 'var(--radius-sm)',
  base: 'var(--radius-base)',
  md: 'var(--radius-md)',
  lg: 'var(--radius-lg)',
  xl: 'var(--radius-xl)',
  '2xl': 'var(--radius-2xl)',
  full: 'var(--radius-full)',
}
