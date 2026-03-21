# Radix UI Theme Integration

This directory contains the Radix UI theme configuration for RentFlow.

## Overview

The theme integrates Radix UI's composable components with RentFlow's design tokens defined in `/src/styles/variables.scss`.

## Configuration

### `radix.config.ts`

Contains:
- **radixThemeConfig**: Radix UI Theme component props
- **colorPalette**: Complete color mapping from design tokens to CSS custom properties
- **spacing**: Spacing scale aligned with RentFlow tokens
- **borderRadius**: Border radius values aligned with design tokens

## Using Colors in Components

All colors are mapped to CSS custom properties, so you can use them in styles:

```tsx
// In component SCSS
.button {
  background-color: var(--color-primary);
  color: var(--color-white);
}
```

## Using Radix UI Components

Radix UI provides unstyled, accessible primitives. Import from `@radix-ui/themes`:

```tsx
import { Button, Text, Card } from '@radix-ui/themes'

export function MyComponent() {
  return (
    <Card>
      <Text>Hello World</Text>
      <Button>Click me</Button>
    </Card>
  )
}
```

## Theme Switching

Dark/light mode is controlled through the `dark-mode` class on `document.documentElement`.

The theme automatically updates in `App.tsx` based on this class.

## Design Tokens Reference

### Colors

- **Primary**: Blue (#3b82f6) with 50-900 shades
- **Success**: Green (#10b981)
- **Warning**: Amber (#f59e0b)
- **Danger**: Red (#ef4444)
- **Info**: Cyan (#06b6d4)
- **Grays**: 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950

### Spacing (4px base unit)

- xs: 0.5rem (8px)
- sm: 0.75rem (12px)
- md: 1rem (16px)
- lg: 1.5rem (24px)
- xl: 2rem (32px)

### Border Radius

- sm: 4px
- base: 6px
- md: 8px
- lg: 12px
- xl: 16px
- 2xl: 24px
- full: 9999px

## CSS Custom Properties

All design tokens are available as CSS custom properties. See `src/styles/variables.scss` for the complete list.

Example usage:
```scss
$custom-color: var(--color-primary);
$padding: var(--padding-md);
$border-radius: var(--radius-lg);
```
